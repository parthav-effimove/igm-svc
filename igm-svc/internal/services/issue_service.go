package services

import (
	"context"
	"encoding/json"
	"fmt"
	"igm-svc/internal/models"
	"igm-svc/internal/repository"
	"log"
	"time"

	//replace with repo
	pb "igm-svc/api/proto/igm/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gorm.io/datatypes"
)

type IssueService struct {
	issueRepo  repository.IssueRepository
	redisRepo  repository.RedisRepository
	OndcClient *OndcClient
	config     *Config
}

type Config struct {
	SubcriberID string
	BAPURI      string
}

func NewIssueService(issueRepo repository.IssueRepository,
	redisRepo repository.RedisRepository,
	ondcClient *OndcClient,
	config *Config,
) *IssueService {
	return &IssueService{
		issueRepo:  issueRepo,
		redisRepo:  redisRepo,
		OndcClient: ondcClient,
		config:     config,
	}
}

func (s *IssueService) CreateIssue(ctx context.Context, req *pb.CreateIssueRequest) (*pb.CreateIssueResponse, error) {
	log.Println("1.reacher CreateIssue [issueservice]")
	err := s.validateCreateRequest(req)
	if err != nil {
		return nil, fmt.Errorf("validation failed :%w", err)
	}
	log.Println("3.reacher after Create][issue_repo]")
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id :%w", err)
	}

	// 3. TODO: Verify order exists and belongs to user
	//orderDetails:= orderClient.VerifyOrder(ctx, req.OrderId, req.UserId)
	// For now, use mock BPP details
	bppID := "preprod.logistics-seller.mp2.in"
	bppURI := "https://preprod.logistics-seller.mp2.in/ondc"

	issue, err := s.buildIssueFromRequest(req, userID, bppID, bppURI)
	if err != nil {
		return nil, fmt.Errorf("failed to build issue:%w", err)
	}

	err = s.issueRepo.Create(ctx, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to save issue :%w", err)
	}
	log.Printf("issue saved to DB :%s", issue.IssueID)
	ondcSend := false
	ondcMessage := ""

	err = s.OndcClient.SendIssue(ctx, issue, "OPEN")
	if err != nil {
		log.Printf("failed to send issue to BPP:%v", err)
		ondcMessage = fmt.Sprintf("Failed to send to BPP: %v", err)
	} else {
		ondcSend = true
		ondcMessage = "Issue sent to BPP successfully"
		log.Printf("[Service] Issue sent to ONDC successfully")
		_ = s.redisRepo.SaveIssueResponse(ctx, issue.TransactionID, map[string]interface{}{
			"action":    "issue",
			"issue_id":  issue.IssueID,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}

	return &pb.CreateIssueResponse{
		IssueId:       issue.IssueID,
		OrderId:       issue.OrderID,
		Status:        issue.Status,
		TransactionId: issue.TransactionID,
		CreatedAt:     issue.CreatedAt.Format(time.RFC3339),
		OndcSent:      ondcSend,
		OndcMessage:   ondcMessage,
	}, nil

}
func (s *IssueService) validateCreateRequest(req *pb.CreateIssueRequest) error {

	if req.UserId == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.OrderId == "" {
		return fmt.Errorf("order_id is required")
	}
	if req.Category == "" {
		return fmt.Errorf("catergory is required")
	}
	if req.SubCategory == "" {
		return fmt.Errorf("sub_category is required")
	}
	if req.IssueType == "" {
		return fmt.Errorf("issue_type is required")
	}
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("atleast one item is required")
	}

	for i, item := range req.Items {
		if item.Id == "" {
			return fmt.Errorf("missing item id at index %d", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("invalid quantity of ite, %s", item.Id)
		}
	}
	validCategoies := []string{"ORDER", "FULFILLMENT", "PAYMENT", "ITEM", "AGENT", "CUSTOMER", "TECHNICAL", "VISIBILITY", "POLICY BREACH", "BUSINESS"}
	if !Contains(validCategoies, req.Category) {
		return fmt.Errorf("invalid category , must be one of %v", validCategoies)
	}
	validIssueTypes := []string{"ISSUE", "GRIEVANCE"}
	if !Contains(validIssueTypes, req.IssueType) {
		return fmt.Errorf("invalid issue type , must be one of %v", validIssueTypes)
	}
	return nil
}
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *IssueService) buildIssueFromRequest(req *pb.CreateIssueRequest,
	userID uuid.UUID,
	bppID, bppURI string,
) (*models.Issue, error) {

	now := time.Now()
	issueID := uuid.New().String()

	imagesJSON, err := json.Marshal(req.ImageUrls)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal images:%w", err)
	}

	orderItems := make([]map[string]interface{}, len(req.Items))
	for i, item := range req.Items {
		orderItems[i] = map[string]interface{}{
			"id":       item.Id,
			"quantity": item.Quantity,
		}
	}

	orderDetailsMap := map[string]interface{}{
		"id":          req.OrderId,
		"provider_id": bppID,
		"items":       orderItems,
	}

	orderDetailsJSON, err := json.Marshal(orderDetailsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marhsal order details :%w", err)
	}

	complaintAction := map[string]interface{}{
		"complaint_action": "OPEN",
		"short_desc":       req.Description,
		"updated_at":       now.Format(time.RFC3339),
		"updated_by": map[string]interface{}{
			"org": map[string]interface{}{
				"name": s.config.SubcriberID,
			},
			"contact": map[string]interface{}{
				"phone": "",
				"email": "", // TODO: Get from user profile
			},
			"person": map[string]interface{}{
				"name": "",
			},
		},
	}

	complaintActionJSON, err := json.Marshal([]interface{}{complaintAction})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal complaint action:%w", err)
	}

	var descURL, descContentType string
	if req.AdditionalDesc != nil {
		descURL = req.AdditionalDesc.Url
		descContentType = req.AdditionalDesc.ContentType
	}

	issue := &models.Issue{
		IssueID:                issueID,
		OrderID:                req.OrderId,
		UserID:                 userID,
		TransactionID:          uuid.New().String(),
		BPPID:                  bppID,
		BPPURI:                 bppURI,
		Category:               req.Category,
		SubCategory:            req.SubCategory,
		IssueType:              req.IssueType,
		Status:                 "OPEN",
		DescriptionShort:       req.Description,
		DescriptionLong:        req.LongDescription,
		DescriptionURL:         descURL,
		DescriptionContentType: descContentType,
		Images:                 datatypes.JSON(imagesJSON),
		OrderDetails:           datatypes.JSON(orderDetailsJSON),
		ComplainantActions:     datatypes.JSON(complaintActionJSON),
		SourceNPID:             s.config.SubcriberID,
		SourceType:             "CONSUMER",
		ExpectedResponseTime:   "PT2H",
		ExpectedResolutionTime: "P1D",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	return issue, nil
}

func (s *IssueService) UpdateIssue(ctx context.Context, req *pb.UpdateIssueRequest) (*pb.UpdateIssueResponse, error) {
	err := ValidateUpdateIssueRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	//TODO veirfy order data

	issue, err := s.issueRepo.GetByIssueID(ctx, req.IssueId)
	if err != nil {
		return nil, fmt.Errorf("issue not found:%w", err)
	}

	issue.Status = req.Status
	issue.IssueType = req.IssueType
	issue.UpdatedAt = time.Now()

	if req.ComplainantActionShortDesc != "" {
		var actions []map[string]interface{}
		if len(issue.ComplainantActions) > 0 {
			_ = json.Unmarshal(issue.ComplainantActions, &actions)
		}
		actions = append(actions, map[string]interface{}{
			"complaint_action": "ESCALATE",
			"short_desc":       req.ComplainantActionShortDesc,
			"updated_at":       issue.UpdatedAt.Format(time.RFC3339),
		})
		actionsJSON, _ := json.Marshal(actions)
		issue.ComplainantActions = datatypes.JSON(actionsJSON)
	}

	err = s.issueRepo.Update(ctx, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue:%w", err)
	}

	ondcSend := false
	ondcMessage := ""
	err = s.OndcClient.SendIssue(ctx, issue, "ESCALATE")
	if err == nil {
		ondcSend = true
		ondcMessage = "issue update sent to BPP"
	}
	return &pb.UpdateIssueResponse{
		IssueId:     issue.IssueID,
		Status:      issue.Status,
		UpdatedAt:   issue.UpdatedAt.Format(time.RFC3339),
		OndcSent:    ondcSend,
		OndcMessage: ondcMessage,
	}, nil
}

func ValidateUpdateIssueRequest(req *pb.UpdateIssueRequest) error {
	if req.UserId == "" {
		return fmt.Errorf("missing required field: user_id")
	}
	if req.IssueId == "" {
		return fmt.Errorf("missing required field: issue_id")
	}
	if req.OrderId == "" {
		return fmt.Errorf("missing required field: order_id")
	}
	if req.Status == "" {
		return fmt.Errorf("missing required field: status")
	}
	if req.IssueType != "" {
		validIssueTypes := []string{"ISSUE", "GRIEVANCE", "DISPUTE"}
		if !Contains(validIssueTypes, req.IssueType) {
			return fmt.Errorf("invalid issue_type. Must be 'ISSUE', 'GRIEVANCE', or 'DISPUTE'")
		}
	}
	return nil
}

func (s *IssueService) CloseIssue(ctx context.Context,req *pb.CloseIssueRequest)(*pb.CloseIssueResponse,error){
	err:=ValidateCloseIssueRequest(req)
	if err!=nil{
		return nil,status.Errorf(codes.InvalidArgument,err.Error())
	}
	//validate order todo

	issue,err:=s.issueRepo.GetByIssueID(ctx,req.IssueId)
	if err!=nil{
		return nil,fmt.Errorf("issue not found :%v",err)
	}
	issue.Status=req.Status
	issue.Rating=req.Rating
	issue.UpdatedAt=time.Now()

	if req.ComplainantActShortDesc!=""{
		var actions []map[string]interface{}
		if len(issue.ComplainantActions)>0{
			_=json.Unmarshal(issue.ComplainantActions,&actions)
		}
		actions=append(actions, map[string]interface{}{
			"complainant_action":"CLOSE",
			"short_desc":req.ComplainantActShortDesc,
			"updated_at":issue.UpdatedAt.Format(time.RFC3339),
		})
		actionsJSON,_:=json.Marshal(actions)
		issue.ComplainantActions=datatypes.JSON(actionsJSON)
	}
	err=s.issueRepo.Update(ctx,issue)
	if err!=nil{
		return nil,fmt.Errorf("failed to update the issue :%w",err)
	}

	ondcSent:=false
	ondcMessage :=""
	err=s.OndcClient.SendIssue(ctx,issue,"CLOSE")
	if err==nil{
		ondcSent=true
		ondcMessage="issue close send to BPP"
	}

	return &pb.CloseIssueResponse{
		IssueId: issue.IssueID,
		Status: issue.Status,
		ClosedAt: issue.UpdatedAt.Format(time.RFC3339),
		OndcSent: ondcSent,
		OndcMessage: ondcMessage,
	},nil

}
func ValidateCloseIssueRequest(req *pb.CloseIssueRequest) error {
	if req.UserId == "" {
		return fmt.Errorf("missing required field: user_id")
	}
	if req.IssueId == "" {
		return fmt.Errorf("missing required field: issue_id")
	}
	if req.OrderId == "" {
		return fmt.Errorf("missing required field: order_id")
	}
	if req.Status == "" {
		return fmt.Errorf("missing required field: status")
	}
	if req.Rating == "" {
		return fmt.Errorf("missing required field: rating")
	}
	validRatings := []string{"THUMBS-UP", "THUMBS-DOWN"}
	if !Contains(validRatings, req.Rating) {
		return fmt.Errorf("invalid rating. Must be 'THUMBS-UP' or 'THUMBS-DOWN'")
	}
	return nil
}

// todo validate functions
func (s *IssueService) ValidateNoActiveIssueExistsWithSameCategory(req *pb.CreateIssueRequest) error {
	return nil

}

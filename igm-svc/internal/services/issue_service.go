package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"igm-svc/internal/models"
	"igm-svc/internal/repository"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type IssueService struct {
	issueRepo repository.IssueRepository
	redisRepo repository.RedisRepository
	config    *Config
}

type Config struct {
	SubcriberID string
	BAPURI      string
}

func NewIssueService(issueRepo repository.IssueRepository,
	redisRepo repository.RedisRepository,
	config *Config,
) *IssueService {
	return &IssueService{
		issueRepo: issueRepo,
		redisRepo: redisRepo,
		config:    config,
	}
}

func (s *IssueService) CreateIssue(ctx context.Context, req *models.IssueCreateRequest, userID uuid.UUID) (*models.Issue, error) {

	issue, err := s.buildIssueFromRequest(req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to build issue:%w", err)
	}

	err = s.issueRepo.Create(ctx, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to save issue :%w", err)
	}
	log.Printf("issue saved to DB :%s", issue.IssueID)

	err = s.sendIssueToBPP(ctx, issue, "OPEN")
	if err!=nil{
		log.Printf("failed to send issue to BPP:%v",err)
	}

	return  issue,nil

}

func (s *IssueService) buildIssueFromRequest(req *models.IssueCreateRequest, userID uuid.UUID) (*models.Issue, error) {
	now := time.Now()
	issueID := uuid.New().String()

	// TODO: Fetch order details from Order Service to get BPPID, BPPURI

	bppID := "demo-bpp-id"
	bppURI := "https://demo-bpp.com"

	imagesJSON, err := json.Marshal(req.ImageURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal images:%w", err)
	}

	orderDetailsMap := map[string]interface{}{
		"id":          req.OrderID,
		"provider_id": bppID,
		"items":       req.Items,
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

	issue := &models.Issue{
		IssueID:                issueID,
		OrderID:                req.OrderID,
		UserID:                 userID,
		TransactionID:          uuid.New().String(),
		BPPID:                  bppID,
		BPPURI:                 bppURI,
		Category:               req.Category,
		SubCategory:            req.SubCategory,
		IssueType:              req.IssueType,
		Status:                 "OPEN",
		DescriptionShort:       req.Description,
		DescriptionLong:        req.LongDesc,
		DescriptionURL:         req.AdditionalDesc.Url,
		DescriptionContentType: req.AdditionalDesc.ContentType,
		Images:                 datatypes.JSON(imagesJSON),
		OrderDetails:           datatypes.JSON(orderDetailsJSON),
		ComplainantActions:     datatypes.JSON(complaintActionJSON),
		SourceNPID:             s.config.SubcriberID,
		SourceType:             "CONSUMER",
		ExpectedResponseTime:   "PT2H",
		ExpectedResolutionTime: "P1D",
		Rating:                 req.Rating,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	return issue, nil
}

func (s *IssueService) sendIssueToBPP(ctx context.Context, issue *models.Issue, operation string) error {
	log.Printf("Sending issue to BPP:%s", issue.BPPURI)

	payload, err := s.buildIssuePayload(issue, operation)

	if err!=nil{
		return fmt.Errorf("failed to build payload:%w",err)
	}
	body,err :=json.Marshal(payload)
	if err!=nil{
		return fmt.Errorf("failed to marhsal payload:%w",err)
	}

	authHeader,err :=s.createAuthHeader(body)
	if err!=nil{
		return fmt.Errorf("failed to create auth header:%w",err)
	}

	url:=issue.BPPURI
	if url[len(url)-1]!='/'{
		url+="/"
	}
	url+="issue"

	httpReq,err :=http.NewRequest("POST",url,bytes.NewBuffer(body))
	if err!=nil{
		return fmt.Errorf("failed to create HTTP req:%w",err)
	}
	httpReq.Header.Set("Content-Type","application/json")
	httpReq.Header.Set("Authorization",authHeader)

	client :=&http.Client{Timeout: 10*time.Second}

	resp,err :=client.Do(httpReq)
	if err!=nil{
		return  fmt.Errorf("http req failed:%w",err)
	}
	defer resp.Body.Close()

	respBody, err :=io.ReadAll(resp.Body)
	if err!=nil{
		return fmt.Errorf("failed to read response body:%w",err)
	}
	log.Printf("BPP response: %s",string(respBody))

	err=s.redisRepo.SaveIssueResponse(ctx,issue.TransactionID,map[string]interface{}{
		"action":"issue",
		"request":payload,
		"response":string(respBody),
	})
	if err!=nil{
		log.Printf("failed to cache issue response %w",err)
	}
	 // TODO: Save to ONDC logs service (via gRPC call)
    // TODO: Submit to transaction logs API

	return nil
}

// buildIssuePayload builds ONDC /issue request payload
func (s *IssueService) buildIssuePayload(issue *models.Issue, operation string) (map[string]interface{}, error) {
	ctx := map[string]interface{}{
		"domain":         "nic2004:60232",
		"country":        "IND",
		"city":           "std:080",
		"action":         "issue",
		"core_version":   "1.2.0",
		"bap_id":         s.config.SubcriberID,
		"bap_uri":        s.config.BAPURI,
		"bpp_id":         issue.BPPID,
		"bpp_uri":        issue.BPPURI,
		"transaction_id": issue.TransactionID,
		"message_id":     uuid.New().String(),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"ttl":            "PT30S",
	}

	//build issue body based on operation

	issueBody, err := s.mapIssueToONDCFormat(issue, operation)
	if err!=nil{
		return nil,err
	}

	return map[string]interface{}{
		"context":ctx,
		"message":map[string]interface{}{
			"issue":issueBody,
		},
	},nil
}

func (s *IssueService) mapIssueToONDCFormat(issue *models.Issue, operation string) (map[string]interface{}, error) {

	var images []string
	if len(issue.Images) > 0 {
		err := json.Unmarshal(issue.Images, &images)
		if err != nil {

			return nil, fmt.Errorf("failed to unmarshal images:%w", err)
		}
	}

	var orderDetails map[string]interface{}
	if len(issue.OrderDetails) > 0 {
		err := json.Unmarshal(issue.OrderDetails, &orderDetails)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal order detail:%w", err)
		}
	}
	var complaintAction map[string]interface{}

	if len(issue.ComplainantActions) > 0 {
		err := json.Unmarshal(issue.ComplainantActions, &complaintAction)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal complaint action:%w", err)
		}
	}

	baseIssue := map[string]interface{}{
		"id":         issue.IssueID,
		"created_at": issue.CreatedAt.Format(time.RFC3339),
		"updated_at": issue.UpdatedAt.Format(time.RFC3339),
	}

	switch operation{
	case "OPEN":
		baseIssue["category"]=issue.Category
		baseIssue["sub-category"]=issue.SubCategory
		baseIssue["complaint_info"]=map[string]interface{}{
			"person":map[string]interface{}{
				"name":issue.UserName,
			},
			"contact":map[string]interface{}{
				"phone":issue.UserPhone,
				"email":issue.UserEmail,
			},
		}
		baseIssue["order_details"]=orderDetails
		baseIssue["description"]=map[string]interface{}{
			"short_desc":issue.DescriptionShort,
			"long_desc":issue.DescriptionLong,
			"additional_desc":map[string]interface{}{
				"url":issue.DescriptionURL,
				"content_type":issue.DescriptionContentType,
			},
			"images":images,
		}
		baseIssue["source"]=map[string]interface{}{
			"network_participant_id":issue.SourceNPID,
			"type":issue.SourceType,
		}
		baseIssue["expected_response_time"]=map[string]interface{}{
			"duration":issue.ExpectedResponseTime,
		}
		baseIssue["expected_resolution_time"]=map[string]interface{}{
			"duration":issue.ExpectedResolutionTime,
		}
		baseIssue["status"]=issue.Status
		baseIssue["issue_type"]=issue.IssueType
		baseIssue["issue_actions"]=map[string]interface{}{
			"complaint_actions":complaintAction,
		}
	case "CLOSE":
		baseIssue["status"]="CLOSED"
		baseIssue["rating"]=issue.Rating
		baseIssue["issue_actions"]=map[string]interface{}{
			"complaint_actions":complaintAction,
		}
	case "ESCALATE":
		baseIssue["status"]=issue.Status
		baseIssue["issue_type"]=issue.IssueType
		baseIssue["issue_actions"]=map[string]interface{}{
			"complaint_actions":complaintAction,
		}
	}
	return baseIssue,nil

}
//TODO implement proper ONDC signature logic
func(s *IssueService)createAuthHeader(body []byte)(string,error){
	return "Signature",nil
}
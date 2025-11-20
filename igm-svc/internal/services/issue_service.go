package services

import (
	"context"
	"encoding/json"
	"fmt"
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



func (s *IssueService) CloseIssue(ctx context.Context, req *pb.CloseIssueRequest) (*pb.CloseIssueResponse, error) {
	err := ValidateCloseIssueRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	//validate order todo

	issue, err := s.issueRepo.GetByIssueID(ctx, req.IssueId)
	if err != nil {
		return nil, fmt.Errorf("issue not found :%v", err)
	}
	issue.Status = req.Status
	issue.Rating = req.Rating
	issue.UpdatedAt = time.Now()

	if req.ComplainantActShortDesc != "" {
		var actions []map[string]interface{}
		if len(issue.ComplainantActions) > 0 {
			_ = json.Unmarshal(issue.ComplainantActions, &actions)
		}
		actions = append(actions, map[string]interface{}{
			"complainant_action": "CLOSE",
			"short_desc":         req.ComplainantActShortDesc,
			"updated_at":         issue.UpdatedAt.Format(time.RFC3339),
		})
		actionsJSON, _ := json.Marshal(actions)
		issue.ComplainantActions = datatypes.JSON(actionsJSON)
	}
	err = s.issueRepo.Update(ctx, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to update the issue :%w", err)
	}

	ondcSent := false
	ondcMessage := ""
	err = s.OndcClient.SendIssue(ctx, issue, "CLOSE")
	if err == nil {
		ondcSent = true
		ondcMessage = "issue close send to BPP"
	}

	return &pb.CloseIssueResponse{
		IssueId:     issue.IssueID,
		Status:      issue.Status,
		ClosedAt:    issue.UpdatedAt.Format(time.RFC3339),
		OndcSent:    ondcSent,
		OndcMessage: ondcMessage,
	}, nil

}




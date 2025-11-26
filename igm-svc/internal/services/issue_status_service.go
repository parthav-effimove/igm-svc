package services

import (
	"context"
	"fmt"
	pb "igm-svc/api/proto/igm/v1"
	"igm-svc/internal/repository"
	"log"

	"github.com/google/uuid"
)

type IssueStatusService struct {
	issueRepo  repository.IssueRepository 
	redisRepo  repository.RedisRepository
	OndcClient *OndcClient
	config     *Config
}

func NewIssueStatusService(
	issueRepo repository.IssueRepository,
	redisRepo repository.RedisRepository,
	ondcClient *OndcClient,
	config *Config,
) *IssueStatusService {
	return &IssueStatusService{
		issueRepo:  issueRepo,
		redisRepo:  redisRepo,
		OndcClient: ondcClient,
		config:     config,
	}
}

// ProcessIssueStatus sends issue_status request to BPP
func (s *IssueStatusService) ProcessIssueStatus(ctx context.Context, req *pb.IssueStatusRequest) error {
	if req == nil {
		return fmt.Errorf("nil request")
	}
	if req.IssueId == "" {
		return fmt.Errorf("missing issue_id")
	}
	if req.UserId == "" {
		return fmt.Errorf("missing user_id")
	}

	
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	
	issue, err := s.issueRepo.GetIssueExistByIssueID(req.IssueId, userID)
	if err != nil {
		return fmt.Errorf("issue not found or access denied: %w", err)
	}

	log.Printf("[IssueStatusService] Sending issue_status for issue: %s to BPP: %s", issue.IssueID, issue.BPPURI)

	// OndcClient.SendIssueStatus builds the context and payload
	
	err = s.OndcClient.SendIssueStatus(ctx, issue)
	if err != nil {
		log.Printf("[IssueStatusService] Failed to send issue_status: %v", err)
		return fmt.Errorf("failed to send issue_status to BPP: %w", err)
	}

	// Publish event to Redis 
	if s.redisRepo != nil {
		event := map[string]interface{}{
			"action":         "issue_status_sent",
			"issue_id":       issue.IssueID,
			"transaction_id": issue.TransactionID,
			"user_id":        req.UserId,
			"bpp_id":         issue.BPPID,
			"bpp_uri":        issue.BPPURI,
		}
		_ = s.redisRepo.SaveIssueResponse(ctx, issue.TransactionID, event)
	}

	log.Printf("[IssueStatusService] issue_status request sent successfully")
	return nil
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"igm-svc/internal/models"
	"igm-svc/internal/repository"
	"log"
	"time"

	pb "github.com/parthav-effimove/ONDC-Protos/protos/ondc/igm/v1"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/datatypes"
)

type IssueStatusService struct {
	issueRepo   repository.IssueRepository
	onIssueRepo repository.OnIssueRepository
	redisRepo   repository.RedisRepository
	OndcClient  *OndcClient
	config      *Config
}

func NewIssueStatusService(
	issueRepo repository.IssueRepository,
	onIssueRepo repository.OnIssueRepository,
	redisRepo repository.RedisRepository,
	ondcClient *OndcClient,
	config *Config,
) *IssueStatusService {
	return &IssueStatusService{
		issueRepo:   issueRepo,
		onIssueRepo: onIssueRepo,
		redisRepo:   redisRepo,
		OndcClient:  ondcClient,
		config:      config,
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

func (s *IssueStatusService) ProcessOnIssueStatus(ctx context.Context, transactionID, messageID string, payload *pb.OnIssuePayload) error {
	if payload == nil {
		return fmt.Errorf("nil payload")
	}

	marshaler := protojson.MarshalOptions{EmitUnpopulated: false}
	raw, err := marshaler.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Save raw callback
	err = s.onIssueRepo.SaveOnIssueCallback(ctx, transactionID, messageID, raw)
	if err != nil {
		log.Printf("warn: SaveOnIssueCallback returned: %v", err)
	}

	if payload.GetIssue() == nil || payload.Issue.GetId() == "" {
		return fmt.Errorf("payload missing issue.id")
	}
	issueID := payload.Issue.Id

	updates := map[string]interface{}{}
	now := time.Now()

	var respondentActionsJSON, resolutionProviderJSON, resolutionJSON []byte

	// Extract Respondent Actions
	ia := payload.Issue.GetIssueActions()
	if ia != nil && len(ia.GetRespondentActions()) > 0 {
		if b, err := marshaler.Marshal(ia); err == nil {
			updates["respondent_actions"] = datatypes.JSON(b)
			respondentActionsJSON = b
		} else {
			log.Printf("warn: failed to marshal respondent actions: %v", err)
		}

		last := ia.RespondentActions[len(ia.RespondentActions)-1]
		if last != nil && last.GetRespondentAction() != "" {
			updates["respondent_status"] = last.GetRespondentAction()
		}
	}

	// Extract Resolution Provider
	rp := payload.Issue.GetResolutionProvider()
	if rp != nil {
		if b, err := marshaler.Marshal(rp); err == nil {
			resolutionProviderJSON = b
			updates["resolution_provider"] = datatypes.JSON(b)
		} else {
			log.Printf("warn: failed to marshal resolution_provider: %v", err)
		}
	}

	// Extract Resolution
	res := payload.Issue.GetResolution()
	if res != nil {
		if b, err := marshaler.Marshal(res); err == nil {
			resolutionJSON = b
			updates["resolution"] = datatypes.JSON(b)
		} else {
			log.Printf("warn: failed to marshal resolution: %v", err)
		}
		if res.GetRefundAmount() != "" {
			updates["refund_amount"] = res.GetRefundAmount()
		}
	}

	// Update Issue Status if present
	if payload.Issue.Status != "" {
		updates["status"] = payload.Issue.Status
	}

	updates["updated_at"] = now

	// Save Response History
	if s.onIssueRepo != nil {
		onIssueStatusResponse := &models.OnIssueStatusResponse{
			IssueID:            issueID,
			TransactionID:      transactionID,
			MessageID:          messageID,
			RespondentActions:  datatypes.JSON(respondentActionsJSON),
			ResolutionProvider: datatypes.JSON(resolutionProviderJSON),
			Resolution:         datatypes.JSON(resolutionJSON),
			CreatedAt:          now,
		}
		err = s.onIssueRepo.SaveOnIssueStatusResponse(ctx, onIssueStatusResponse)
		if err != nil {
			log.Printf("warn: failed to save on_issue_status history: %v", err)
		}
	}

	// Push to Redis
	if s.redisRepo != nil {
		event := map[string]interface{}{
			"action":         "on_issue_status",
			"issue_id":       issueID,
			"transaction_id": transactionID,
			"message_id":     messageID,
			"timestamp":      now.Format(time.RFC3339),
			"status":         payload.Issue.Status,
		}
		if len(respondentActionsJSON) > 0 {
			var ra interface{}
			_ = json.Unmarshal(respondentActionsJSON, &ra)
			event["respondent_actions"] = ra
		}
		if len(resolutionProviderJSON) > 0 {
			var rp interface{}
			_ = json.Unmarshal(resolutionProviderJSON, &rp)
			event["resolution_provider"] = rp
		}
		if len(resolutionJSON) > 0 {
			var rs interface{}
			_ = json.Unmarshal(resolutionJSON, &rs)
			event["resolution"] = rs
		}
		if err := s.redisRepo.SaveIssueResponse(ctx, transactionID, event); err != nil {
			log.Printf("warn: failed to push redis event: %v", err)
		}
	}

	// Update Issue in DB
	err = s.onIssueRepo.UpdateIssueFromOnIssue(ctx, issueID, updates)
	if err != nil {
		return fmt.Errorf("failed to update issue from on_issue_status: %w", err)
	}

	return nil
}

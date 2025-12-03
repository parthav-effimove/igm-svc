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

	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/datatypes"
)

type OnIssueService struct {
	onIssueRepo repository.OnIssueRepository
	redisRepo   repository.RedisRepository
	OndcClient  *OndcClient
	config      *Config
}

func NewOnIssueService(onIssueRepo repository.OnIssueRepository,
	redisRepo repository.RedisRepository,
	ondcClient *OndcClient,
	config *Config) *OnIssueService {
	return &OnIssueService{
		onIssueRepo: onIssueRepo,
		redisRepo:   redisRepo,
		OndcClient:  ondcClient,
		config:      config,
	}
}

func (h *OnIssueService) ProcessOnIssue(ctx context.Context, transactionID, messageID string, payload *pb.OnIssuePayload) error {
	if payload == nil {
		return fmt.Errorf("nil payload")
	}
	marshaler := protojson.MarshalOptions{EmitUnpopulated: false}
	raw, err := marshaler.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	} else {
		err = h.onIssueRepo.SaveOnIssueCallback(ctx, transactionID, messageID, raw)
		if err != nil {

			log.Printf("warn: SaveOndcCallback returned: %v", err)
		}
	}
	//todo ondccallback

	if payload.GetIssue() == nil || payload.Issue.GetId() == "" {
		return fmt.Errorf("payload missing issue.id")
	}
	issueID := payload.Issue.Id

	updates := map[string]interface{}{}

	now := time.Now()

	var respondentActionsJSON, resolutionProviderJSON, resolutionJSON []byte

	ia := payload.Issue.GetIssueActions()
	if ia != nil && len(ia.GetRespondentActions()) > 0 {
		if b, err := marshaler.Marshal(ia); err == nil {
			updates["respondent_actions"] = datatypes.JSON(b)
			respondentActionsJSON = b
		} else {
			log.Printf("warn:failed to marshal respondent action :%v", err)
		}

		last := ia.RespondentActions[len(ia.RespondentActions)-1]
		if last != nil && last.GetRespondentAction() != "" {
			updates["respondent_status"] = last.GetRespondentAction()
		}
	}

	rp := payload.Issue.GetResolutionProvider()
	if rp != nil {
		if b, err := marshaler.Marshal(rp); err == nil {
			resolutionProviderJSON = b
			updates["resolution_provider"] = datatypes.JSON(b)
		} else {
			log.Printf("warn: failed to marshal resolution_provider: %v", err)
		}
	}

	res := payload.Issue.GetResolution()
	if res != nil {
		if b, err := marshaler.Marshal(res); err == nil {
			resolutionJSON = b
			updates["resolution"] = datatypes.JSON(b)
		} else {
			log.Printf("warn: failed to marshal resolution:%v", err)
		}
		if res.GetRefundAmount() != "" {
			updates["refund_amount"] = res.GetRefundAmount()
		}
	}

	updates["updated_at"] = now
	if h.onIssueRepo != nil {
		onIssuestatusResponse := &models.OnIssueStatusResponse{
			IssueID:            issueID,
			TransactionID:      transactionID,
			MessageID:          messageID,
			RespondentActions:  datatypes.JSON(respondentActionsJSON),
			ResolutionProvider: datatypes.JSON(resolutionProviderJSON),
			Resolution:         datatypes.JSON(resolutionJSON),
			CreatedAt:          now,
		}
		err = h.onIssueRepo.SaveOnIssueStatusResponse(ctx, onIssuestatusResponse)
		if err != nil {
			log.Printf("warn: failed to save on_issue history: %v", err)
		}
	}
	if h.redisRepo != nil {
		event := map[string]interface{}{
			"action":         "on_issue_status",
			"issue_id":       issueID,
			"transaction_id": transactionID,
			"message_id":     messageID,
			"timestamp":      now.Format(time.RFC3339),
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
		if err := h.redisRepo.SaveIssueResponse(ctx, transactionID, event); err != nil {
			log.Printf("warn: failed to push redis event: %v", err)
		}

	}

	err = h.onIssueRepo.UpdateIssueFromOnIssue(ctx, issueID, updates)
	if err != nil {
		return fmt.Errorf("failed to update issue from on_issue:%w", err)
	}

	return nil

}

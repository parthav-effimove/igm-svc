package handlers

import (
	"context"
	"fmt"
	pb "igm-svc/api/proto/igm/v1"
	"log"
)

func (h *IssueHandler) HandleIssueStatus(ctx context.Context, req *pb.IssueStatusRequest) (*pb.IssueStatusResponse, error) {

	log.Printf("[Handler] HandleIssueStatus called for issue: %s", req.IssueId)
	err := h.issueStatusService.ProcessIssueStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to process issue_status: %w", err)
	}

	return &pb.IssueStatusResponse{Status: "SUCCESS", Message: "issue_status processed"}, nil
}

func (h *IssueHandler) HandleOnIssueStatus(ctx context.Context, req *pb.OnIssueStatusRequest) (*pb.OnIssueStatusResponse, error) {
	log.Printf("[ONDC] Received on_issue_status callback: issue_id=%s,transaction_id=%s, message_id=%s", req.IssueId, req.TransactionId, req.MessageId)

	if req.Payload == nil {
		return &pb.OnIssueStatusResponse{Status: "ERROR", Message: "empty payload"}, nil
	}

	err := h.issueStatusService.ProcessOnIssueStatus(ctx, req.TransactionId, req.MessageId, req.Payload)
	if err != nil {
		log.Printf("ProcessOnIssueStatus failed: %v", err)
		return &pb.OnIssueStatusResponse{Status: "ERROR", Message: fmt.Sprintf("%v", err)}, nil
	}

	return &pb.OnIssueStatusResponse{Status: "SUCCESS", Message: "callback processed"}, nil
}

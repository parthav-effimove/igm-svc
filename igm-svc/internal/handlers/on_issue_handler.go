package handlers

import (
	"context"
	pb "igm-svc/api/proto/igm/v1"
	"log"
)

func (h *IssueHandler) HandleOnIssue(ctx context.Context, req *pb.OnIssueRequest) (*pb.OnIssueResponse, error) {
	log.Printf("[ONDC] Received in_issue callback: transaction_id=%s,message_id:=%s", req.TransactionId, req.MessageId)

	return nil, nil
}
func (h *IssueHandler) HandleOnIssueStatus(ctx context.Context, req *pb.OnIssueStatusRequest) (*pb.OnIssueStatusResponse, error) {
	//todo
	return nil, nil
}

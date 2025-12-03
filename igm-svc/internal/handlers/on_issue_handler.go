package handlers

import (
	"context"
	"fmt"
	"log"

	pb "github.com/parthav-effimove/ONDC-Protos/protos/ondc/igm/v1"
)

func (h *IssueHandler) HandleOnIssue(ctx context.Context, req *pb.OnIssueRequest) (*pb.OnIssueResponse, error) {

	log.Printf("[ONDC] Received in_issue callback: transaction_id=%s,message_id:=%s", req.TransactionId, req.MessageId)

	if req.Payload == nil {
		return &pb.OnIssueResponse{Status: "ERROR", Message: "empty payload"}, nil
	}

	err := h.onIssueService.ProcessOnIssue(ctx, req.TransactionId, req.MessageId, req.Payload)
	if err != nil {
		log.Printf("ProcessOnIssue failed %v", err)
		return &pb.OnIssueResponse{Status: "ERROR", Message: fmt.Sprintf("%v", err)}, nil
	}
	return &pb.OnIssueResponse{Status: "SUCCESS", Message: "callback processed"}, nil

}

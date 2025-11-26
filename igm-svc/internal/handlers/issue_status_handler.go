package handlers

import (
	"context"
	"fmt"
	pb "igm-svc/api/proto/igm/v1"
	"log"
)

func (h *IssueHandler) HandleIssueStatus(ctx context.Context, req *pb.IssueStatusRequest) (*pb.IssueStatusResponse, error) {
	
	log.Printf("[Handler] HandleIssueStatus called for issue: %s", req.IssueId)
	err:=h.issueStatusService.ProcessIssueStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to process issue_status: %w", err)
	}

	return &pb.IssueStatusResponse{Status: "SUCCESS", Message: "issue_status processed"}, nil
}

func (h *IssueHandler) HandleOnIssueStatus(ctx context.Context, req *pb.OnIssueStatusRequest) (*pb.OnIssueStatusResponse, error) {
	//todo
	return nil, nil
}

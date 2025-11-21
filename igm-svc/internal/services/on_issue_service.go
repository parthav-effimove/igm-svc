package services

import (
	"context"
	"encoding/json"
	pb "igm-svc/api/proto/igm/v1"
	"log"
)

func (h *IssueService) HandleOnIssue(ctx context.Context, req *pb.OnIssueRequest) (*pb.CloseIssueResponse, error) {
	// TODO: Parse req.Payload
	var payloadMap map[string]interface{}
	err:=json.Unmarshal([]byte(req.Payload),&payloadMap)
	if err!=nil{
		log.Printf("Failed to parse payload :%v",err)
		return &pb.OnIssueResponse{
			Status: :"ERROR",
			Message:
		}
	}
	err := h.issueRepo.SaveOnIssueCallback()
}

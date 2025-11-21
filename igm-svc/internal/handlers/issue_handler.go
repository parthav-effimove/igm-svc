package handlers

import (
	"context"
	pb "igm-svc/api/proto/igm/v1"
	"igm-svc/internal/services"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IssueHandler struct {
	pb.UnimplementedIssueServiceServer
	issueService *services.IssueService
}

func NewIssueHandler(issueService *services.IssueService) *IssueHandler {
	return &IssueHandler{
		issueService: issueService,
	}
}

func (h *IssueHandler) CreateIssue(ctx context.Context, req *pb.CreateIssueRequest) (*pb.CreateIssueResponse, error) {
	log.Printf("[Handler] Create Issue called for user:%s, order:%s", req.UserId, req.OrderId)

	resp, err := h.issueService.CreateIssue(ctx, req)
	if err != nil {
		log.Printf("[Handler] CreateIssue Failed :%v", err)
		return nil, err
	}

	// if isValidationError(err){
	// return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)}
	// }
	log.Printf("[Handler] Create issue sucecces issue_id:%s", resp.IssueId)
	return resp, nil
}

func (h *IssueHandler) UpdateIssue(ctx context.Context, req *pb.UpdateIssueRequest) (*pb.UpdateIssueResponse, error) {
	log.Printf("[Handler] UpdateIssue  called for user:%s ,order:%s", req.UserId, req.OrderId)
	resp, err := h.issueService.UpdateIssue(ctx, req)
	if err != nil {
		log.Printf("[handler] Update Issue failed:%v", err)
		return nil, status.Errorf(codes.Internal, "failed to update issue :%v", err)
	}
	return resp, nil
}

func (h *IssueHandler) CloseIssue(ctx context.Context, req *pb.CloseIssueRequest) (*pb.CloseIssueResponse, error) {
	log.Printf("[Handler] CloseIssue called for user:%s , order:%s ", req.UserId, req.OrderId)
	resp, err := h.issueService.CloseIssue(ctx, req)
	if err != nil {
		log.Printf("[handler] Close Issue failed :%v", err)
		return nil, status.Errorf(codes.Internal, "failed to close issue :%v", err)
	}

	return resp, nil
}

func (h *IssueHandler) GetIssue(ctx context.Context, req *pb.GetIssueRequest) (*pb.GetIssueResponse, error) {
	log.Printf("[Handler] GetIssue called for user:%s, order:%s", req.UserId, req.IssueId)
	resp, err := h.issueService.GetIssue(ctx, req)
	if err != nil {
		log.Printf("[handler] Get Issues failed :%v", err)
		return nil, status.Errorf(codes.Internal, "failed to Get Issues issue :%v", err)
	}
	return resp, nil
}

func (h *IssueHandler) ListIssues(ctc context.Context, req *pb.ListIssueRequest) (*pb.ListIssueResponse, error) {
	log.Printf("[Handler] ListIssue called by user:%s", req.UserId)
	resp, err := h.issueService.GetIssuesByUser(ctc, req)
	if err != nil {
		log.Printf("[Handler ListIssues failed:%v]", err)
		return nil, status.Errorf(codes.Internal, "failed to List Issue for user :%v", err)
	}
	return resp, nil
}

func (h *IssueHandler) ListIssueByOrder(ctx context.Context, req *pb.ListIssueByOrderRequest) (*pb.ListIssueResponse, error) {
	log.Printf("[Handler] ListIssueByOrder called by user :%s for order:%s",req.UserId,req.OrderId)
	resp,err:=h.issueService.GetIssueByOrder(ctx,req)
	if err != nil {
		log.Printf("[handler] Get Issues by Order failed :%v", err)
		return nil, status.Errorf(codes.Internal, "failed to Get Issues for order :%v", err)
	}
	return resp, nil
}
func (h *IssueHandler) HandleOnIssue(ctx context.Context, req *pb.OnIssueRequest) (*pb.OnIssueResponse, error) {
		log.Printf("[ONDC] Received in_issue callback: transaction_id=%s,message_id:=%s",req.TransactionId,req.MessageId)

	return nil, nil
}
func (h *IssueHandler) HandleOnIssueStatus(ctx context.Context, req *pb.OnIssueStatusRequest) (*pb.OnIssueStatusResponse, error) {
	//todo
	return nil, nil
}

// func isValidationError(err error) bool {
//     if err == nil {
//         return false
//     }
//     errStr := err.Error()
//     return contains(errStr, "validation") ||
//            contains(errStr, "required") ||
//            contains(errStr, "invalid")
// }

// func contains(str, substr string) bool {
//     return len(str) >= len(substr) && (str == substr ||
//            len(str) > len(substr) && (str[:len(substr)] == substr ||
//            str[len(str)-len(substr):] == substr ||
//            containsHelper(str, substr)))
// }

// func containsHelper(str, substr string) bool {
//     for i := 0; i <= len(str)-len(substr); i++ {
//         if str[i:i+len(substr)] == substr {
//             return true
//         }
//     }
//     return false
// }

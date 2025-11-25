package services

import (
	"encoding/json"
	"fmt"
	pb "igm-svc/api/proto/igm/v1"
	"igm-svc/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func (s *IssueService) validateCreateRequest(req *pb.CreateIssueRequest) error {

	if req.UserId == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.OrderId == "" {
		return fmt.Errorf("order_id is required")
	}
	if req.Category == "" {
		return fmt.Errorf("catergory is required")
	}
	if req.SubCategory == "" {
		return fmt.Errorf("sub_category is required")
	}
	if req.IssueType == "" {
		return fmt.Errorf("issue_type is required")
	}
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("atleast one item is required")
	}

	for i, item := range req.Items {
		if item.Id == "" {
			return fmt.Errorf("missing item id at index %d", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("invalid quantity of ite, %s", item.Id)
		}
	}
	validCategoies := []string{"ORDER", "FULFILLMENT", "PAYMENT", "ITEM", "AGENT", "CUSTOMER", "TECHNICAL", "VISIBILITY", "POLICY BREACH", "BUSINESS"}
	if !Contains(validCategoies, req.Category) {
		return fmt.Errorf("invalid category , must be one of %v", validCategoies)
	}
	validIssueTypes := []string{"ISSUE", "GRIEVANCE"}
	if !Contains(validIssueTypes, req.IssueType) {
		return fmt.Errorf("invalid issue type , must be one of %v", validIssueTypes)
	}
	return nil
}

func (s *IssueService) buildIssueFromRequest(req *pb.CreateIssueRequest,
	userID uuid.UUID,
	bppID, bppURI string,
) (*models.Issue, error) {

	now := time.Now()
	issueID := uuid.New().String()

	imagesJSON, err := json.Marshal(req.ImageUrls)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal images:%w", err)
	}

	orderItems := make([]map[string]interface{}, len(req.Items))
	for i, item := range req.Items {
		orderItems[i] = map[string]interface{}{
			"id":       item.Id,
			"quantity": item.Quantity,
		}
	}

	orderDetailsMap := map[string]interface{}{
		"id":          req.OrderId,
		"provider_id": bppID,
		"items":       orderItems,
	}

	orderDetailsJSON, err := json.Marshal(orderDetailsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marhsal order details :%w", err)
	}

	complaintAction := map[string]interface{}{
		"complainant_action": "OPEN",
		"short_desc":         req.Description,
		"updated_at":         now.Format(time.RFC3339),
		"updated_by": map[string]interface{}{
			"org": map[string]interface{}{
				"name": s.config.SubcriberID,
			},
			"contact": map[string]interface{}{
				"phone": "",
				"email": "", // TODO: Get from user profile
			},
			"person": map[string]interface{}{
				"name": "",
			},
		},
	}

	complaintActionJSON, err := json.Marshal([]interface{}{complaintAction})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal complaint action:%w", err)
	}

	var descURL, descContentType string
	if req.AdditionalDesc != nil {
		descURL = req.AdditionalDesc.Url
		descContentType = req.AdditionalDesc.ContentType
	}

	issue := &models.Issue{
		IssueID:                issueID,
		OrderID:                req.OrderId,
		UserID:                 userID,
		TransactionID:          uuid.New().String(),
		BPPID:                  bppID,
		BPPURI:                 bppURI,
		Category:               req.Category,
		SubCategory:            req.SubCategory,
		IssueType:              req.IssueType,
		Status:                 "OPEN",
		DescriptionShort:       req.Description,
		DescriptionLong:        req.LongDescription,
		DescriptionURL:         descURL,
		DescriptionContentType: descContentType,
		Images:                 datatypes.JSON(imagesJSON),
		OrderDetails:           datatypes.JSON(orderDetailsJSON),
		ComplainantActions:     datatypes.JSON(complaintActionJSON),
		SourceNPID:             s.config.SubcriberID,
		SourceType:             "CONSUMER",
		ExpectedResponseTime:   "PT2H",
		ExpectedResolutionTime: "P1D",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	return issue, nil
}

func ValidateUpdateIssueRequest(req *pb.UpdateIssueRequest) error {
	if req.UserId == "" {
		return fmt.Errorf("missing required field: user_id")
	}
	if req.IssueId == "" {
		return fmt.Errorf("missing required field: issue_id")
	}
	if req.OrderId == "" {
		return fmt.Errorf("missing required field: order_id")
	}
	if req.Status == "" {
		return fmt.Errorf("missing required field: status")
	}
	if req.IssueType != "" {
		validIssueTypes := []string{"ISSUE", "GRIEVANCE", "DISPUTE"}
		if !Contains(validIssueTypes, req.IssueType) {
			return fmt.Errorf("invalid issue_type. Must be 'ISSUE', 'GRIEVANCE', or 'DISPUTE'")
		}
	}
	return nil
}

func ValidateCloseIssueRequest(req *pb.CloseIssueRequest) error {
	if req.UserId == "" {
		return fmt.Errorf("missing required field: user_id")
	}
	if req.IssueId == "" {
		return fmt.Errorf("missing required field: issue_id")
	}
	if req.OrderId == "" {
		return fmt.Errorf("missing required field: order_id")
	}
	if req.Status == "" {
		return fmt.Errorf("missing required field: status")
	}
	if req.Rating == "" {
		return fmt.Errorf("missing required field: rating")
	}
	validRatings := []string{"THUMBS-UP", "THUMBS-DOWN"}
	if !Contains(validRatings, req.Rating) {
		return fmt.Errorf("invalid rating. Must be 'THUMBS-UP' or 'THUMBS-DOWN'")
	}
	return nil
}

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
	
// todo validate functions
func (s *IssueService) ValidateNoActiveIssueExistsWithSameCategory(req *pb.CreateIssueRequest) error {
	return nil

}
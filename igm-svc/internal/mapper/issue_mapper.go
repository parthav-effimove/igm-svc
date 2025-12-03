package mapper

import (
	"encoding/json"
	"igm-svc/internal/models"
	"time"

	pb "github.com/parthav-effimove/ONDC-Protos/protos/ondc/igm/v1"
)

func ToProtoIssue(m *models.Issue) *pb.Issue {
	if m == nil {
		return nil
	}

	proto := &pb.Issue{
		IssueId:          m.IssueID,
		OrderId:          m.OrderID,
		UserId:           m.UserID.String(),
		TransactionId:    m.TransactionID,
		Category:         m.Category,
		SubCategory:      m.SubCategory,
		IssueType:        m.IssueType,
		Status:           m.Status,
		DescriptionShort: m.DescriptionShort,
		DescriptionLong:  m.DescriptionLong,
		ImageUrls:        []string{},
		BppId:            m.BPPID,
		BppUri:           m.BPPURI,
		CreatedAt:        m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        m.UpdatedAt.Format(time.RFC3339),
	}
	if len(m.Images) > 0 {
		var imgs []string
		if err := json.Unmarshal(m.Images, &imgs); err == nil {
			proto.ImageUrls = imgs
		}
	}
	return proto
}

func ToProtoIssues(ms []*models.Issue) []*pb.Issue {
	if ms == nil {
		return nil
	}

	out := make([]*pb.Issue, len(ms))

	for _, m := range ms {
		out = append(out, ToProtoIssue(m))
	}
	return out
}

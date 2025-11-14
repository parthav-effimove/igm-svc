package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Issue struct {
	ID                     uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	IssueID                string         `gorm:"uniqueIndex;not null" json:"issue_id"` // ONDC IssueBody.ID
	OrderID                string         `gorm:"not null;index" json:"order_id"`
	UserID                 uuid.UUID      `gorm:"not null;index" json:"user_id"`
	BPPID                  string         `json:"bpp_id"`
	BPPURI                 string         `json:"bpp_uri"`
	UserName               string         `json:"user_name"`
	UserPhone              string         `json:"user_phone"`
	UserEmail              string         `json:"user_email"`
	TransactionID          string         `json:"transation_id"`
	DescriptionShort       string         `json:"description_short"`
	DescriptionLong        string         `json:"description_long"`
	DescriptionURL         string         `json:"description_url"`
	DescriptionContentType string         `json:"description_content_type"`
	Images                 datatypes.JSON `json:"images"` // JSON array of image URLs or objects
	Category               string         `json:"category"`
	SubCategory            string         `json:"sub_category"`
	IssueType              string         `json:"issue_type"`
	SourceNPID             string         `json:"source_npid"`
	SourceType             string         `json:"source_type"`
	ExpectedResponseTime   string         `json:"expected_response_time"`
	ExpectedResolutionTime string         `json:"expected_resolution_time"`
	Status                 string         `json:"status"`
	RespondentStatus       string         `json:"respondent_status"`
	Rating                 string         `json:"rating"`
	OrderDetails           datatypes.JSON `json:"order_details"`       // JSON object: {id, provider_id, state, items, fulfillments}
	ComplainantActions     datatypes.JSON `json:"complainant_actions"` // JSON array
	RespondentActions      datatypes.JSON `json:"respondent_actions"`
	ResolutionProvider     datatypes.JSON `json:"resolution_provider"`
	Resolution             datatypes.JSON `json:"resolution"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `json:"deleted_at"`
}

type IssueCreateRequest struct {
	OrderID        string   `json:"order_id"`
	Category       string   `json:"category"`
	SubCategory    string   `json:"sub_category"`
	IssueType      string   `json:"issue_type"`
	Description    string   `json:"description"`
	LongDesc       string   `json:"long_desc"`
	ImageURLs      []string `json:"image_urls"`
	Rating         string   `json:"rating"`
	AdditionalDesc struct {
		Url         string `json:"url"`
		ContentType string `json:"content_type"`
	} `json:"additional_desc"`
	Items []IssueItem `json:"items"`
}

type IssueItem struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}

type IssueUpdateRequest struct {
	IssueID                    string `json:"issue_id"`
	OrderID                    string `json:"order_id"`
	IssueType                  string `json:"issue_type"`
	Status                     string `json:"status"`
	Rating                     string `json:"rating"`
	ComplainantActionShortDesc string `json:"complainant_action_short_desc"`
}

type OnIssueStatusResponse struct {
	ID                 uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	IssueID            string         `gorm:"uniqueIndex;not null" json:"issue_id"`
	RespondentActions  datatypes.JSON `json:"respondent_actions" gorm:"type:jsonb"`
	ResolutionProvider datatypes.JSON `json:"resolution_provider" gorm:"type:jsonb"`
	Resolution         datatypes.JSON `json:"resolution" gorm:"type:jsonb"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

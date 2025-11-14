package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type Issue struct {
    ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

    // Core identifiers
    IssueID       string    `gorm:"uniqueIndex;not null" json:"issue_id"`
    OrderID       string    `gorm:"not null;index" json:"order_id"`
    UserID        uuid.UUID `gorm:"not null;index" json:"user_id"`
    TransactionID string    `gorm:"index" json:"transaction_id"`

    // Network participants
    BPPID  string `json:"bpp_id"`
    BPPURI string `json:"bpp_uri"`

    // User/Complainant info
    UserName  string `json:"user_name"`
    UserPhone string `json:"user_phone"`
    UserEmail string `json:"user_email"`

    // Issue classification
    Category         string `json:"category"`
    SubCategory      string `json:"sub_category"`
    IssueType        string `json:"issue_type"`
    Status           string `json:"status"`
    RespondentStatus string `json:"respondent_status"`

    // Description
    DescriptionShort       string         `json:"description_short"`
    DescriptionLong        string         `json:"description_long"`
    DescriptionURL         string         `json:"description_url"`
    DescriptionContentType string         `json:"description_content_type"`
    Images                 datatypes.JSON `json:"images" gorm:"type:jsonb"` // JSON array

    // Source
    SourceNPID string `json:"source_npid"`
    SourceType string `json:"source_type"`

    // Expected timelines
    ExpectedResponseTime   string `json:"expected_response_time"`
    ExpectedResolutionTime string `json:"expected_resolution_time"`

    // Rating
    Rating string `json:"rating"`

    // JSONB fields
    OrderDetails       datatypes.JSON `json:"order_details" gorm:"type:jsonb"`
    ComplainantActions datatypes.JSON `json:"complainant_actions" gorm:"type:jsonb"`
    RespondentActions  datatypes.JSON `json:"respondent_actions" gorm:"type:jsonb"`
    ResolutionProvider datatypes.JSON `json:"resolution_provider" gorm:"type:jsonb"`
    Resolution         datatypes.JSON `json:"resolution" gorm:"type:jsonb"`
}

func (Issue) TableName() string {
    return "issues"
}

// OnIssueStatusResponse stores BPP responses for on_issue_status
type OnIssueStatusResponse struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    UpdatedAt time.Time `json:"updated_at"`

    IssueID            string         `gorm:"uniqueIndex;not null" json:"issue_id"`
    RespondentActions  datatypes.JSON `json:"respondent_actions" gorm:"type:jsonb"`
    ResolutionProvider datatypes.JSON `json:"resolution_provider" gorm:"type:jsonb"`
    Resolution         datatypes.JSON `json:"resolution" gorm:"type:jsonb"`
}

func (OnIssueStatusResponse) TableName() string {
    return "on_issue_status_responses"
}

// Request DTOs (for API validation)
type IssueCreateRequest struct {
    OrderID     string   `json:"order_id" binding:"required"`
    Category    string   `json:"category" binding:"required"`
    SubCategory string   `json:"sub_category" binding:"required"`
    IssueType   string   `json:"issue_type" binding:"required"`
    Description string   `json:"description" binding:"required"`
    LongDesc    string   `json:"long_desc"`
    ImageURLs   []string `json:"image_urls"`
    Rating      string   `json:"rating"`
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
    IssueID                    string `json:"issue_id" binding:"required"`
    OrderID                    string `json:"order_id" binding:"required"`
    IssueType                  string `json:"issue_type"`
    Status                     string `json:"status"`
    Rating                     string `json:"rating"`
    ComplainantActionShortDesc string `json:"complainant_action_short_desc"`
}
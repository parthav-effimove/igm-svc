package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type Issue struct {
    ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
    IssueID       string         `gorm:"uniqueIndex;not null" json:"issue_id"`
    OrderID       string         `gorm:"not null;index" json:"order_id"`
    UserID        uuid.UUID      `gorm:"not null;index;type:uuid" json:"user_id"`
    TransactionID string         `gorm:"index" json:"transaction_id"`
    
    // Network participants - Fix column name mapping
    BPPID  string `gorm:"column:bpp_id;index" json:"bpp_id"`
    BPPURI string `gorm:"column:bpp_uri" json:"bpp_uri"`  
    
    // User/Complainant info
    UserName  string `gorm:"column:user_name" json:"user_name"`
    UserPhone string `gorm:"column:user_phone" json:"user_phone"`
    UserEmail string `gorm:"column:user_email" json:"user_email"`
    
    // Issue classification
    Category         string `gorm:"index" json:"category"`
    SubCategory      string `gorm:"column:sub_category" json:"sub_category"`
    IssueType        string `gorm:"column:issue_type" json:"issue_type"`
    Status           string `gorm:"index" json:"status"`
    RespondentStatus string `gorm:"column:respondent_status" json:"respondent_status"`
    
    // Description
    DescriptionShort       string         `gorm:"column:description_short" json:"description_short"`
    DescriptionLong        string         `gorm:"column:description_long" json:"description_long"`
    DescriptionURL         string         `gorm:"column:description_url" json:"description_url"`
    DescriptionContentType string         `gorm:"column:description_content_type" json:"description_content_type"`
    Images                 datatypes.JSON `gorm:"type:jsonb" json:"images"`
    
    // Source - Fix column name mapping
    SourceNPID string `gorm:"column:source_npid" json:"source_npid"`  
    SourceType string `gorm:"column:source_type" json:"source_type"`
    
    // Expected timelines
    ExpectedResponseTime   string `gorm:"column:expected_response_time" json:"expected_response_time"`
    ExpectedResolutionTime string `gorm:"column:expected_resolution_time" json:"expected_resolution_time"`
    
    // Rating
    Rating string `json:"rating"`
    
    // JSONB fields
    OrderDetails       datatypes.JSON `gorm:"type:jsonb;column:order_details" json:"order_details"`
    ComplainantActions datatypes.JSON `gorm:"type:jsonb;column:complainant_actions" json:"complainant_actions"`
    RespondentActions  datatypes.JSON `gorm:"type:jsonb;column:respondent_actions" json:"respondent_actions"`
    ResolutionProvider datatypes.JSON `gorm:"type:jsonb;column:resolution_provider" json:"resolution_provider"`
    Resolution         datatypes.JSON `gorm:"type:jsonb" json:"resolution"`
    
    // Timestamps
    CreatedAt time.Time      `gorm:"not null;default:now()" json:"created_at"`
    UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName overrides the table name
func (Issue) TableName() string {
    return "issues"
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
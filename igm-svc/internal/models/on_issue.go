package models

import (
    "time"

    "gorm.io/datatypes"
)

type OnIssueStatusResponse struct {
    ID                 uint           `gorm:"primaryKey;autoIncrement" json:"id"`
    IssueID            string         `gorm:"index;not null" json:"issue_id"`
    TransactionID      string         `json:"transaction_id"`
    MessageID          string         `json:"message_id"`
    RespondentActions  datatypes.JSON `json:"respondent_actions" gorm:"type:jsonb"`
    ResolutionProvider datatypes.JSON `json:"resolution_provider" gorm:"type:jsonb"`
    Resolution         datatypes.JSON `json:"resolution" gorm:"type:jsonb"`
    CreatedAt          time.Time      `json:"created_at"`
}

type OndcCallback struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	TransactionID string         `json:"transaction_id"`
	MessageID     string         `json:"message_id"`
	Payload       datatypes.JSON `json:"payload" gorm:"type:jsonb"`
	CreatedAt     time.Time      `json:"created_at"`
}
package repository

import (
	"context"
	"fmt"
	"igm-svc/internal/models"
	"log"
	"time"

	"gorm.io/gorm"
)

type OnIssueRepository interface {
	SaveOnIssueCallback(ctx context.Context,transactionID, messageID string, payload []byte) error
	UpdateIssueFromOnIssue(ctx context.Context, issueID string, updates map[string]interface{}) error
}

type onIssueRepository struct{
	db *gorm.DB
}

func NewOnIssueRepository(db *gorm.DB)OnIssueRepository{
	return &onIssueRepository{db: db}
}

func (r *onIssueRepository) SaveOnIssueCallback(ctx context.Context,transactionID, messageID string, payload []byte) error {
	//todo - integrate with dedicated logging service.
	log.Printf("Saving on_issue callback: transaction_id=%s, message_id=%s, payload=%s", transactionID, messageID, payload)
	return nil
}

func (r *onIssueRepository) UpdateIssueFromOnIssue(ctx context.Context, issueID string, updates map[string]interface{}) error{
	if issueID ==""{
		return fmt.Errorf("empty issue id")
	}
	if len(updates)==0{
		return nil
	}

	if _, ok :=updates["updated_at"]; !ok{
		updates["updated_at"]=time.Now()
	}

	tx:=r.db.WithContext(ctx).Begin()

	if tx.Error !=nil{
		return fmt.Errorf("failed to begin transaction :%w",tx.Error)
	}

	err:=tx.Model(&models.Issue{}).Where("issue_id = ?",issueID).Updates(updates).Error
	if err!=nil{
		_=tx.Rollback()
		return fmt.Errorf("failed to update issue:%w",err)
	}
	err=tx.Commit().Error
	if err!=nil{
		return fmt.Errorf("failed to commit issue update:%w",err)
	}
	return nil
}


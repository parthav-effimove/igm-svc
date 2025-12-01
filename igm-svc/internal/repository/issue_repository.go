package repository

import (
	"context"
	"fmt"
	"igm-svc/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IssueRepository interface {
	Create(ctx context.Context, issue *models.Issue) error
	GetByID(ctx context.Context, id uint) (*models.Issue, error)
	GetByIssueID(ctx context.Context, issueID string) (*models.Issue, error)
	GetByOrderID(ctx context.Context, orderID string) ([]*models.Issue, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Issue,int,error)
	GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*models.Issue, error)
	Update(ctx context.Context, issue *models.Issue) error
	GetIssueExistByIssueID(issueID string, userID uuid.UUID)(*models.Issue,error)
	HasActiveIssueWithSameCategory(userID uuid.UUID, category string, orderID string) (bool, error)
}

type issueRepository struct {
	db *gorm.DB
}

func NewIssueRepository(db *gorm.DB) IssueRepository {
	return &issueRepository{db: db}
}

func (r *issueRepository) Create(ctx context.Context, issue *models.Issue) error {

	if issue == nil {
		return fmt.Errorf("issue cannot be nil")
	}
	return r.db.WithContext(ctx).Create(issue).Error
}

func (r *issueRepository) GetByID(ctx context.Context, id uint) (*models.Issue, error) {
	var issue models.Issue
	err := r.db.WithContext(ctx).First(&issue, id).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

func (r *issueRepository) GetByIssueID(ctx context.Context, issueID string) (*models.Issue, error) {
	var issue models.Issue
	err := r.db.WithContext(ctx).Where("issue_id= ?", issueID).First(&issue).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

func (r *issueRepository) GetByOrderID(ctx context.Context, orderID string) ([]*models.Issue, error) {
	var issues []*models.Issue
	err := r.db.WithContext(ctx).
		Where("order_id= ?", orderID).
		Order("created_at DESC").
		Find(&issues).Error
	return issues, err

}

func (r *issueRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Issue,int, error) {
	var issues []*models.Issue
	var total int64
	err:=r.db.Model(&models.Issue{}).Where("user_id=?",userID).Count(&total).Error
	if err!=nil{
		return nil,0,err
	}
	err = r.db.WithContext(ctx).
		Where("user_id= ?", userID).
		Limit(limit).Offset(offset).
		Find(&issues).Error
	return issues,int(total), err
}

func (r *issueRepository) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*models.Issue, error) {
	var issues []*models.Issue
	err := r.db.WithContext(ctx).
		Where("transaction_id= ?", transactionID).
		Order("created_at DESC").
		Find(&issues).Error
	return issues, err
}

func (r *issueRepository) Update(ctx context.Context, issue *models.Issue) error {
	return r.db.WithContext(ctx).Save(issue).Error
}
func (r *issueRepository)GetIssueExistByIssueID(issueID string, userID uuid.UUID)(*models.Issue,error){
		var issue models.Issue
		err:=r.db.Where("issue_id=? AND user_id=?",issueID,userID).First(&issue).Error
		 if err != nil {
        return nil, fmt.Errorf("no issue found with this issue_id")
    	}
    	return &issue, nil
	}

func (r *issueRepository)HasActiveIssueWithSameCategory(userID uuid.UUID, category string, orderID string) (bool, error){
	var issue models.Issue
	err:=r.db.Where("user_id=? AND category=? AND order_id=? AND status='OPEN'",userID,category,orderID).First(&issue).Error
	if err!=nil{
		return false,err
	}
	return true,nil
}



	
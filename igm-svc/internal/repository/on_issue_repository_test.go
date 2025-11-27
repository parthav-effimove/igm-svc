package repository

import (
	"context"
	"igm-svc/internal/models"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setUpTestDB(t *testing.T) *gorm.DB {
	_ = godotenv.Load("../../.env")
	dsn := os.Getenv("DATABASE_URL")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "failed to connect to db")

	return db
}

func TestOnIssueCallback(t *testing.T) {
	db := setUpTestDB(t)
	repo := NewOnIssueRepository(db)
	ctx := context.Background()

	payload := []byte(`{"foo": "baar"}`)
	err := repo.SaveOnIssueCallback(ctx, "tx-1", "msg-1", payload)

	require.NoError(t, err)

	var found models.OndcCallback
	err = db.Where("transaction_id= ?", "tx-1").First(&found).Error
	require.NoError(t, err)
	require.Equal(t, "msg-1", found.MessageID)
	require.Equal(t, payload, []byte(found.Payload))
}

func TestSaveOnIssueStatusResponse(t *testing.T) {
	db := setUpTestDB(t)
	repo := NewOnIssueRepository(db)
	ctx := context.Background()

	row := &models.OnIssueStatusResponse{
		IssueID:           "issue-1",
		TransactionID:     "tx-1",
		MessageID:         "msg-1",
		RespondentActions: datatypes.JSON([]byte(`{"action":"refund"}`)),
		ResolutionProvider: datatypes.JSON([]byte(`{"provider":"support-team"}`)),
		Resolution: datatypes.JSON([]byte(`{"status":"resolved","details":"Refund processed"}`)),
		CreatedAt: time.Now(),
	}

	err:=repo.SaveOnIssueStatusResponse(ctx,row)
	require.NoError(t,err)

	var found models.OnIssueStatusResponse
	err=db.Where("issue_id = ?","issue-1").First(&found).Error
	require.NoError(t,err)
	require.Equal(t,"msg-1",found.MessageID)
	require.Equal(t,"tx-1",found.TransactionID)
}

func TestUpdateIssueFromOnIssue(t *testing.T){
	db :=setUpTestDB(t)
	repo:=NewOnIssueRepository(db)
	ctx:=context.Background()

	issue:=&models.Issue{IssueID: "issue-2",Status: "OPEN"}
	err:=db.Create(issue).Error
	require.NoError(t,err)

	updates:=map[string]interface{}{"status":"CLOSED"}
	err=repo.UpdateIssueFromOnIssue(ctx,"issue-2",updates)
	require.NoError(t,err)

	var found models.Issue
	err=db.Where("issue_id=?","issue-2").First(&found).Error
	require.NoError(t,err)
	require.Equal(t,"CLOSED",found.Status)
}

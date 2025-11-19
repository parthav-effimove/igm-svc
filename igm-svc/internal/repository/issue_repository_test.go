package repository

import (
	"context"
	"encoding/json"
	"igm-svc/internal/models"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/joho/godotenv"
)

func setupTestDB(t *testing.T) *gorm.DB{

	_ = godotenv.Load("../../.env")
	databaseURL :=os.Getenv("DATABASE_URL")

	db,err :=gorm.Open(postgres.Open(databaseURL),&gorm.Config{})
	require.NoError(t,err,"failed to connect to db")

	

	return db
}

func cleanupTestDB(t *testing.T, db *gorm.DB){
	err :=db.Exec("TRUNCATE TABLE issues RESTART IDENTITY CASCADE").Error
	require.NoError(t,err,"failed to clean up test db")
}

func TestIssueRepositiry_Create(t *testing.T){
	db:=setupTestDB(t)
	defer cleanupTestDB(t,db)

	repo:=NewIssueRepository(db)

	ctx:=context.Background()

	userID:=uuid.New()
	issueID :=uuid.New().String()

	imageURLs :=[]string{
		"example.jgp",
		"expale.jpg",
	}
	imageURLsJSON,err :=json.Marshal(imageURLs)
	require.NoError(t,err)

	 orderDetails := map[string]interface{}{
        "id":          "order-123",
        "provider_id": "provider-456",
        "items": []map[string]interface{}{
            {"id": "item-1", "quantity": 2},
        },
    }
    orderDetailsJSON, err := json.Marshal(orderDetails)
    require.NoError(t, err)

	complaintActions :=[]map[string]interface{}{
		{
			"complaint_action":"OPEN",
			"short_desc":"item damage",
			"updated_at":time.Now().Format(time.RFC3339),
		},
	}

	complaintActionsJSON,err :=json.Marshal(complaintActions)
	require.NoError(t,err)

	issue :=&models.Issue{
		IssueID: issueID,
		OrderID: "order-123",
		UserID: userID,
		TransactionID: uuid.New().String(),
		BPPID: "bpp-test-id",
		BPPURI: "https://bpp-test.com",
		Category: "ITEM",
		SubCategory: "ITM01",
		IssueType: "ISSUE",
		Status: "OPEN",
		DescriptionShort: "damaged ",
		DescriptionLong: "prodcut was dmaaged",
		DescriptionContentType: "applicatuib/pdf",
		Images: datatypes.JSON(imageURLsJSON),
		OrderDetails: datatypes.JSON(orderDetailsJSON),
		ComplainantActions: datatypes.JSON(complaintActionsJSON),
		SourceNPID: "prepod.effimove.in",
		SourceType: "CONSUMER",
		ExpectedResponseTime: "PT2H",
		ExpectedResolutionTime: "P1D",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		
	}

	err=repo.Create(ctx,issue)
	assert.NoError(t,err,"create should not return error")

	assert.NotZero(t,issueID,"issue id should be auto generated")

	var savedIssue models.Issue
	err=db.Where("issue_id=?",issueID).First(&savedIssue).Error
	assert.NoError(t,err,"should find saved issue")

	assert.Equal(t,issueID,savedIssue.IssueID)
	assert.Equal(t,"order-123",savedIssue.OrderID)
	assert.Equal(t,userID,savedIssue.UserID)
	assert.Equal(t,"ITEM",savedIssue.Category)
	assert.Equal(t,"OPEN",savedIssue.Status)

	 var savedImages []string
    err = json.Unmarshal(savedIssue.Images, &savedImages)
    assert.NoError(t, err)
    assert.Equal(t, imageURLs, savedImages)
}
func TestIssueRepository_Create_DuplicateIssueID(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    repo := NewIssueRepository(db)
    ctx := context.Background()
    
    issueID := uuid.New().String()
    

    issue1 := &models.Issue{
        IssueID:       issueID,
        OrderID:       "order-123",
        UserID:        uuid.New(),
        TransactionID: uuid.New().String(),
        BPPID:         "bpp-1",
        BPPURI:        "https://bpp1.com",
        Category:      "ITEM",
        Status:        "OPEN",
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }
    
    err := repo.Create(ctx, issue1)
    assert.NoError(t, err)
    
    // Try to create duplicate
    issue2 := &models.Issue{
        IssueID:       issueID, 
        OrderID:       "order-456",
        UserID:        uuid.New(),
        TransactionID: uuid.New().String(),
        BPPID:         "bpp-2",
        BPPURI:        "https://bpp2.com",
        Category:      "FULFILLMENT",
        Status:        "OPEN",
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }
    
    err = repo.Create(ctx, issue2)
    assert.Error(t, err, "Should fail due to unique constraint on issue_id")
    assert.Contains(t, err.Error(), "duplicate key", "Error should mention duplicate key")
}
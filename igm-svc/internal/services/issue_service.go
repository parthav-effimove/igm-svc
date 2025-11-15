package services

import (
	"context"
	"encoding/json"
	"fmt"
	"igm-svc/internal/models"
	"igm-svc/internal/repository"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type IssueService struct {
	issueRepo repository.IssueRepository
	redisRepo repository.RedisRepository
	config    *Config
}

type Config struct {
	SubcriberID string
	BAPURI      string
}

func NewIssueService(issueRepo repository.IssueRepository,
	redisRepo repository.RedisRepository,
	config *Config,
)*IssueService{
	return &IssueService{
		issueRepo: issueRepo,
		redisRepo: redisRepo,
		config: config,
	}
}

func(s *IssueService) CreateIssue(ctx context.Context,req *models.IssueCreateRequest,userID uuid.UUID)(*models.Issue,error){

	issue, err:=s.buildIssueFromRequest(req,userID)
	if err!=nil{
		return nil,fmt.Errorf("failed to build issue:%w",err)
	}

	err=s.issueRepo.Create(ctx,issue)
	if err!=nil{
		return nil,fmt.Errorf("failed to save issue :%w",err)
	}
	log.Printf("issue saved to DB :%s",issue.IssueID)

	err=s.se

}

func(s *IssueService)buildIssueFromRequest(req *models.IssueCreateRequest,userID uuid.UUID)(*models.Issue,error){
	now :=time.Now()
	issueID :=uuid.New().String()

	// TODO: Fetch order details from Order Service to get BPPID, BPPURI

	bppID:="demo-bpp-id"
	bppURI := "https://demo-bpp.com"

	imagesJSON, err := json.Marshal(req.ImageURLs)
	if err!=nil{
		return nil,fmt.Errorf("failed to marshal images:%w",err)
	}

	orderDetailsMap :=map[string]interface{}{
		"id": req.OrderID,
		"provider_id":bppID,
		"items":req.Items,
	}

	orderDetailsJSON, err :=json.Marshal(orderDetailsMap)
	if err !=nil{
		return nil,fmt.Errorf("failed to marhsal order details :%w",err)
	}

	complaintAction :=map[string]interface{}{
		"complaint_action":"OPEN",
		"short_desc":req.Description,
		"updated_at":now.Format(time.RFC3339),
		"updated_by":map[string]interface{}{
			"org":map[string]interface{}{
				"name":s.config.SubcriberID,
			},
			"contact":map[string]interface{}{
				"phone":"",
				"email":"",// TODO: Get from user profile
			},
			"person":map[string]interface{}{
				"name":"",
			},
		},
	}

	complaintActionJSON,err :=json.Marshal([]interface{}{complaintAction})
	if err!=nil{
		return  nil,fmt.Errorf("failed to marshal complaint action:%w",err)
	}

	issue :=&models.Issue{
		IssueID: issueID,
		OrderID: req.OrderID,
		UserID: userID,
		TransactionID: uuid.New().String(),
		BPPID: bppID,
		BPPURI: bppURI,
		Category: req.Category,
		SubCategory: req.SubCategory,
		IssueType: req.IssueType,
		Status: "OPEN",
		DescriptionShort: req.Description,
		DescriptionLong: req.LongDesc,
		DescriptionURL: req.AdditionalDesc.Url,
		DescriptionContentType: req.AdditionalDesc.ContentType,
		Images: datatypes.JSON(imagesJSON),
		OrderDetails: datatypes.JSON(orderDetailsJSON),
		ComplainantActions: datatypes.JSON(complaintActionJSON),
		SourceNPID: s.config.SubcriberID,
		SourceType: "CONSUMER",
		ExpectedResponseTime: "PT2H",
		ExpectedResolutionTime: "P1D",
		Rating: req.Rating,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return issue,nil
}

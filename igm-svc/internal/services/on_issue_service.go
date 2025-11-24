package services

import (
	"context"
	"fmt"
	pb "igm-svc/api/proto/igm/v1"
	"igm-svc/internal/repository"
	"log"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/datatypes"
)

type OnIssueService struct {
	onIssueRepo repository.OnIssueRepository
	redisRepo   repository.RedisRepository
	OndcClient  *OndcClient
	config      *Config
}

func NewOnIssueService(onIssueRepo repository.OnIssueRepository,
	redisRepo repository.RedisRepository,
	ondcClient *OndcClient,
	config *Config) *OnIssueService {
	return &OnIssueService{
		onIssueRepo: onIssueRepo,
		redisRepo:   redisRepo,
		OndcClient:  ondcClient,
		config:      config,
	}
}

func (h *OnIssueService) ProcessOnIssue(ctx context.Context, transactionID, messageID string, payload *pb.OnIssuePayload) error {
	if payload == nil {
		return fmt.Errorf("nil payload")
	}
	marshaler := protojson.MarshalOptions{EmitUnpopulated: false}
	raw,err:=marshaler.Marshal(payload)
	//todo ondccallback
	err = h.onIssueRepo.SaveOnIssueCallback(ctx, transactionID, messageID, raw)
	if err != nil{
		
		log.Printf("warn: SaveOndcCallback returned: %v", err)
	}

	if payload.GetIssue()==nil || payload.Issue.GetId()==""{
		return fmt.Errorf("payload missing issue.id")
	}
	issueID :=payload.Issue.Id

	updates :=map[string]interface{}{}

	now:=time.Now()

	

	 ia :=payload.Issue.GetIssueActions()
	 if ia != nil && len(ia.GetRespondentActions())>0{
		if b, err :=marshaler.Marshal(ia); err ==nil{
			updates["respondent_actions"]=datatypes.JSON(b)
		}else{
			log.Printf("warn:failed to marshal respondent action :%v",err)
		}
		
		last :=ia.RespondentActions[len(ia.RespondentActions)-1]
		if last !=nil && last.GetRespondentAction() !=""{
			updates["status"]=last.GetRespondentAction()
		}
	 }

	 rp :=payload.Issue.GetResolutionProvider()
	 if rp!=nil{
		if b,err:=marshaler.Marshal(rp);err==nil{
			updates["resolution_provider"]=datatypes.JSON(b)
		}else{
			log.Printf("warn: failed to marshal resolution_provider: %v",err)
		}
	 }

	 res :=payload.Issue.GetResolution()
	 if res!=nil{
		if b,err:=marshaler.Marshal(res);err==nil{
			updates["resolution"]=datatypes.JSON(b)
		}else{
			log.Printf("warn: failed to marshal resolution:%v",err)
		}
		if res.GetRefundAmount() !=""{
			updates["refund_amount"]=res.GetRefundAmount()
		}
	 }

	 updates["updated_at"]=now

	 err=h.onIssueRepo.UpdateIssueFromOnIssue(ctx,issueID,updates)
	 if err!=nil{
		return fmt.Errorf("failed to update issue from on_issue:%w",err)
	 }

	 return nil

}

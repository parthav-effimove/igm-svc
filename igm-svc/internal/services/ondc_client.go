package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"igm-svc/internal/models"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type OndcClient struct {
	httpClient   *http.Client
	subscriberID string
	bapURI       string
}

func NewOndcClient(subscriberID, bapURI string) *OndcClient {
	return &OndcClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		subscriberID: subscriberID,
		bapURI:       bapURI,
	}
}

func (c *OndcClient) SendIssue(ctx context.Context, issue *models.Issue, operation string) error {
	log.Printf("sending issue to BPP :%s", issue.BPPURI)

	payload, err := c.buildIssuePayload(issue, operation)
	if err != nil {
		return fmt.Errorf("failed to build the payload :%w", err)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	log.Printf("[ONDC] Request payload: %s", string(body))
	authHeader, err := c.createAuthHeader(body)
	if err != nil {
		return fmt.Errorf("failed to create auth header: %w", err)
	}

	url := issue.BPPURI
	if url[len(url)-1] != '/' {
		url += "/"
	}
	url += "issue"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", authHeader)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("BPP response status: %d", resp.StatusCode)
	log.Printf("BPP response body: %s", string(respBody))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("BPP returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil

}

func (c *OndcClient) buildIssuePayload(issue *models.Issue, operation string) (map[string]interface{}, error) {
	ctx := map[string]interface{}{
		"domain":         "nic2004:60232",
		"country":        "IND",
		"city":           "std:080",
		"action":         "issue",
		"core_version":   "1.2.0",
		"bap_id":         c.subscriberID,
		"bap_uri":        c.bapURI,
		"bpp_id":         issue.BPPID,
		"bpp_uri":        issue.BPPURI,
		"transaction_id": issue.TransactionID,
		"message_id":     uuid.New().String(),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"ttl":            "PT30S",
	}
	issueBody, err := c.mapIssueToONDCFormat(issue, operation)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"context": ctx,
		"message": map[string]interface{}{
			"issue": issueBody,
		},
	}, nil
}

func (c *OndcClient) mapIssueToONDCFormat(issue *models.Issue, operation string) (map[string]interface{}, error) {
	var images []string
	if len(issue.Images) > 0 {
		if err := json.Unmarshal(issue.Images, &images); err != nil {
			return nil, fmt.Errorf("failed to unmarshal images: %w", err)
		}
	}

	var orderDetails map[string]interface{}
	if len(issue.OrderDetails) > 0 {
		if err := json.Unmarshal(issue.OrderDetails, &orderDetails); err != nil {
			return nil, fmt.Errorf("failed to unmarshal order details: %w", err)
		}
	}

	var complainantActions []map[string]interface{}
	if len(issue.ComplainantActions) > 0 {
		if err := json.Unmarshal(issue.ComplainantActions, &complainantActions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal complainant actions: %w", err)
		}
	}

	baseIssue := map[string]interface{}{
		"id":         issue.IssueID,
		"created_at": issue.CreatedAt.Format(time.RFC3339),
		"updated_at": issue.UpdatedAt.Format(time.RFC3339),
	}

	switch operation {
	case "OPEN":
		baseIssue["category"] = issue.Category
		baseIssue["sub_category"] = issue.SubCategory
		baseIssue["complainant_info"] = map[string]interface{}{
			"person": map[string]interface{}{
				"name": issue.UserName,
			},
			"contact": map[string]interface{}{
				"phone": issue.UserPhone,
				"email": issue.UserEmail,
			},
		}
		baseIssue["order_details"] = orderDetails
		baseIssue["description"] = map[string]interface{}{
			"short_desc": issue.DescriptionShort,
			"long_desc":  issue.DescriptionLong,
			"additional_desc": map[string]interface{}{
				"url":          issue.DescriptionURL,
				"content_type": issue.DescriptionContentType,
			},
			"images": images,
		}
		baseIssue["source"] = map[string]interface{}{
			"network_participant_id": issue.SourceNPID,
			"type":                   issue.SourceType,
		}
		baseIssue["expected_response_time"] = map[string]interface{}{
			"duration": issue.ExpectedResponseTime,
		}
		baseIssue["expected_resolution_time"] = map[string]interface{}{
			"duration": issue.ExpectedResolutionTime,
		}
		baseIssue["status"] = issue.Status
		baseIssue["issue_type"] = issue.IssueType
		baseIssue["issue_actions"] = map[string]interface{}{
			"complainant_actions": complainantActions,
		}

	case "CLOSE":
		baseIssue["status"] = "CLOSED"
		baseIssue["rating"] = issue.Rating
		baseIssue["issue_actions"] = map[string]interface{}{
			"complainant_actions": complainantActions,
		}

	case "ESCALATE":
		baseIssue["status"] = issue.Status
		baseIssue["issue_type"] = issue.IssueType
		baseIssue["issue_actions"] = map[string]interface{}{
			"complainant_actions": complainantActions,
		}
	}

	return baseIssue, nil
}

func (c *OndcClient) createAuthHeader(body []byte) (string, error) {
	// Placeholder - return dummy signature
	return "Signature keyId=\"preprod.effimove.in|unique-key-id|ed25519\",algorithm=\"ed25519\",created=\"timestamp\",expires=\"timestamp\",headers=\"(created) (expires) digest\",signature=\"base64-signature\"", nil
}

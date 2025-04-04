package api

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testStand/internal/acquirer/helper"
	"time"
)

type Client struct {
	baseAddress string
	apiKey      string
	client      *http.Client
}

const (
	authEndpoint    = "fim/est3dgate"
	requestEndpoint = "fim/api"
)

func NewClient(baseAddress, apiKey string, timeout *int) *Client {
	client := http.DefaultClient
	if timeout != nil {
		client.Timeout = time.Duration(*timeout) * time.Second
	}
	return &Client{
		baseAddress: baseAddress,
		apiKey:      apiKey,
		client:      client,
	}
}

func (c *Client) EstAuth(ctx context.Context, req *Est3DGateRequest, storeKey string) (string, string, string, string, error) {
	req.Hash = GenerateHash(req, storeKey)

	formData := url.Values{
		"clientid":                        {req.ClientId},
		"storetype":                       {req.StoreType},
		"trantype":                        {req.TranType},
		"currency":                        {req.Currency},
		"oid":                             {req.Oid},
		"rnd":                             {req.Rnd},
		"pan":                             {req.Pan},
		"Ecom_Payment_Card_ExpDate_Month": {req.EcomPaymentCardExpDateMonth},
		"Ecom_Payment_Card_ExpDate_Year":  {req.EcomPaymentCardExpDateYear},
		"cv2":                             {req.Cv2},
		"hashAlgorithm":                   {req.HashAlgorithm},
		"hash":                            {req.Hash},
		"encoding":                        {req.Encoding},
	}

	log.Printf("Sending request to /fim/est3dgate: %s", formData.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, helper.JoinUrl(c.baseAddress, authEndpoint), strings.NewReader(formData.Encode()))
	if err != nil {
		return "", "", "", "", err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", "", "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to read response body: %v", err)
	}
	log.Printf("Response from /fim/est3dgate: status=%d, body=%s", resp.StatusCode, string(body))

	parsed, err := ParseEstResponse(string(body))
	if err != nil {
		return "", "", "", "", err
	}

	log.Printf("Parsed response: %+v", parsed)

	md := parsed["md"]
	cavv := parsed["cavv"]
	eci := parsed["eci"]
	xid := parsed["xid"]
	mdStatus := parsed["mdStatus"]

	if mdStatus != "1" {
		errMsg := parsed["ErrMsg"]
		return "", "", "", "", fmt.Errorf("Auth failed, mdStatus: %s, ErrMsg: %s", mdStatus, errMsg)
	}

	return md, cavv, eci, xid, nil
}

func (c *Client) MakePayment(ctx context.Context, request *Request) (*Response, error) {
	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, requestEndpoint)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, endpoint string) error {
	body, err := xml.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	log.Printf("Sending request to /fim/api: %s", string(body))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %v", err)
	}
	log.Printf("Response: status=%d, body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("something went wrong: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	err = xml.NewDecoder(bytes.NewReader(respBody)).Decode(outResponse)
	if err != nil {
		return fmt.Errorf("failed to decode, response: %v, body=%s", err, string(respBody))
	}

	return nil
}

func ParseEstResponse(response string) (map[string]string, error) {
	values, err := url.ParseQuery(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	keys := []string{"md", "cavv", "eci", "xid", "mdStatus", "ErrMsg"}
	result := make(map[string]string)

	for _, key := range keys {
		if value := values.Get(key); value != "" {
			result[key] = value
		}
	}

	return result, nil
}

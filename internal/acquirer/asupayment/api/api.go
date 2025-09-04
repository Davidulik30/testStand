package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"testStand/internal/acquirer/helper"
)

const (
	PaymentEndpoint = "deposit"
	PayoutEndpoint  = "withdraw"
	StatusEndpoint  = "status"
)

type Client struct {
	apikey      string
	baseAddress string
	secretkey   string
	client      *http.Client
	Timeout     *int
}

func NewClient(ctx context.Context, baseAddress, SecretKey string, ApiKey string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		apikey:      ApiKey,
		baseAddress: baseAddress,
		secretkey:   SecretKey,
		client:      client,
		Timeout:     timeout,
	}
}

func (c *Client) MakeDeposit(ctx context.Context, request Request) (*Response, error) {
	if c.secretkey != "" {
		request.Signature = CalcSignature(request.Merchant, request.CardData.CardNumber, request.Amount, c.secretkey)
	}
	resp := &Response{}
	err := c.makeRequest(ctx, request, PaymentEndpoint, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) MakeWithdraw(ctx context.Context, request Request) (*Response, error) {
	if c.secretkey != "" {
		request.Signature = CalcSignature(request.Merchant, request.CardData.CardNumber, request.Amount, c.secretkey)
	}
	resp := &Response{}
	err := c.makeRequest(ctx, request, PayoutEndpoint, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CheckStatus(ctx context.Context, request StatusRequest) (*StatusResponse, error) {

	if c.secretkey != "" {
		request.Sign = CalcSignature(request.MerchId, request.ID, "", c.secretkey)
	}
	resp := &StatusResponse{}
	err := c.makeRequest(ctx, request, StatusEndpoint, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) makeRequest(ctx context.Context, payload any, endpoint string, outResponse any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := helper.JoinUrl(c.baseAddress, endpoint)

	log.Printf("[asupayment] Final request URL: %s", url)
	log.Printf("[asupayment] Final request JSON: %s", string(body))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	if c.apikey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apikey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Server response status: %s", resp.Status)
	log.Printf("Server response body: %s", string(respBody))

	if resp.StatusCode >= http.StatusInternalServerError {
		return errors.New("declined state due to network or internal error")
	}

	err = json.Unmarshal(respBody, outResponse)
	if err != nil {
		return err
	}
	return nil
}

package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"testStand/internal/acquirer/helper"
)

const (
	OfferEndpoint   = "offer/external"
	SignGenEndpoint = "user/generate-signature-key"
)

type Client struct {
	email        string
	password     string
	apikey       string
	baseAddress  string
	signatureKey string
	client       *http.Client
	Timeout      *int
}

func NewClient(ctx context.Context, baseAddress, Email string, Password string, ApiKey string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		email:       Email,
		password:    Password,
		apikey:      ApiKey,
		baseAddress: baseAddress,
		client:      client,
		Timeout:     timeout,
	}
}

func (c *Client) MakeTransaction(ctx context.Context, request Request) (*Response, error) {
	log.Printf("[aplex][MakeTransaction] Входные данные: %+v", request)
	if request.Direction == "BUY" {
		log.Printf("[aplex][MakeTransaction] Тип транзакции: BUY (Payment)")
	} else if request.Direction == "SELL" {
		log.Printf("[aplex][MakeTransaction] Тип транзакции: SELL (Payout)")
	} else {
		log.Printf("[aplex][MakeTransaction] Неизвестный тип транзакции: %s", request.Direction)
	}
	resp := &Response{}
	err := c.makeRequest(ctx, request, OfferEndpoint, resp)
	if err != nil {
		log.Printf("[aplex][MakeTransaction] Ошибка: %v", err)
		return nil, err
	}
	log.Printf("[aplex][MakeTransaction] Ответ: %+v", resp)
	return resp, nil
}

func (c *Client) makeRequest(ctx context.Context, payload any, endpoint string, outResponse any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[aplex][makeRequest] Ошибка маршалинга payload: %v", err)
		return err
	}

	log.Printf("[aplex] Final request URL: %s", helper.JoinUrl(c.baseAddress, endpoint))
	log.Printf("[aplex] Final request JSON: %s", string(body))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewReader(body))
	if err != nil {
		log.Printf("[aplex][makeRequest] Ошибка создания запроса: %v", err)
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

	if resp.StatusCode >= http.StatusInternalServerError {
		return errors.New("declined state due to network or internal error")
	}

	err = json.NewDecoder(resp.Body).Decode(&outResponse)
	if err != nil {
		return nil // error EOF, because invalid url
	}
	return nil
}

func (c *Client) ValidateSignature(id string, status string, hash string, ctx context.Context) (bool, error) {

	err := c.GetSignatureKey(ctx)
	if err != nil {
		return false, err
	}

	signCalculated := hmac.New(sha256.New, []byte(c.signatureKey))
	if _, err := signCalculated.Write([]byte(fmt.Sprintf("id=%s\nstatus=%s", id, status))); err != nil {
		return false, err
	}
	cbSignHex, err := hex.DecodeString(hash)
	if err != nil {
		return false, err
	}
	return hmac.Equal(cbSignHex, signCalculated.Sum(nil)), nil
}

func (c *Client) GetSignatureKey(ctx context.Context) error {

	request := UserLoad{
		Email:    c.email,
		Password: c.password,
	}
	resp := &Response{}
	err := c.makeRequest(ctx, request, SignGenEndpoint, resp)

	if err != nil {
		log.Printf("[aplex][GetSignatureKey] Ошибка: %v", err)
		return err
	}
	c.signatureKey = resp.Signatrue

	return nil
}

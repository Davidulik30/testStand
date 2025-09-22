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
	"io"
	"log"
	"net/http"
	"testStand/internal/acquirer/helper"
)

const (
	OfferEndpoint = "offer/external"
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

// Универсальный метод для транзакций (BUY/SELL)
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

	url := helper.JoinUrl(c.baseAddress, endpoint)

	log.Printf("[aplex] Final request URL: %s", url)
	log.Printf("[aplex] Final request JSON: %s", string(body))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
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

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= http.StatusInternalServerError {
		return errors.New("declined state due to network or internal error")
	}

	err = json.Unmarshal(respBody, outResponse)
	if err != nil {
		return err
	}
	return nil
}

func ValidateSignature(id, status, hash, secretKey string) (bool, error) {
	signCalculated := hmac.New(sha256.New, []byte(secretKey))
	if _, err := signCalculated.Write([]byte(fmt.Sprintf("id=%s\nstatus=%s", id, status))); err != nil {
		return false, err
	}
	cbSignHex, err := hex.DecodeString(hash)
	if err != nil {
		return false, err
	}
	return hmac.Equal(cbSignHex, signCalculated.Sum(nil)), nil
}

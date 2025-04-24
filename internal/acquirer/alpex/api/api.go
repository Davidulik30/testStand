package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testStand/internal/acquirer/helper"
)

type Client struct {
	baseAddress  string
	email        string
	password     string
	token        string
	SignatureKey string
	client       *http.Client
}

const (
	offerEndpoint        = "v1/offer/external"
	signatureKeyEndpoint = "/v1/user/generate-signature-key"
	accessTokenEndpoint  = "/v1/auth/login"
)

func NewClient(ctx context.Context, baseAddress, email, password string) *Client {

	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		email:       email,
		password:    password,
		client:      client,
	}

}

// MakePayout
func (c *Client) MakePayout(ctx context.Context, request *Request) (*Response, error) {

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, offerEndpoint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// MakePayment
func (c *Client) MakePayment(ctx context.Context, request *Request) (*Response, error) {

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, offerEndpoint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// makeRequest
func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, endpoint string) error {

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewReader(body))
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if len(c.token) != 0 {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&outResponse)
	if err != nil {
		return err // error EOF, because invalid url
	}

	return nil
}

func (c *Client) GetSignatureKey(ctx context.Context) error {

	user := User{
		Email:    c.email,
		Password: c.password,
	}

	keyMap := make(map[string]string)
	err := c.makeRequest(ctx, user, &keyMap, signatureKeyEndpoint)
	if err != nil {
		return err
	}
	c.SignatureKey = keyMap["signature_key"]

	return nil
}

func (c *Client) GetAccessToken(ctx context.Context) error {

	user := User{
		Email:    c.email,
		Password: c.password,
	}

	tokenMap := make(map[string]string)
	err := c.makeRequest(ctx, user, &tokenMap, accessTokenEndpoint)
	if err != nil {
		return err
	}
	c.token = tokenMap["access_token"]

	return nil
}

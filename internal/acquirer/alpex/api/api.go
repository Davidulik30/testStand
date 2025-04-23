package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testStand/internal/acquirer/helper"
)

type Client struct {
	baseAddress string
	accessToken string
	email       string
	password    string
	Id          string
	client      *http.Client
}

const (
	PayoutEndpoint       = "v1/offer/external"
	PaymentEndpoint      = "v1/offer/external"
	SignatureKeyEndpoint = "/v1/user/generate-signature-key"
	AccessTokenEndpoint  = "/v1/auth/login"
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
	err := c.makeRequest(ctx, request, resp, PayoutEndpoint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// MakePayment
func (c *Client) MakePayment(ctx context.Context, request *Request) (*Response, error) {

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, PaymentEndpoint)
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

	token, err := c.getAccessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&outResponse)
	if err != nil {
		return nil // error EOF, because invalid url
	}

	return nil
}

func (c *Client) GetSignatureKey(ctx context.Context) (*UserSignToken, error) {

	user := User{
		Email:    c.email,
		Password: c.password,
	}

	key := &UserSignToken{}

	err := c.makeRequest(ctx, user, key, SignatureKeyEndpoint)

	if err != nil {
		return nil, err
	}

	return key, nil
}

func (c *Client) getAccessToken(ctx context.Context) (*UserSignToken, error) {

	user := User{
		Email:    c.email,
		Password: c.password,
	}

	token := &UserSignToken{}

	body, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, AccessTokenEndpoint), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil, err // error EOF, because invalid url
	}

	return token, nil
}

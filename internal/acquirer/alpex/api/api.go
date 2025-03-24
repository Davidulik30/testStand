package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testStand/internal/acquirer/helper"
)

type Client struct {
	baseAddress string
	client      *http.Client
	login       string
	password    string
}

const (
	loginEnd     = "v1/auth/login"
	offerEnd     = "v1/offer/external"
	signatureEnd = "v1/user/generate-signature-key"
)

func NewClient(baseAddress, login string, password string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		client:      client,
		login:       login,
		password:    password,
	}
}

// MakePayout
func (c *Client) MakePayInOut(request *Request) (*Response, error) {

	resp := &Response{}
	err := c.makeRequest(request, resp, offerEnd)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// makeRequest
func (c *Client) makeRequest(payload, outResponse any, endpoint string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	apiKey, err := c.GetApiKey()
	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

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

func (c *Client) CheckCallBack(callback *Callback) (bool, error) {
	apiKey, err := c.GetApiKey()
	if err != nil {
		return false, err
	}
	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, signatureEnd), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()
	var result SignatureKey
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	signCounted := GenerateSignature(callback, result.SignatureKey)
	if signCounted != callback.Signature {
		return false, errors.New("signatures do not match")
	}

	return true, nil
}

func (c *Client) GetApiKey() (string, error) {
	data := ApiKeyRequest{
		Email:    c.login,
		Password: c.password,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	body := bytes.NewBuffer(jsonData)

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, loginEnd), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result AccessToken
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

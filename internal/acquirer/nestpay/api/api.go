package api

import (
	"bytes"
	"context"
	"encoding/xml"
	"github.com/go-playground/form"
	"github.com/google/go-querystring/query"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testStand/internal/acquirer/helper"
)

type Client struct {
	baseAddress string
	currency    string
	client      *http.Client
}

const (
	paymentEndpoint     = "/fim/est3dgate"
	checkStatusEndpoint = "/fim/api"
)

func NewClient(ctx context.Context, baseAddress string) *Client {

	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		client:      client,
	}
}

// MakePayment
func (c *Client) MakePayment(ctx context.Context, request *Request) (*PaymentResponse, error) {

	values, err := query.Values(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, paymentEndpoint), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoder := form.NewDecoder()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	respValues, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		return nil, err
	}

	paymentResp := &PaymentResponse{}

	err = decoder.Decode(paymentResp, respValues)
	if err != nil {
		return nil, err
	}

	return paymentResp, nil
}

func (c *Client) CheckStatus(ctx context.Context, request *StatusRequest) (*StatusResponse, error) {

	reqBody, err := xml.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, checkStatusEndpoint), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	statusResp := &StatusResponse{}
	err = xml.NewDecoder(resp.Body).Decode(statusResp)

	return statusResp, nil
}

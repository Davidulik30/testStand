package api

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testStand/internal/acquirer/helper"
)

const (
	PaymentEndpoint = "/fim/est3dgate"
	AuthEndpoint    = "/fim/api"
)

type Client struct {
	baseAddress string
	client      *http.Client
	storekey    string
}

func NewClient(ctx context.Context, baseAddress, storekey string) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		client:      client,
		storekey:    storekey,
	}
}

func (c *Client) MakeTransaction(ctx context.Context, payload Request) (*PaymentResponse, error) {
	hashParams := map[string]string{
		"ClientId":                        payload.ClientId,
		"oId":                             payload.OId,
		"storetype":                       payload.StoreType,
		"lang":                            payload.Lang,
		"amount":                          payload.Amount,
		"currency":                        payload.Currency,
		"okUrl":                           payload.OkUrl,
		"failUrl":                         payload.FailUrl,
		"trantype":                        payload.TransactionType,
		"Ecom_Payment_Card_ExpDate_Year":  payload.EcomPaymentCardExpDateYear,
		"Ecom_Payment_Card_ExpDate_Month": payload.EcomPaymentCardExpDateMonth,
		"rnd":                             payload.Rnd,
		"hashAlgorithm":                   payload.HashAlgorithm,
		"cv2":                             payload.Cv2,
		"pan":                             payload.Pan,
		"encoding":                        payload.Encoding,
	}

	form := url.Values{}
	for k, v := range hashParams {
		form.Set(k, v)
	}
	form.Set("hash", c.GenerateNestPayHash(hashParams, c.storekey))

	reqUrl := helper.JoinUrl(c.baseAddress, PaymentEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Server response status: %s", resp.Status)
	log.Printf("Server response body: %s", string(respBody))

	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, errors.New("declined state due to network or internal error")
	}

	paymentResp, err := c.ParsePaymentResponse(string(respBody))
	if err != nil {
		return nil, err
	}
	return paymentResp, nil
}

func (c *Client) OrderStatusQuery(ctx context.Context, order OrderStatus) (*StatusResponse, error) {
	xmlBody, err := xml.Marshal(order)
	if err != nil {
		return nil, err
	}

	url := helper.JoinUrl(c.baseAddress, AuthEndpoint)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(xmlBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/xml")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var statusResp StatusResponse
	if err := xml.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, err
	}
	return &statusResp, nil
}

func (c *Client) GenerateNestPayHash(params map[string]string, storeKey string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return strings.ToLower(keys[i]) < strings.ToLower(keys[j])
	})

	escapedStoreKey := strings.ReplaceAll(storeKey, "\\", "\\\\")
	escapedStoreKey = strings.ReplaceAll(escapedStoreKey, "|", "\\|")

	var sb strings.Builder
	for _, key := range keys {
		lowerKey := strings.ToLower(key)
		if lowerKey == "encoding" || lowerKey == "hash" || lowerKey == "countdown" {
			continue
		}

		value := params[key]
		escapedValue := strings.ReplaceAll(value, "\\", "\\\\")
		escapedValue = strings.ReplaceAll(escapedValue, "|", "\\|")

		sb.WriteString(escapedValue)
		sb.WriteString("|")
	}

	sb.WriteString(escapedStoreKey)

	hashBytes := sha512.Sum512([]byte(sb.String()))
	hashBase64 := base64.StdEncoding.EncodeToString(hashBytes[:])

	return hashBase64
}

func (c *Client) ParsePaymentResponse(raw string) (*PaymentResponse, error) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return nil, err
	}

	resp := &PaymentResponse{
		AcqBin:               values.Get("ACQBIN"),
		ExpMonth:             values.Get("Ecom_Payment_Card_ExpDate_Month"),
		ErrMsg:               values.Get("ErrMsg"),
		PAResVerified:        values.Get("PAResVerified"),
		ACSReferenceNumber:   values.Get("TDS2.acsReferenceNumber"),
		AuthTimestamp:        values.Get("TDS2.authTimestamp"),
		ThreeDSServerTransID: values.Get("TDS2.threeDSServerTransID"),
		TransStatus:          values.Get("TDS2.transStatus"),
		ThreedID:             values.Get("THREED_ID"),
		TransID:              values.Get("TransID"),
		Amount:               values.Get("amount"),
		CAVV:                 values.Get("cavv"),
		ECI:                  values.Get("eci"),
		MD:                   values.Get("md"),
		MDStatus:             values.Get("mdStatus"),
		OrderID:              values.Get("oid"),
		Currency:             values.Get("currency"),
		ClientID:             values.Get("clientid"),
		MerchantID:           values.Get("merchantID"),
		StoreType:            values.Get("storetype"),
		TransactionStatus:    values.Get("paresTxStatus"),
		ThreeDSError:         values.Get("mdErrorMsg"),
		ClientIP:             values.Get("clientIp"),
		XID:                  values.Get("xid"),
	}

	return resp, nil
}

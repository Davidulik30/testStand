package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	StatusPending  = "PENDING"
	StatusReleased = "RELEASED"
	StatusDeclined = "DECLINED"
	Payout         = "SELL"
	Payment        = "BUY"
)

type AccessToken struct {
	AccessToken string `json:"access_token"`
}
type SignatureKey struct {
	SignatureKey string `json:"signature_key"`
}
type ApiKeyRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type Request struct {
	FiatSymbol      string `json:"fiat_symbol"`
	FiatAmount      string `json:"fiat_amount"`
	CustomerName    string `json:"customer_name"`
	CustomerAddress string `json:"customer_address"`
	Direction       string `json:"direction"`
	GateID          string `json:"gate_id"`
	ExternalID      string `json:"external_id"`
	Webhook         string `json:"webhook_url"`
}

type Gate struct {
	Id   string `json:"_id"`
	Name string `json:"name"`
}

type PayMethod struct {
	Id      string `json:"_id"`
	Gate    Gate   `json:"gate"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Person  string `json:"person"`
	IsTemp  bool   `json:"is_temporary"`
}

type Response struct {
	Id         string    `json:"_id"`
	PayMethod  PayMethod `json:"payment_method"`
	Direction  string    `json:"direction"`
	Amount     string    `json:"amount"`
	Status     string    `json:"status"`
	Signature  string    `json:"signature"`
	ExternalId string    `json:"external_id"`
	Error      string    `json:"error"`
}

type Callback struct {
	Id         string `json:"_id"`
	Status     string `json:"status"`
	Signature  string `json:"signature"`
	ExternalId string `json:"external_id"`
}

func GenerateSignature(callback *Callback, secretKey string) string {
	sum := fmt.Sprintf("id=%s\nstatus=%s", callback.Id, callback.Status)
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(sum))
	return hex.EncodeToString(h.Sum(nil))
}

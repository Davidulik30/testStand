package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	Pending  = "PENDING"
	Released = "RELEASED"
	Declined = "DECLINED"
)

type Request struct {
	CustomerName    string `json:"customer_name"`
	CustomerAddress string `json:"customer_address"`
	Direction       string `json:"direction"`
	Amount          int64  `json:"fiat_amount"`
	Symbol          string `json:"fiat_symbol"`
	GateId          string `json:"gate_id"`
	WebhookUrl      string `json:"webhook_url"`
	Signature       string `json:"signature"`
}

type Response struct {
	Status        string        `json:"status"`
	Id            string        `json:"_id"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Signature     string        `json:"signature"`
	Direction     string        `json:"direction"`
	Amount        int           `json:"amount"`
	AmountFiat    int           `json:"amount_fiat"`
	ApproveCode   string        `json:"approve_code"`
}

type PaymentMethod struct {
	Id      string `json:"_id"`
	Gate    Gate   `json:"gate"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Person  string `json:"person"`
}

type Gate struct {
	Id       string `json:"_id"`
	BankName string `json:"name"`
}

type Callback struct {
	Id               string `json:"_id"`
	UserRef          string `json:"user_ref"`
	Status           string `json:"status"`
	Description      string `json:"description"`
	TimestampUpdated string `json:"timestamp_updated"`
	Amount           string `json:"amount"`
	Sign             string `json:"signature"`
}

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserSignToken struct {
	SignKey string `json:"signature_key"`
}

func CreateSign(id, status, key string) string {
	hashString := fmt.Sprintf("id=%s\nstatus=%s", id, status)
	hmac := hmac.New(sha256.New, []byte(key))
	hmac.Write([]byte(hashString))
	sign := hex.EncodeToString(hmac.Sum(nil))
	return sign
}

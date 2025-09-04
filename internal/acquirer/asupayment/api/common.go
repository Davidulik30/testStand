package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	Pending    = "new"
	Reconciled = "success"
	Decline    = "cancelled"
)

type CardData struct {
	OwnerName    string `json:"owner_name"`
	CardNumber   string `json:"card_number"`
	ExpiredMonth string `json:"expired_month"`
	ExpiredYear  string `json:"expired_year"`
}

type Payload map[string]string

type Request struct {
	Merchant   string   `json:"merchant"`
	WithdrawID string   `json:"withdraw_id"`
	CardData   CardData `json:"card_data"`
	Amount     string   `json:"amount"`
	Signature  string   `json:"signature"`
	Payload    Payload  `json:"payload"`
}

type StatusRequest struct {
	MerchId string `json:"merchant_id"`
	ID      string `json:"id"`
	Sign    string `json:"sign,omitempty"`
}

type Response struct {
	ID          int    `json:"id"`
	Status      string `json:"status"`
	CurrID      int    `json:"currID"`
	Curr        string `json:"curr"`
	Amount      int64  `json:"amount"`
	Number      string `json:"number"`
	Info        string `json:"info"`
	BankTitle   string `json:"bankTitle"`
	BankID      string `json:"bankID"`
	Bank        string `json:"bank"`
	Page        string `json:"page"`
	Card        string `json:"card"`
	FIO         string `json:"fio"`
	Holder      string `json:"holder"`
	Error       string `json:"error"`
	Nspk        string `json:"nspk"`
	PaymentLink string `json:"paymentLink"`
}

type StatusResponse struct {
	Type       string `json:"type"`
	ID         int    `json:"ID"`
	CurrID     int    `json:"currID"`
	Amount     int64  `json:"amount"`
	Label      string `json:"label"`
	Memo       string `json:"memo"`
	Status     int    `json:"status"`
	StatusText string `json:"statusText"`
	Error      string `json:"error"`
}

type Callback struct {
	ShopID     int    `json:"shopID"`
	ID         int    `json:"id"`
	CurrID     int    `json:"currID"`
	Curr       string `json:"curr"`
	Amount     int64  `json:"amount"`
	Label      string `json:"label"`
	UserID     string `json:"userID"`
	Memo       string `json:"memo"`
	Info       string `json:"info,omitempty"`
	Status     int    `json:"status"`
	StatusText string `json:"statusText"`
	Error      string `json:"error"`
	Sign       string `json:"sign"`
}

func CalcSignature(merchant, cardOrPhone, amount, secretKey string) string {
	data := merchant + cardOrPhone + amount + secretKey
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func UnmarshalCallback(body string, cb *Callback) error {
	if err := json.Unmarshal([]byte(body), cb); err != nil {
		return fmt.Errorf("failed to unmarshal callback: %w", err)
	}
	return nil
}

package api

import (
	"crypto/sha256"
	"encoding/hex"
)

const (
	Pending    = "new"
	Reconciled = "executed"
	Decline    = "cancelled"
)

type Request struct {
	MerchId    string      `json:"merchant"`
	WithdrawId string      `json:"withdraw_id"`
	Data       CardDetails `json:"card_data"`
	Signature  string      `json:"signature"`
	Amount     string      `json:"amount"`
}

type Response struct {
	Status string `json:"status"`
	Id     string `json:"id"`
	// if error
	Detail string `json:"detail"`
	Code   string `json:"code"`
}

type CardDetails struct {
	CardNumber string `json:"card_number"`
}

type StatusRequest struct {
	Id string `json:"id"`
}

func createSign(request *Request, key string) string {
	hashString := request.MerchId + request.Data.CardNumber + request.Amount + key
	sum := sha256.Sum256([]byte(hashString))
	sign := hex.EncodeToString(sum[:])
	return sign
}

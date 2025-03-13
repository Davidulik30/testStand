package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"
)

type CardData struct {
	CardNumber string `json:"card_number"`
}

type Request struct {
	Merchant   string   `json:"merchant"`
	WithdrawId string   `json:"withdraw_id"`
	Amount     string   `json:"amount"`
	Signature  string   `json:"signature"`
	CardData   CardData `json:"card_data"`
}

type Response struct {
	Status string `json:"status"`
	Id     string `json:"id"`
	Error  string `json:"error"`
}

func GenerateSignature(request *Request, key string) string {
	sum := fmt.Sprintf("%s%s%s%s", request.Merchant, request.CardData.CardNumber, request.Amount, key)
	sha := sha256.Sum256([]byte(sum))
	return hex.EncodeToString(sha[:])
}

func CleanCardNumber(cardNumber string) string {
	var builder strings.Builder
	for _, char := range cardNumber {
		if unicode.IsDigit(char) {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

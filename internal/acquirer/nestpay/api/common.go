package api

import (
	"crypto/sha512"
	"encoding/base64"
	"strings"
)

type Request struct {
	XMLName       struct{} `xml:"CC5Request"`
	Name          string   `xml:"Name"`
	Password      string   `xml:"Password"`
	ClientId      string   `xml:"ClientId"`
	Type          string   `xml:"Type"`
	Number        string   `xml:"Number"`
	PayerAuthCode string   `xml:"PayerAuthenticationCode"`
	PayerSecLevel string   `xml:"PayerSecurityLevel"`
	PayerTxnId    string   `xml:"PayerTxnId"`
	Currency      string   `xml:"Currency"`
	OrderId       string   `xml:"OrderId"`
}

type Response struct {
	ErrMsg         string `xml:"ErrMsg"`
	ProcReturnCode string `xml:"ProcReturnCode"`
	Response       string `xml:"Response"`
	OrderId        string `xml:"OrderId"`
}

type Est3DGateRequest struct {
	ClientId                    string `url:"clientid"`
	StoreType                   string `url:"storetype"`
	TranType                    string `url:"trantype"`
	Currency                    string `url:"currency"`
	Oid                         string `url:"oid"`
	Rnd                         string `url:"rnd"`
	Pan                         string `url:"pan"`
	EcomPaymentCardExpDateMonth string `url:"Ecom_Payment_Card_ExpDate_Month"`
	EcomPaymentCardExpDateYear  string `url:"Ecom_Payment_Card_ExpDate_Year"`
	Cv2                         string `url:"cv2"`
	HashAlgorithm               string `url:"hashAlgorithm"`
	Hash                        string `url:"hash"`
	Encoding                    string `url:"encoding"`
}

func GenerateHash(req *Est3DGateRequest, storeKey string) string {
	params := []string{
		req.ClientId,                    // clientid
		req.Currency,                    // currency
		req.Cv2,                         // cv2
		req.EcomPaymentCardExpDateMonth, // Ecom_Payment_Card_ExpDate_Month
		req.EcomPaymentCardExpDateYear,  // Ecom_Payment_Card_ExpDate_Year
		req.HashAlgorithm,               // hashAlgorithm
		req.Oid,                         // oid
		req.Pan,                         // pan
		req.Rnd,                         // rnd
		req.StoreType,                   // storetype
		req.TranType,                    // trantype
	}
	data := strings.Join(params, "|") + "|" + storeKey
	hash := sha512.Sum512([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:])
}

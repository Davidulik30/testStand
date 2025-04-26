package api

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/fatih/structs"
	"sort"
	"strings"
)

const (
	StatusApproved = "Approve"
	StatusDeclined = "Declined"
	StatusError    = "Error"
)

type Request struct {
	ClientId                        string `form:"ClientId"`
	StoreType                       string `form:"storeType"`
	TranType                        string `form:"tranType"`
	Amount                          string `form:"amount"`
	Currency                        string `form:"Currency"`
	Lang                            string `form:"lang"`
	Oid                             string `form:"oid"`
	Rnd                             string `form:"rnd"`
	Pan                             string `form:"pan"`
	Ecom_Payment_Card_ExpDate_Month string `form:"Ecom_Payment_Card_ExpDate_Month"`
	Ecom_Payment_Card_ExpDate_Year  string `form:"Ecom_Payment_Card_ExpDate_Year"`
	Cv2                             string `form:"cv2"`
	HashAlgorithm                   string `form:"hashAlgorithm"`
	Encoding                        string `form:"encoding"`
	Hash                            string `form:"hash"`
}

type PaymentResponse struct {
	ErrMsg   string `form:"ErrMsg"`
	Amount   int    `form:"amount"`
	Cavv     string `form:"cavv"`
	ClientIP string `form:"clientIp"`
	ClientID string `form:"clientid"`
	Currency int    `form:"currency"`
	ECI      string `form:"eci"`
	MD       string `form:"md"`
	Oid      string `form:"oid"`
	XID      string `form:"xid"`
}

type StatusRequest struct {
	XMLName                 xml.Name `xml:"CC5Request"`
	ClientId                string   `xml:"ClientId"`
	Name                    string   `xml:"Name"`
	Password                string   `xml:"Password"`
	Currency                string   `xml:"Currency"`
	Oid                     string   `xml:"oid"`
	Type                    string   `xml:"Type"`
	Amount                  int      `xml:"amount"`
	Number                  string   `xml:"Number"`
	PayerAuthenticationCode string   `xml:"PayerAuthenticationCode"`
	PayerSecurityLevel      string   `xml:"PayerSecurityLevel"`
	PayerTxnId              string   `xml:"PayerTxnId"`
	IPAddress               string   `xml:"IPAddress"`
}

type StatusResponse struct {
	ErrMsg         string `xml:"ErrMsg"`
	ProcReturnCode string `xml:"ProcReturnCode"`
	Status         string `xml:"Response"`
	OrderId        string `xml:"OrderId"`
}

func setHash(request *Request, storeKey string) error {

	requestMap := structs.Map(request)

	var sortedKeys []string
	for key := range requestMap {
		lowerKey := strings.ToLower(key)
		if lowerKey != "hash" && lowerKey != "encoding" {
			sortedKeys = append(sortedKeys, key)
		}
	}

	sort.Strings(sortedKeys)

	var dataBuilder strings.Builder
	for key := range sortedKeys {
		dataBuilder.WriteString(fmt.Sprintf("%v", requestMap[sortedKeys[key]]))
		dataBuilder.WriteString("|")
	}

	dataBuilder.WriteString(storeKey)

	hashValue := dataBuilder.String()
	hash := sha512.Sum512([]byte(hashValue))
	request.Hash = base64.StdEncoding.EncodeToString(hash[:])

	return nil
}

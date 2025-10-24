package api

import "encoding/xml"

const (
	StatusCancelled = "Decline"
	StatusApproved  = "Approve"
	StatusError     = "Error"
)

type Request struct {
	ClientId                    string `json:"clientid"`
	OId                         string `json:"oid"`
	Amount                      string `json:"amount"`
	Currency                    string `json:"currency"`
	OkUrl                       string `json:"okurl"`
	FailUrl                     string `json:"failurl"`
	Installment                 string `json:"installment"`
	Rnd                         string `json:"rnd"`
	Hash                        string `json:"hash"`
	EcomPaymentCardExpDateMonth string `json:"Ecom_Payment_Card_ExpDate_Month,omitempty"`
	EcomPaymentCardExpDateYear  string `json:"Ecom_Payment_Card_ExpDate_Year,omitempty"`
	Cv2                         string `json:"cv2,omitempty"`
	TransactionType             string `json:"trantype"`
	StoreType                   string `json:"storetype"`
	HashAlgorithm               string `json:"hashAlgorithm"`
	Lang                        string `json:"lang"`
	Encoding                    string `json:"encoding"`
	Pan                         string `json:"pan,omitempty"`
}

type PaymentResponse struct {
	AcqBin               string `json:"acq_bin" form:"ACQBIN"`
	ExpMonth             string `json:"exp_month" form:"Ecom_Payment_Card_ExpDate_Month"`
	ErrMsg               string `json:"err_msg" form:"ErrMsg"`
	PAResVerified        string `json:"pares_verified" form:"PAResVerified"`
	ACSReferenceNumber   string `json:"acs_reference_number" form:"TDS2.acsReferenceNumber"`
	AuthTimestamp        string `json:"auth_timestamp" form:"TDS2.authTimestamp"`
	ThreeDSServerTransID string `json:"three_ds_server_trans_id" form:"TDS2.threeDSServerTransID"`
	TransStatus          string `json:"trans_status" form:"TDS2.transStatus"`
	ThreedID             string `json:"three_d_id" form:"THREED_ID"`
	TransID              string `json:"trans_id" form:"TransID"`
	Amount               string `json:"amount" form:"amount"`
	CAVV                 string `json:"cavv" form:"cavv"`
	ECI                  string `json:"eci" form:"eci"`
	MD                   string `json:"md" form:"md"`
	MDStatus             string `json:"md_status" form:"mdStatus"`
	OrderID              string `json:"order_id" form:"oid"`
	Currency             string `json:"currency" form:"currency"`
	ClientID             string `json:"client_id" form:"clientid"`
	MerchantID           string `json:"merchant_id" form:"merchantID"`
	StoreType            string `json:"store_type" form:"storetype"`
	TransactionStatus    string `json:"transaction_status" form:"paresTxStatus"`
	ThreeDSError         string `json:"three_d_error" form:"mdErrorMsg"`
	ClientIP             string `json:"client_ip" form:"clientIp"`
	XID                  string `json:"xid"`
}

type OrderStatus struct {
	XMLName                 xml.Name `xml:"CC5Request"`
	Name                    string   `xml:"Name"`
	Password                string   `xml:"Password"`
	ClientId                string   `xml:"ClientId"`
	Type                    string   `xml:"Type"`
	OrderId                 string   `xml:"OrderId"`
	Total                   string   `xml:"Total"`
	Currency                string   `xml:"Currency"`
	Number                  string   `xml:"Number"`
	PayerTxnId              string   `xml:"PayerTxnId"`
	PayerSecurityLevel      string   `xml:"PayerSecurityLevel"`
	PayerAuthenticationCode string   `xml:"PayerAuthenticationCode"`
}

type StatusResponse struct {
	ErrMsg  string `xml:"ErrMsg"`
	Status  string `xml:"Response"`
	OrderId string `xml:"OrderId"`
}

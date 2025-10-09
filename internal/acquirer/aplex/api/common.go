// Структура для запроса оплаты (payment)

package api

const (
	StatusCancelled = "DECLINED"
	StatusApproved  = "RELEASED"
	StatusPending   = "PENDING"
)

type Payload map[string]string

type UserLoad struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Request struct {
	FiatSymbol     string `json:"fiat_symbol"`
	FiatAmount     string `json:"fiat_amount"`
	CustomerName   string `json:"customer_name"`
	CustomerAdress string `json:"customer_address,omitempty"`
	Direction      string `json:"direction"`
	ExternalID     string `json:"external_id,omitempty"`
	GateId         string `json:"gate_id,omitempty"`
	WebHookUrl     string `json:"webhook_url,omitempty"`
}

type Response struct {
	ID          int    `json:"id"`
	Status      string `json:"status"`
	CurrID      int    `json:"currID"`
	Curr        string `json:"curr"`
	Amount      string `json:"amount"`
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
	Signatrue   string `json:"signature_key"`
}

type Callback struct {
	ID        string `json:"_id"`
	Status    string `json:"status"`
	Signature string `json:"signature"`
}

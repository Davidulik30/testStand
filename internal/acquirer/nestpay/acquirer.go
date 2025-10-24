package nestpay

import (
	"context"
	"strconv"

	"testStand/internal/acquirer"
	"testStand/internal/acquirer/helper"
	"testStand/internal/acquirer/nestpay/api"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/labstack/gommon/log"
)

const (
	StoreType       = "3d_pay"
	Lang            = "en"
	HashAlgorithm   = "ver3"
	Encoding        = "utf-8"
	TransactionType = "Auth"
	RandomString    = "str"
	Installment     = ""
	OkUrl           = "/ok"
	FailUrl         = "/fail"
)

type ChannelParams struct {
	ClientId    string `json:"clientId"`
	Currency    string `json:"currency"`
	StoreKey    string `json:"storekey"`
	ApiName     string `json:"apiname"`
	ApiPassword string `json:"apipassword"`
}

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}
type GatewayParams struct {
	Transport Transport `json:"transport"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams ChannelParams
	callbackUrl   string
}

func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.StoreKey),
		channelParams: *channelParams,
		dbClient:      db,
	}
}
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	logger := log.New("nestpay-payment")
	request := api.Request{
		ClientId:                    a.channelParams.ClientId,
		OId:                         strconv.FormatInt(txn.TxnId, 10),
		Amount:                      strconv.FormatInt(txn.TxnAmountSrc, 10),
		Currency:                    a.channelParams.Currency,
		OkUrl:                       a.callbackUrl + OkUrl,
		FailUrl:                     a.callbackUrl + OkUrl,
		Rnd:                         RandomString,
		TransactionType:             TransactionType,
		StoreType:                   StoreType,
		Lang:                        Lang,
		HashAlgorithm:               HashAlgorithm,
		Encoding:                    Encoding,
		Installment:                 Installment,
		Pan:                         txn.PaymentData.Object.Credentials,
		EcomPaymentCardExpDateYear:  txn.PaymentData.Object.ExpYear,
		EcomPaymentCardExpDateMonth: txn.PaymentData.Object.ExpMonth,
		Cv2:                         txn.PaymentData.Object.Cvv,
	}

	response, err := a.api.MakeTransaction(ctx, request)
	if err != nil {
		logger.Error("Payment error: ", err)
		return nil, err
	}
	logger.Infof("Payment response: %+v", response)

	statusRequest := api.OrderStatus{
		Name:                    a.channelParams.ApiName,
		Password:                a.channelParams.ApiPassword,
		ClientId:                response.ClientID,
		Type:                    TransactionType,
		OrderId:                 strconv.FormatInt(txn.TxnId, 10),
		Total:                   response.Amount,
		Currency:                response.Currency,
		Number:                  response.MD,
		PayerTxnId:              response.XID,
		PayerSecurityLevel:      response.ECI,
		PayerAuthenticationCode: response.CAVV,
	}
	statusResp, err := a.api.OrderStatusQuery(ctx, statusRequest)
	if err != nil {
		logger.Error("Payment error: ", err)
		return nil, err
	}

	if response.ErrMsg != "" {
		logger.Error("Payment rejected: ", response.ErrMsg)
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info:   map[string]string{"ps_error_code": response.ErrMsg},
		}, nil
	}

	return handleStatus(statusResp.Status)
}

func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

func handleStatus(status string) (*acquirer.TransactionStatus, error) {
	tr := &acquirer.TransactionStatus{}
	switch status {
	case api.StatusApproved:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case api.StatusCancelled:
		tr.Status = acquirer.REJECTED
		return tr, nil
	default:
		tr.Status = acquirer.PENDING
		return tr, nil
	}
}

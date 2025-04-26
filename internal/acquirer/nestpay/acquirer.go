package nestpay

import (
	"context"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/helper"
	"testStand/internal/acquirer/nestpay/api"
	"testStand/internal/models"
	"testStand/internal/repos"
)

type GatewayParams struct {
	Transport Transport `json:"transport"`
}

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type ChannelParams struct {
	ClientId string `json:"clientId"`
	Name     string `json:"name"`
	Password string `json:"password"`
	StoreKey string `json:"store_key"`
	Currency string `json:"currency"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
	callbackUrl   string
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams) *Acquirer {

	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.StoreKey),
		dbClient:      db,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	requestBody := &api.Request{
		ClientId:                        a.channelParams.ClientId,
		StoreType:                       "3d_pay",
		Amount:                          strconv.FormatInt(txn.TxnAmountSrc, 10),
		Currency:                        a.channelParams.Currency,
		TranType:                        "Auth",
		Lang:                            "tr",
		Rnd:                             "string",
		Pan:                             txn.PaymentData.Object.Credentials,
		Ecom_Payment_Card_ExpDate_Month: txn.PaymentData.Object.ExpMonth,
		Ecom_Payment_Card_ExpDate_Year:  txn.PaymentData.Object.ExpYear,
		Cv2:                             txn.PaymentData.Object.Cvv,
		Oid:                             strconv.FormatInt(txn.TxnId, 10),
		HashAlgorithm:                   "ver3",
		Encoding:                        "utf-8",
	}

	paymentResp, err := a.api.MakePayment(ctx, requestBody)
	if err != nil {
		return nil, err
	}

	if len(paymentResp.ErrMsg) > 0 {
		status := &acquirer.TransactionStatus{Status: acquirer.REJECTED}
		status.Info = map[string]string{
			"ps_error_message": helper.DecodeUnicode(paymentResp.ErrMsg),
		}
		return status, nil
	}

	statusBody := &api.StatusRequest{
		ClientId:                a.channelParams.ClientId,
		Name:                    a.channelParams.Name,
		Password:                a.channelParams.Password,
		Currency:                a.channelParams.Currency,
		Oid:                     paymentResp.Oid,
		Type:                    "Auth",
		Amount:                  paymentResp.Amount,
		Number:                  paymentResp.MD,
		PayerAuthenticationCode: paymentResp.Cavv,
		PayerSecurityLevel:      paymentResp.ECI,
		PayerTxnId:              paymentResp.XID,
		IPAddress:               paymentResp.ClientIP,
	}

	statusResp, err := a.api.CheckStatus(ctx, statusBody)
	if err != nil {
		return nil, err
	}

	switch statusResp.Status {
	case api.StatusApproved:
		return &acquirer.TransactionStatus{
			Status: acquirer.APPROVED,
		}, nil
	case api.StatusDeclined, api.StatusError:
		status := &acquirer.TransactionStatus{Status: acquirer.REJECTED}
		if len(statusResp.ErrMsg) > 0 {
			status.Info = map[string]string{
				"ps_error_message": helper.DecodeUnicode(statusResp.ErrMsg),
			}
		}
		return status, nil
	default:
		return &acquirer.TransactionStatus{
			Status: acquirer.PENDING,
		}, nil
	}
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

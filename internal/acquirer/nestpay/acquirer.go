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
	ApiName     string `json:"api_name"`
	ApiPassword string `json:"api_password"`
	Currency    string `json:"currency"`
	ClientId    string `json:"client_id"`
	StoreKey    string `json:"store_key"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
}

func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams) *Acquirer {
	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(gatewayParams.Transport.BaseAddress, channelParams.ClientId, gatewayParams.Transport.Timeout),
		dbClient:      db,
	}
}

func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	req := &api.Est3DGateRequest{
		ClientId:                    a.channelParams.ClientId,
		StoreType:                   "3d_pay",
		TranType:                    "Auth",
		Currency:                    a.channelParams.Currency,
		Oid:                         strconv.FormatInt(txn.TxnId, 10),
		Rnd:                         "asdf",
		Pan:                         txn.PaymentData.Object.Credentials,
		EcomPaymentCardExpDateMonth: txn.PaymentData.Object.ExpMonth,
		EcomPaymentCardExpDateYear:  txn.PaymentData.Object.ExpYear,
		Cv2:                         txn.PaymentData.Object.Cvv,
		HashAlgorithm:               "ver3",
		Encoding:                    "utf-8",
	}

	md, cavv, eci, xid, err := a.api.EstAuth(ctx, req, a.channelParams.StoreKey)
	if err != nil {
		return nil, err
	}

	apiReq := &api.Request{
		Name:          a.channelParams.ApiName,
		Password:      a.channelParams.ApiPassword,
		ClientId:      a.channelParams.ClientId,
		Type:          "Auth",
		Number:        md,
		PayerAuthCode: cavv,
		PayerSecLevel: eci,
		PayerTxnId:    xid,
		Currency:      a.channelParams.Currency,
		OrderId:       strconv.FormatInt(txn.TxnId, 10),
	}

	response, err := a.api.MakePayment(ctx, apiReq)
	if err != nil {
		return nil, err
	}

	if response.Response != "Approve" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_message": response.ErrMsg,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		Status:   acquirer.APPROVED,
		GtwTxnId: &response.OrderId,
	}

	return tr, nil
}

func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

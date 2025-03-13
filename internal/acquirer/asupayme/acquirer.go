package asupayme

import (
	"context"
	"fmt"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/asupayme/api"
	"testStand/internal/acquirer/helper"
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
	ApiKey    string `json:"api_key"`
	MerchId   string `json:"merchant_id"`
	SecretKey string `json:"secret_key"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams) *Acquirer {
	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(gatewayParams.Transport.BaseAddress, channelParams.ApiKey, gatewayParams.Transport.Timeout),
		dbClient:      db,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	request := &api.Request{
		Merchant:   a.channelParams.MerchId,
		WithdrawId: fmt.Sprintf("%d", txn.TxnId),
		Amount:     fmt.Sprintf("%d", txn.TxnAmountSrc),
		CardData: api.CardData{
			CardNumber: txn.PaymentData.Object.Credentials,
		},
	}
	request.Signature = api.GenerateSignature(request, a.channelParams.SecretKey)
	response, err := a.api.MakePayout(request)
	
	request.CardData.CardNumber = api.CleanCardNumber(request.CardData.CardNumber)

	if err != nil {
		return nil, err
	}

	if response.Status != "success" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		Status:   acquirer.APPROVED,
		GtwTxnId: &response.Id,
	}

	//return handleStatus(tr, response.Status)
	return tr, err
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

/*
func handleStatus(tr *acquirer.TransactionStatus, status string) (*acquirer.TransactionStatus, error) {
	switch status {
	case api.Reconciled:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case api.Decline:
		tr.Status = acquirer.REJECTED
		return tr, nil
	case api.Pending:
		fallthrough
	default:
		tr.Status = acquirer.PENDING
		return tr, nil
	}
}
*/

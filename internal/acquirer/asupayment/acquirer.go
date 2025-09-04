package asupayment

import (
	"context"
	"fmt"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/asupayment/api"
	"testStand/internal/models"
	"testStand/internal/repos"
)

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type ChannelParams struct {
	SecretKey string `json:"secret_key"`
	ApiKey    string `json:"api_key"`
	MerchId   string `json:"merchant_id"`
}

type GatewayParams struct {
	Transport Transport `json:"transport"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams ChannelParams
	gatewayParams GatewayParams
	callbackUrl   string
}

func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.SecretKey, channelParams.ApiKey, gatewayParams.Transport.Timeout),
		channelParams: *channelParams,
		dbClient:      db,
		gatewayParams: *gatewayParams,
		callbackUrl:   callbackUrl,
	}
}

func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return nil, fmt.Errorf("method Payment not implemented for asupayment")
}

func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	request := api.Request{
		Merchant:   a.channelParams.MerchId,
		WithdrawID: fmt.Sprintf("%d", txn.TxnId),
		CardData: api.CardData{
			OwnerName:    txn.Customer.FullName,
			CardNumber:   txn.PaymentData.Object.Credentials,
			ExpiredMonth: txn.PaymentData.Object.ExpMonth,
			ExpiredYear:  txn.PaymentData.Object.ExpYear,
		},
		Amount: fmt.Sprintf("%d", txn.TxnAmountSrc),
		Payload: api.Payload{
			"field1": txn.TxnInfo["field1"],
			"field2": txn.TxnInfo["field2"],
		},
	}

	response, err := a.api.MakeWithdraw(ctx, request)
	if err != nil {
		return nil, err
	}

	if response.Error != "" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info:   map[string]string{"ps_error_code": response.Error},
		}, nil
	}

	gtwTxnId := strconv.Itoa(response.ID)
	tr := &acquirer.TransactionStatus{
		GtwTxnId: &gtwTxnId,
	}
	return handleStatus(tr, response.Status)
}

func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return nil, fmt.Errorf("not implemented for asupayment")
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return nil, fmt.Errorf("not implemented for asupayment")
}

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

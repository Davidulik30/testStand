package alpex

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/labstack/gommon/log"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/alpex/api"
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
	Login      string `json:"login"`
	Password   string `json:"password"`
	WebhookURL string `json:"webhook_url"`
	GateId     string `json:"gate_id"`
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
		api:           api.NewClient(gatewayParams.Transport.BaseAddress, channelParams.Login, channelParams.Password, gatewayParams.Transport.Timeout),
		dbClient:      db,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	request := &api.Request{
		FiatSymbol:      txn.TxnCurrencySrc,
		FiatAmount:      strconv.FormatInt(txn.TxnAmountSrc, 10),
		CustomerName:    txn.Customer.FullName,
		CustomerAddress: txn.Customer.Address,
		Direction:       api.Payment,
		GateID:          a.channelParams.GateId,
		ExternalID:      strconv.FormatInt(txn.TxnId, 10),
		Webhook:         a.channelParams.WebhookURL,
	}
	response, err := a.api.MakePayInOut(request)
	if err != nil {
		return nil, err
	}
	if response.Status != "PENDING" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		Outputs: map[string]string{
			"credentials": response.PayMethod.Address,
			"bank":        response.PayMethod.Gate.Name,
			"description": response.PayMethod.Person,
		},
		GtwTxnId: &response.Id,
	}

	return tr, err
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	request := &api.Request{
		FiatSymbol:      txn.TxnCurrencySrc,
		FiatAmount:      strconv.FormatInt(txn.TxnAmountSrc, 10),
		CustomerName:    txn.Customer.FullName,
		CustomerAddress: txn.Customer.Address,
		Direction:       api.Payout,
		GateID:          a.channelParams.GateId,
		ExternalID:      strconv.FormatInt(txn.TxnId, 10),
		Webhook:         a.channelParams.WebhookURL,
	}
	response, err := a.api.MakePayInOut(request)
	if err != nil {
		return nil, err
	}
	if response.Status != "PENDING" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		Status:   acquirer.PENDING,
		GtwTxnId: &response.Id,
	}

	return tr, err
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	logger := log.New("dev")

	callbackBody, ok := txn.TxnInfo["callback"]

	if !ok {
		return nil, errors.New("callback body is missing")
	}

	callback := api.Callback{}
	err := json.Unmarshal([]byte(callbackBody), &callback)
	if err != nil {
		logger.Error("Error unmarshalling callback body - ", callbackBody)
		return nil, err
	}

	ok, err = a.api.CheckCallBack(&callback)
	if err != nil {
		return nil, err
	}
	tr := &acquirer.TransactionStatus{}
	if callback.Status == api.StatusReleased {
		tr.Status = acquirer.APPROVED
	} else if callback.Status == api.StatusDeclined {
		tr.Status = acquirer.REJECTED
	}
	return tr, err
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

/*
func handleStatus(tr *acquirer.TransactionStatus, status string) (*acquirer.TransactionStatus, error) {
	switch status {
	case api.StatusReleased:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case api.StatusDeclined:
		tr.Status = acquirer.REJECTED
		return tr, nil
	case api.StatusPending:
		fallthrough
	default:
		tr.Status = acquirer.PENDING
		return tr, nil
	}
}
*/

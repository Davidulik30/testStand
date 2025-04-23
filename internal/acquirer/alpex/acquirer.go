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

const (
	DirectionSell = "SELL"
	DirectionBuy  = "BUY"
)

type GatewayParams struct {
	Transport Transport `json:"transport"`
}

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type ChannelParams struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	SecretKey string `json:"secret_key"`
	GateId    string `json:"gate_id"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
	callbackUrl   string
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {

	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.Email, channelParams.Password),
		dbClient:      db,
		callbackUrl:   callbackUrl,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	requestBody := &api.Request{
		Id:              strconv.FormatInt(txn.TxnId, 10),
		Symbol:          txn.TxnCurrencySrc,
		Amount:          txn.TxnAmountSrc,
		Direction:       DirectionBuy,
		CustomerName:    txn.Customer.FullName,
		CustomerAddress: txn.PaymentData.Object.Credentials,
		WebhookUrl:      "https://webhook.site/2e121990-aeaf-4680-98c2-babc293fbb11",
	}

	err := a.api.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	response, err := a.api.MakePayment(ctx, requestBody)
	if err != nil {
		return nil, err
	}

	if response.Status != "PENDING" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.ApproveCode,
			},
		}, nil
	}

	outputs := map[string]string{"credentials": response.PaymentMethod.Address}
	if len(response.PaymentMethod.Gate.BankName) != 0 {
		outputs["bank"] = response.PaymentMethod.Gate.BankName
	}
	if len(response.PaymentMethod.Person) != 0 {
		outputs["description"] = response.PaymentMethod.Person
	}

	return &acquirer.TransactionStatus{
		Status:   acquirer.PENDING,
		GtwTxnId: &response.Id,
	}, nil
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	fullName := txn.Customer.FullName
	if len(fullName) == 0 {
		return nil, errors.New("customer's fullName is required")
	}

	address := txn.PaymentData.Object.Credentials
	if len(address) == 0 {
		return nil, errors.New("customer's address is required")
	}

	requestData := &api.Request{
		Id:              strconv.FormatInt(txn.TxnId, 10),
		Symbol:          txn.TxnCurrencySrc,
		Amount:          txn.TxnAmountSrc,
		CustomerName:    fullName,
		CustomerAddress: address,
		Direction:       DirectionSell,
		WebhookUrl:      "https://webhook.site/2e121990-aeaf-4680-98c2-babc293fbb11",
		GateId:          a.channelParams.GateId,
	}

	err := a.api.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	response, err := a.api.MakePayout(ctx, requestData)
	if err != nil {
		return nil, err
	}

	if response.Status != "PENDING" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.ApproveCode,
			},
		}, nil
	}

	return &acquirer.TransactionStatus{
		Status:   acquirer.PENDING,
		GtwTxnId: &response.Id,
	}, nil
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

	signatureKey, err := a.api.GetSignatureKey(ctx)
	if err != nil {
		return nil, err
	}
	neededCallbackSign := api.CreateSign(callback.Id, callback.Status, signatureKey.SignKey)

	if callback.Sign != neededCallbackSign {
		return nil, errors.New("invalid Callback")
	}
	return handleStatus(callback.Status)
}

func handleStatus(status string) (*acquirer.TransactionStatus, error) {
	tr := &acquirer.TransactionStatus{}
	switch status {
	case api.Released:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case api.Declined:
		tr.Status = acquirer.REJECTED
		return tr, nil
	case api.Pending:
		fallthrough
	default:
		tr.Status = acquirer.PENDING
		return tr, nil
	}
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

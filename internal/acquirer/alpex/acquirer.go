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
	Id        string `json:"id"`
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
		ExternalId:      strconv.FormatInt(txn.TxnId, 10),
		Id:              strconv.FormatInt(txn.TxnId, 10),
		Symbol:          txn.TxnCurrencySrc,
		Amount:          txn.TxnAmountSrc,
		Status:          txn.TxnStatusId,
		Direction:       DirectionBuy,
		CustomerName:    txn.Customer.FullName,
		CustomerAddress: txn.Customer.Address,
		Credentials:     txn.PaymentData.Object.Credentials,
		WebhookUrl:      "https://webhook.site/2e121990-aeaf-4680-98c2-babc293fbb11",
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
		Status:  acquirer.PENDING,
		Outputs: outputs,
	}, nil
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	requestData, err := a.fillPayoutRequest(ctx, txn)
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

	tr := &acquirer.TransactionStatus{
		Status:   acquirer.PENDING,
		GtwTxnId: &response.Id,
	}

	return tr, nil
}

func (a *Acquirer) fillPayoutRequest(ctx context.Context, txn *models.Transaction) (*api.Request, error) {

	fullName := txn.Customer.FullName
	if len(fullName) == 0 {
		return nil, errors.New("customer's fullName is required")
	}

	address := txn.Customer.Address
	if len(address) == 0 {
		return nil, errors.New("customer's address is required")
	}

	request := &api.Request{
		ExternalId:      strconv.FormatInt(txn.TxnId, 10),
		Id:              strconv.FormatInt(txn.TxnId, 10),
		Status:          txn.TxnStatusId,
		Symbol:          txn.TxnCurrencySrc,
		Amount:          txn.TxnAmountSrc,
		CustomerName:    fullName,
		CustomerAddress: address,
		Credentials:     txn.PaymentData.Object.Credentials,
		Direction:       DirectionSell,
		WebhookUrl:      "https://webhook.site/2e121990-aeaf-4680-98c2-babc293fbb11",
		GateId:          a.channelParams.Id,
	}

	return request, nil
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
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

	tr := &acquirer.TransactionStatus{}

	signatureKey, err := a.api.GetSignatureKey(ctx)
	if err != nil {
		return nil, err
	}
	neededCallbackSign := api.CreateSign(callback.Id, callback.Status, signatureKey.SignKey)

	if callback.Sign != neededCallbackSign {
		logger.Error("Invalid Callback")
		return tr, nil
	}
	return handleStatus(tr, callback.Status)
}

func handleStatus(tr *acquirer.TransactionStatus, status string) (*acquirer.TransactionStatus, error) {
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

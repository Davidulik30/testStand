package aplex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/aplex/api"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/labstack/gommon/log"
)

const (
	DirectionBuy  = "BUY"
	DirectionSell = "SELL"
	Webhooklink   = "https://webhook.site/1c568273-747f-4db2-aae8-6a72789ebf8f"
)

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type ChannelParams struct {
	SecretKey string `json:"secret_key"`
	ApiKey    string `json:"api_key"`
}

type GatewayParams struct {
	Transport      Transport       `json:"transport"`
	PaymentMethods []PaymentMethod `json:"payment_methods"`
}

type PaymentMethod struct {
	Id    string  `json:"id"`
	GtwId *string `json:"gtw_id"`
}

type Acquirer struct {
	api            *api.Client
	dbClient       *repos.Repo
	channelParams  ChannelParams
	gatewayParams  GatewayParams
	paymentMethods []PaymentMethod
	callbackUrl    string
}

func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		api:            api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.SecretKey, channelParams.ApiKey, gatewayParams.Transport.Timeout),
		channelParams:  *channelParams,
		dbClient:       db,
		paymentMethods: gatewayParams.PaymentMethods,
		gatewayParams:  *gatewayParams,
		callbackUrl:    callbackUrl,
	}
}

func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	logger := log.New("aplex-payment")

	request := api.Request{
		FiatSymbol:   txn.TxnCurrencySrc,
		FiatAmount:   strconv.FormatInt(txn.TxnAmountSrc, 10),
		CustomerName: txn.Customer.FullName,
		Direction:    DirectionBuy,
		WebHookUrl:   Webhooklink,
		ExternalID:   strconv.FormatInt(txn.TxnId, 10),
	}

	response, err := a.api.MakeTransaction(ctx, request)
	if err != nil {
		logger.Error("Payment error: ", err)
		return nil, err
	}
	logger.Infof("Payment response: %+v", response)

	if response.Error != "" {
		logger.Error("Payment rejected: ", response.Error)
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info:   map[string]string{"ps_error_code": response.Error},
		}, nil
	}

	gtwTxnId := strconv.Itoa(response.ID)
	tr := &acquirer.TransactionStatus{
		GtwTxnId: &gtwTxnId,
	}
	logger.Infof("Payment status: %s, gtwTxnId: %s", response.Status, gtwTxnId)

	return handleStatus(tr, response.Status)
}

func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	logger := log.New("aplex-payout")

	// Проверка обязательных параметров
	if txn.Customer.FullName == "" {
		logger.Error("Payout error: отсутствует customer name")
		return nil, errors.New("customer name is required")
	}
	if txn.Customer.Address == "" {
		logger.Error("Payout error: отсутствует customer adress")
		return nil, errors.New("customer adress is required")
	}

	request := api.Request{
		FiatSymbol:     txn.TxnCurrencySrc,
		FiatAmount:     strconv.FormatInt(txn.TxnAmountSrc, 10),
		CustomerName:   txn.Customer.FullName,
		CustomerAdress: txn.Customer.Address,
		Direction:      DirectionSell,
		GateId:         *a.gatewayParams.PaymentMethods[0].GtwId,
		ExternalID:     strconv.FormatInt(txn.TxnId, 10),
		WebHookUrl:     Webhooklink,
	}
	logger.Infof("Payout request: %+v", request)

	response, err := a.api.MakeTransaction(ctx, request)
	if err != nil {
		logger.Error("Payout error: ", err)
		return nil, err
	}
	logger.Infof("Payout response: %+v", response)

	if response.Error != "" {
		logger.Error("Payout rejected: ", response.Error)
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info:   map[string]string{"ps_error_code": response.Error},
		}, nil
	}

	gtwTxnId := strconv.Itoa(response.ID)
	tr := &acquirer.TransactionStatus{
		GtwTxnId: &gtwTxnId,
	}
	logger.Infof("Payout status: %s, gtwTxnId: %s", response.Status, gtwTxnId)

	return handleStatus(tr, response.Status)
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

	txnStatus := acquirer.TransactionStatus{}

	if txn.GtwTxnId == nil && callback.ID != "" {
		txnStatus.GtwTxnId = &callback.ID
	}

	isValid, err := api.ValidateSignature(callback.ID, callback.Status, callback.Signature, a.channelParams.SecretKey)
	if err != nil {
		logger.Error("error validating signature: ", err)
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info:   map[string]string{"ps_error_code": err.Error()},
		}, nil
	}
	if !isValid {
		logger.Info("invalid signature")
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info:   map[string]string{"ps_error_code": "invalid signature"},
		}, nil
	}

	return handleStatus(&txnStatus, callback.Status)
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return nil, fmt.Errorf("not implemented for asupayment")
}

func handleStatus(tr *acquirer.TransactionStatus, status string) (*acquirer.TransactionStatus, error) {
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

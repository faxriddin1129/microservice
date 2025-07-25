package octo

import (
	"github.com/gin-gonic/gin"
	"gofax-billing/internal/constants"
	"gofax-billing/internal/models"
	"gofax-billing/pkg/env"
	"gofax-billing/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GenerateShopApiLinkByCard(transaction *models.Transaction) (interface{}, int, string) {

	generateData := map[string]interface{}{
		"octo_shop_id":        OCTO_SHOP,
		"octo_secret":         OCOT_SECRET,
		"shop_transaction_id": "billing_" + strconv.Itoa(int(transaction.ID)),
		"auto_capture":        true,
		"test":                OCTO_TEST_MODE,
		"init_time":           time.Now().Format("2006-01-02 15:04:05"),
		"total_sum":           transaction.Amount,
		"currency":            transaction.Currency,
		"tag":                 nil,
		"description":         "Оплата заказа Asialuxe № " + transaction.OrderId,
		"return_url":          transaction.ReturnUrl,
		"language":            "ru",
		"notify_url":          env.GetEnv("BASE_URL") + "/api/octo/notify",
		"user_data": map[string]interface{}{
			"user_id": transaction.UserId,
			"phone":   transaction.Phone,
			"email":   transaction.Email,
		},
		"basket": []interface{}{
			map[string]interface{}{
				"position_desc": "Asialuxe Product " + transaction.ProductId,
				"price":         transaction.Amount,
				"count":         1,
			},
		},
		"payment_methods": []interface{}{
			map[string]string{
				"method": transaction.CardType,
			},
		},
	}
	remote, _ := getUUID(generateData)

	if remote.Error != 0 {
		return map[string]interface{}{
			"ID":     0,
			"Link":   nil,
			"Method": nil,
		}, http.StatusBadRequest, remote.ErrMessage
	}

	getLinkData := map[string]interface{}{
		"pan":            transaction.CardNumber,
		"exp":            transaction.CardExpire,
		"cardHolderName": transaction.Phone,
		"cvc2":           transaction.CardCvv,
		"email":          transaction.Email,
		"method":         transaction.CardType,
	}

	remoteLink, _ := getLink(getLinkData, remote.OctoPaymentUUID)

	if remoteLink.Error != 0 {
		return map[string]interface{}{
			"ID":     0,
			"Link":   nil,
			"Method": nil,
		}, http.StatusBadRequest, remoteLink.ErrMessage
	}

	return map[string]interface{}{
		"ID":     transaction.ID,
		"Link":   remoteLink.Data.RedirectURL,
		"Method": "GET",
	}, http.StatusOK, "FastPay successful"
}

func GenerateShopApiLink(transaction *models.Transaction) (interface{}, int, string) {

	generateData := map[string]interface{}{
		"octo_shop_id":        OCTO_SHOP,
		"octo_secret":         OCOT_SECRET,
		"shop_transaction_id": "billing_" + strconv.Itoa(int(transaction.ID)),
		"auto_capture":        true,
		"test":                OCTO_TEST_MODE,
		"init_time":           time.Now().Format("2006-01-02 15:04:05"),
		"total_sum":           transaction.Amount,
		"currency":            transaction.Currency,
		"tag":                 nil,
		"description":         "Оплата заказа Asialuxe № " + transaction.OrderId,
		"return_url":          transaction.ReturnUrl,
		"language":            "ru",
		"notify_url":          env.GetEnv("BASE_URL") + "/api/octo/notify",
		"user_data": map[string]interface{}{
			"user_id": transaction.UserId,
			"phone":   transaction.Phone,
			"email":   transaction.Email,
		},
		"basket": []interface{}{
			map[string]interface{}{
				"position_desc": "Asialuxe Product " + transaction.ProductId,
				"price":         transaction.Amount,
				"count":         1,
			},
		},
		"payment_methods": []interface{}{
			map[string]string{
				"method": "humo",
			},
			map[string]string{
				"method": "uzcard",
			},
			map[string]string{
				"method": "bank_card",
			},
		},
	}
	remote, _ := getUUID(generateData)

	if remote.Error != 0 {
		return map[string]interface{}{
			"ID":     0,
			"Link":   nil,
			"Method": nil,
		}, http.StatusBadRequest, remote.ErrMessage
	}

	return map[string]interface{}{
		"ID":     transaction.ID,
		"Link":   remote.Data.OctoPayURL,
		"Method": "GET",
	}, http.StatusOK, "FastPay successful"
}

func NotifyShopApi(form *OctoNotifyResponse, c *gin.Context) {

	trId := strings.TrimPrefix(form.ShopTransactionID, "billing_")
	trIdInt, _ := strconv.ParseInt(trId, 10, 64)

	transaction := models.TransactionGetById(trIdInt)
	if transaction.ID == 0 {
		utils.RespondJson(c, nil, http.StatusNotFound, "Transaction not found")
		return
	}

	if form.Status == "succeeded" {
		transaction.Status = constants.STATUS_SUCCESS
	}
	if form.Status == "canceled" {
		transaction.Status = constants.STATUS_CANCEL
	}
	if form.Status == "wait_user_action" || form.Status == "created" || form.Status == "waiting_for_capture" {
		transaction.Status = constants.STATUS_PENDING
	}
	_, err := models.TransactionUpdate(&transaction)
	if err != nil {
		utils.RespondJson(c, nil, http.StatusInternalServerError, "Internal server error. Transaction failed save")
		return
	}

	utils.RespondJson(c, nil, http.StatusOK, "Success fully")
}

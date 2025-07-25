package requests

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gofax-billing/internal/constants"
	"gofax-billing/internal/models"
	"gofax-billing/pkg/octo"
	"gofax-billing/pkg/utils"
	"net/http"
	"time"
)

type FastPayByCardForm struct {
	UserId     uint    `json:"UserId" validate:"required,gt=0"`
	Amount     float64 `json:"Amount" validate:"required,gt=0"`
	Provider   string  `json:"Provider" validate:"required,provider"`
	Currency   string  `json:"Currency" validate:"required,currency"`
	OrderId    string  `json:"OrderId" validate:"required,gt=0"`
	ProductId  string  `json:"ProductId" validate:"required,gt=0"`
	ReturnUrl  string  `json:"ReturnUrl" validate:"url"`
	Email      string  `json:"Email" validate:"required,email"`
	Phone      string  `json:"Phone"`
	CardNumber string  `json:"CardNumber" validate:"required"`
	CardExpire string  `json:"CardExpire" validate:"required"`
	CardCvv    string  `json:"CardCvv"`
	CardType   string  `json:"CardType" validate:"required"`
	Platform   string  `json:"Platform" validate:"required,gt=0"`
}

func FastPayByCardValidate(c *gin.Context) {
	var form FastPayByCardForm

	if err := c.ShouldBindJSON(&form); err != nil {
		utils.RespondJson(c, nil, http.StatusBadRequest, err.Error())
		return
	}

	err := validate.Struct(form)
	msg := ""
	if err != nil {
		errorMessage := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			errMsg := fmt.Sprintf("%s %s", err.Field(), err.Tag())
			errorMessage[err.Field()] = errMsg
			if msg == "" {
				msg = errMsg
			}
		}
		utils.RespondJson(c, errorMessage, http.StatusBadRequest, msg)
		return
	}

	transaction := models.Transaction{
		Type:        constants.TYPE_FAST_PAY,
		Status:      constants.STATUS_PENDING,
		Currency:    form.Currency,
		Provider:    form.Provider,
		Amount:      form.Amount,
		State:       0,
		Reason:      0,
		UUID:        uuid.New().String(),
		CreateTime:  time.Now().Unix(),
		PerformTime: time.Now().Unix(),
		OrderId:     form.OrderId,
		ReturnUrl:   form.ReturnUrl,
		ProductId:   form.ProductId,
		Email:       form.Email,
		Phone:       form.Phone,
		UserId:      form.UserId,
		CardNumber:  form.CardNumber,
		CardExpire:  form.CardExpire,
		CardCvv:     form.CardCvv,
		CardType:    form.CardType,
		Platform:    form.Platform,
	}

	err = utils.DB.Create(&transaction).Error
	if err != nil {
		utils.RespondJson(c, nil, http.StatusInternalServerError, "Internal server error. Transaction failed save")
		return
	}

	if transaction.Provider == constants.ProviderOcto {
		data, code, msg := octo.GenerateShopApiLinkByCard(&transaction)
		utils.RespondJson(c, data, code, msg)
		return
	}

	utils.RespondJson(c, nil, http.StatusNotFound, "Provider not found or inactive")
	return
}

package router

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func hasRoute(engine *gin.Engine, method string, path string) bool {
	for _, route := range engine.Routes() {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}

func withStripeConfig(t *testing.T, apiSecret string, webhookSecret string, priceID string) {
	t.Helper()

	oldAPISecret := setting.StripeApiSecret
	oldWebhookSecret := setting.StripeWebhookSecret
	oldPriceID := setting.StripePriceId

	setting.StripeApiSecret = apiSecret
	setting.StripeWebhookSecret = webhookSecret
	setting.StripePriceId = priceID

	t.Cleanup(func() {
		setting.StripeApiSecret = oldAPISecret
		setting.StripeWebhookSecret = oldWebhookSecret
		setting.StripePriceId = oldPriceID
	})
}

func withCreemConfig(t *testing.T, apiKey string, products string, webhookSecret string, testMode bool) {
	t.Helper()

	oldAPIKey := setting.CreemApiKey
	oldProducts := setting.CreemProducts
	oldWebhookSecret := setting.CreemWebhookSecret
	oldTestMode := setting.CreemTestMode

	setting.CreemApiKey = apiKey
	setting.CreemProducts = products
	setting.CreemWebhookSecret = webhookSecret
	setting.CreemTestMode = testMode

	t.Cleanup(func() {
		setting.CreemApiKey = oldAPIKey
		setting.CreemProducts = oldProducts
		setting.CreemWebhookSecret = oldWebhookSecret
		setting.CreemTestMode = oldTestMode
	})
}

func withEpayConfig(t *testing.T, payAddress string, epayID string, epayKey string) {
	t.Helper()

	oldPayAddress := operation_setting.PayAddress
	oldEpayID := operation_setting.EpayId
	oldEpayKey := operation_setting.EpayKey

	operation_setting.PayAddress = payAddress
	operation_setting.EpayId = epayID
	operation_setting.EpayKey = epayKey

	t.Cleanup(func() {
		operation_setting.PayAddress = oldPayAddress
		operation_setting.EpayId = oldEpayID
		operation_setting.EpayKey = oldEpayKey
	})
}

func TestSetApiRouterOmitsStripeRoutesWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withStripeConfig(t, "", "", "")

	engine := gin.New()
	SetApiRouter(engine)

	assert.False(t, hasRoute(engine, http.MethodPost, "/api/stripe/webhook"))
	assert.False(t, hasRoute(engine, http.MethodPost, "/api/user/stripe/pay"))
	assert.False(t, hasRoute(engine, http.MethodPost, "/api/user/stripe/amount"))
}

func TestSetApiRouterMountsWebhookWithoutTopupPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withStripeConfig(t, "sk_test_123", "whsec_test_123", "")

	engine := gin.New()
	SetApiRouter(engine)

	assert.True(t, hasRoute(engine, http.MethodPost, "/api/stripe/webhook"))
	assert.False(t, hasRoute(engine, http.MethodPost, "/api/user/stripe/pay"))
	assert.False(t, hasRoute(engine, http.MethodPost, "/api/user/stripe/amount"))
}

func TestSetApiRouterMountsStripeRoutesWhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withStripeConfig(t, "sk_test_123", "whsec_test_123", "price_test_123")

	engine := gin.New()
	SetApiRouter(engine)

	assert.True(t, hasRoute(engine, http.MethodPost, "/api/stripe/webhook"))
	assert.True(t, hasRoute(engine, http.MethodPost, "/api/user/stripe/pay"))
	assert.True(t, hasRoute(engine, http.MethodPost, "/api/user/stripe/amount"))
}

func TestSetApiRouterOmitsCreemRoutesWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withCreemConfig(t, "", "[]", "", false)

	engine := gin.New()
	SetApiRouter(engine)

	assert.False(t, hasRoute(engine, http.MethodPost, "/api/creem/webhook"))
	assert.False(t, hasRoute(engine, http.MethodPost, "/api/user/creem/pay"))
	assert.False(t, hasRoute(engine, http.MethodPost, "/api/subscription/creem/pay"))
}

func TestSetApiRouterMountsCreemRoutesWhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withCreemConfig(t, "creem_key", `[{"productId":"prod_1"}]`, "creem_whsec", false)

	engine := gin.New()
	SetApiRouter(engine)

	assert.True(t, hasRoute(engine, http.MethodPost, "/api/creem/webhook"))
	assert.True(t, hasRoute(engine, http.MethodPost, "/api/user/creem/pay"))
	assert.True(t, hasRoute(engine, http.MethodPost, "/api/subscription/creem/pay"))
}

func TestSetApiRouterOmitsEpayRoutesWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withEpayConfig(t, "", "", "")

	engine := gin.New()
	SetApiRouter(engine)

	assert.False(t, hasRoute(engine, http.MethodPost, "/api/user/epay/notify"))
	assert.False(t, hasRoute(engine, http.MethodGet, "/api/user/epay/notify"))
	assert.False(t, hasRoute(engine, http.MethodPost, "/api/user/pay"))
	assert.False(t, hasRoute(engine, http.MethodPost, "/api/user/amount"))
	assert.False(t, hasRoute(engine, http.MethodPost, "/api/subscription/epay/notify"))
	assert.False(t, hasRoute(engine, http.MethodGet, "/api/subscription/epay/notify"))
	assert.False(t, hasRoute(engine, http.MethodPost, "/api/subscription/epay/pay"))
	assert.False(t, hasRoute(engine, http.MethodGet, "/api/subscription/epay/return"))
	assert.False(t, hasRoute(engine, http.MethodPost, "/api/subscription/epay/return"))
}

func TestSetApiRouterMountsEpayRoutesWhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withEpayConfig(t, "https://epay.example.com", "partner_1", "key_1")

	engine := gin.New()
	SetApiRouter(engine)

	assert.True(t, hasRoute(engine, http.MethodPost, "/api/user/epay/notify"))
	assert.True(t, hasRoute(engine, http.MethodGet, "/api/user/epay/notify"))
	assert.True(t, hasRoute(engine, http.MethodPost, "/api/user/pay"))
	assert.True(t, hasRoute(engine, http.MethodPost, "/api/user/amount"))
	assert.True(t, hasRoute(engine, http.MethodPost, "/api/subscription/epay/notify"))
	assert.True(t, hasRoute(engine, http.MethodGet, "/api/subscription/epay/notify"))
	assert.True(t, hasRoute(engine, http.MethodPost, "/api/subscription/epay/pay"))
	assert.True(t, hasRoute(engine, http.MethodGet, "/api/subscription/epay/return"))
	assert.True(t, hasRoute(engine, http.MethodPost, "/api/subscription/epay/return"))
}

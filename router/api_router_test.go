package router

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/setting"
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

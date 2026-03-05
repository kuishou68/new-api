package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func init() {
	Register("google", &GoogleProvider{})
}

// GoogleProvider implements OAuth for Google
type GoogleProvider struct{}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	IDToken     string `json:"id_token"`
}

type googleUser struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func (p *GoogleProvider) GetName() string {
	return "Google"
}

func (p *GoogleProvider) IsEnabled() bool {
	return common.GoogleOAuthEnabled
}

func (p *GoogleProvider) ExchangeToken(ctx context.Context, code string, c *gin.Context) (*OAuthToken, error) {
	if code == "" {
		return nil, NewOAuthError(i18n.MsgOAuthInvalidCode, nil)
	}

	logger.LogDebug(ctx, "[OAuth-Google] ExchangeToken: code=%s...", code[:min(len(code), 10)])

	redirectURI := fmt.Sprintf("%s/oauth/google", c.Request.Header.Get("Origin"))
	if redirectURI == "/oauth/google" {
		redirectURI = fmt.Sprintf("%s/oauth/google", c.Request.Header.Get("Referer"))
		if strings.Contains(redirectURI, "?") {
			redirectURI = redirectURI[:strings.Index(redirectURI, "?")]
		}
	}

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", common.GoogleClientId)
	data.Set("client_secret", common.GoogleClientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{
		Timeout: 20 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Google] ExchangeToken error: %s", err.Error()))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "Google"}, err.Error())
	}
	defer res.Body.Close()

	logger.LogDebug(ctx, "[OAuth-Google] ExchangeToken response status: %d", res.StatusCode)

	var tokenResponse googleTokenResponse
	err = common.DecodeJson(res.Body, &tokenResponse)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Google] ExchangeToken decode error: %s", err.Error()))
		return nil, err
	}

	if tokenResponse.AccessToken == "" {
		logger.LogError(ctx, "[OAuth-Google] ExchangeToken failed: empty access token")
		return nil, NewOAuthError(i18n.MsgOAuthTokenFailed, map[string]any{"Provider": "Google"})
	}

	logger.LogDebug(ctx, "[OAuth-Google] ExchangeToken success: scope=%s", tokenResponse.Scope)

	return &OAuthToken{
		AccessToken: tokenResponse.AccessToken,
		TokenType:   tokenResponse.TokenType,
		Scope:       tokenResponse.Scope,
	}, nil
}

func (p *GoogleProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUser, error) {
	logger.LogDebug(ctx, "[OAuth-Google] GetUserInfo: fetching user info")

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	client := http.Client{
		Timeout: 20 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Google] GetUserInfo error: %s", err.Error()))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "Google"}, err.Error())
	}
	defer res.Body.Close()

	logger.LogDebug(ctx, "[OAuth-Google] GetUserInfo response status: %d", res.StatusCode)

	if res.StatusCode != http.StatusOK {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Google] GetUserInfo failed: status=%d", res.StatusCode))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthGetUserErr, map[string]any{"Provider": "Google"}, fmt.Sprintf("status %d", res.StatusCode))
	}

	var gUser googleUser
	err = common.DecodeJson(res.Body, &gUser)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Google] GetUserInfo decode error: %s", err.Error()))
		return nil, err
	}

	if gUser.Id == "" {
		logger.LogError(ctx, "[OAuth-Google] GetUserInfo failed: empty id")
		return nil, NewOAuthError(i18n.MsgOAuthUserInfoEmpty, map[string]any{"Provider": "Google"})
	}

	logger.LogDebug(ctx, "[OAuth-Google] GetUserInfo success: id=%s, name=%s, email=%s",
		gUser.Id, gUser.Name, gUser.Email)

	return &OAuthUser{
		ProviderUserID: gUser.Id,
		Username:       gUser.GivenName,
		DisplayName:    gUser.Name,
		Email:          gUser.Email,
	}, nil
}

func (p *GoogleProvider) IsUserIDTaken(providerUserID string) bool {
	return model.IsGoogleIdAlreadyTaken(providerUserID)
}

func (p *GoogleProvider) FillUserByProviderID(user *model.User, providerUserID string) error {
	user.GoogleId = providerUserID
	return user.FillUserByGoogleId()
}

func (p *GoogleProvider) SetProviderUserID(user *model.User, providerUserID string) {
	user.GoogleId = providerUserID
}

func (p *GoogleProvider) GetProviderPrefix() string {
	return "google_"
}

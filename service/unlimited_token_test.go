package service

import (
	"context"
	"testing"

	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedUnlimitedToken(t *testing.T, id int, userId int, key string, remainQuota int) {
	t.Helper()
	seedToken(t, id, userId, key, remainQuota)
	require.NoError(t, model.DB.Model(&model.Token{}).Where("id = ?", id).Update("unlimited_quota", true).Error)
}

func TestPreConsumeTokenQuota_UnlimitedTokenSkipsMutation(t *testing.T) {
	truncate(t)

	const userID, tokenID = 101, 101
	const tokenRemain = 9000

	seedUser(t, userID, 5000)
	seedUnlimitedToken(t, tokenID, userID, "sk-unlimited-pre", tokenRemain)

	info := &relaycommon.RelayInfo{
		UserId:         userID,
		TokenId:        tokenID,
		TokenKey:       "sk-unlimited-pre",
		TokenUnlimited: true,
	}

	require.NoError(t, PreConsumeTokenQuota(info, 3000))
	assert.Equal(t, tokenRemain, getTokenRemainQuota(t, tokenID))
	assert.Equal(t, 0, getTokenUsedQuota(t, tokenID))
}

func TestPostConsumeQuota_UnlimitedTokenSkipsMutation(t *testing.T) {
	truncate(t)

	const userID, tokenID = 102, 102
	const initQuota = 5000
	const tokenRemain = 8000

	seedUser(t, userID, initQuota)
	seedUnlimitedToken(t, tokenID, userID, "sk-unlimited-post", tokenRemain)

	info := &relaycommon.RelayInfo{
		UserId:         userID,
		TokenId:        tokenID,
		TokenKey:       "sk-unlimited-post",
		TokenUnlimited: true,
	}

	require.NoError(t, PostConsumeQuota(info, 1200, 0, false))
	assert.Equal(t, initQuota-1200, getUserQuota(t, userID))
	assert.Equal(t, tokenRemain, getTokenRemainQuota(t, tokenID))
	assert.Equal(t, 0, getTokenUsedQuota(t, tokenID))
}

func TestBillingSessionSettle_UnlimitedTokenSkipsMutation(t *testing.T) {
	truncate(t)

	const userID, tokenID = 103, 103
	const initQuota = 7000
	const tokenRemain = 6000

	seedUser(t, userID, initQuota)
	seedUnlimitedToken(t, tokenID, userID, "sk-unlimited-session", tokenRemain)

	session := &BillingSession{
		relayInfo: &relaycommon.RelayInfo{
			UserId:         userID,
			TokenId:        tokenID,
			TokenKey:       "sk-unlimited-session",
			TokenUnlimited: true,
		},
		funding:          &WalletFunding{userId: userID},
		preConsumedQuota: 1000,
	}

	require.NoError(t, session.Settle(1500))
	assert.Equal(t, initQuota-500, getUserQuota(t, userID))
	assert.Equal(t, tokenRemain, getTokenRemainQuota(t, tokenID))
	assert.Equal(t, 0, getTokenUsedQuota(t, tokenID))
}

func TestRefundTaskQuota_UnlimitedTokenSkipsMutation(t *testing.T) {
	truncate(t)

	ctx := context.Background()

	const userID, tokenID, channelID = 104, 104, 104
	const initQuota, preConsumed = 10000, 2500
	const tokenRemain = 7500

	seedUser(t, userID, initQuota)
	seedUnlimitedToken(t, tokenID, userID, "sk-unlimited-task", tokenRemain)
	seedChannel(t, channelID)

	task := makeTask(userID, channelID, preConsumed, tokenID, BillingSourceWallet, 0)

	RefundTaskQuota(ctx, task, "task failed")

	assert.Equal(t, initQuota+preConsumed, getUserQuota(t, userID))
	assert.Equal(t, tokenRemain, getTokenRemainQuota(t, tokenID))
	assert.Equal(t, 0, getTokenUsedQuota(t, tokenID))
}

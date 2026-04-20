package model

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTopupRechargeTest(t *testing.T) {
	t.Helper()

	require.NoError(t, DB.AutoMigrate(&User{}, &TopUp{}, &Log{}))
	require.NoError(t, LOG_DB.AutoMigrate(&Log{}))
	truncateTopupRechargeTables(t)
	t.Cleanup(func() {
		truncateTopupRechargeTables(t)
	})
}

func truncateTopupRechargeTables(t *testing.T) {
	t.Helper()

	require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&TopUp{}).Error)
	require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Log{}).Error)
	require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{}).Error)
}

func createRechargeTestUser(t *testing.T, username string, quota int) *User {
	t.Helper()

	user := &User{
		Username:    username,
		Password:    "password123",
		DisplayName: username,
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Quota:       quota,
		Group:       "default",
		AffCode:     username + "_aff",
	}
	require.NoError(t, DB.Create(user).Error)
	return user
}

func TestRechargeRejectsNonStripeOrder(t *testing.T) {
	setupTopupRechargeTest(t)

	user := createRechargeTestUser(t, "topup_reject", 12345)
	topUp := &TopUp{
		UserId:        user.Id,
		Amount:        1000,
		Money:         9000,
		TradeNo:       "USR17NOwRXNuf1776374675",
		PaymentMethod: "alipay",
		Status:        common.TopUpStatusPending,
		CreateTime:    common.GetTimestamp(),
	}
	require.NoError(t, DB.Create(topUp).Error)

	err := Recharge(topUp.TradeNo, "cus_exploit")
	require.Error(t, err)

	var reloadedTopUp TopUp
	require.NoError(t, DB.First(&reloadedTopUp, topUp.Id).Error)
	assert.Equal(t, common.TopUpStatusPending, reloadedTopUp.Status)
	assert.Zero(t, reloadedTopUp.CompleteTime)

	var reloadedUser User
	require.NoError(t, DB.First(&reloadedUser, user.Id).Error)
	assert.Equal(t, 12345, reloadedUser.Quota)
	assert.Empty(t, reloadedUser.StripeCustomer)

	var logs []Log
	require.NoError(t, DB.Find(&logs).Error)
	assert.Len(t, logs, 0)
}

func TestRechargeCompletesStripeOrder(t *testing.T) {
	setupTopupRechargeTest(t)

	user := createRechargeTestUser(t, "topup_stripe", 100)
	topUp := &TopUp{
		UserId:        user.Id,
		Amount:        10,
		Money:         12.5,
		TradeNo:       "ref_test_stripe_topup",
		PaymentMethod: "stripe",
		Status:        common.TopUpStatusPending,
		CreateTime:    common.GetTimestamp(),
	}
	require.NoError(t, DB.Create(topUp).Error)

	require.NoError(t, Recharge(topUp.TradeNo, "cus_test_123"))

	var reloadedTopUp TopUp
	require.NoError(t, DB.First(&reloadedTopUp, topUp.Id).Error)
	assert.Equal(t, common.TopUpStatusSuccess, reloadedTopUp.Status)
	assert.NotZero(t, reloadedTopUp.CompleteTime)

	var reloadedUser User
	require.NoError(t, DB.First(&reloadedUser, user.Id).Error)
	expectedAddedQuota := int(decimal.NewFromFloat(topUp.Money).Mul(decimal.NewFromFloat(common.QuotaPerUnit)).IntPart())
	assert.Equal(t, 100+expectedAddedQuota, reloadedUser.Quota)
	assert.Equal(t, "cus_test_123", reloadedUser.StripeCustomer)

	var logs []Log
	require.NoError(t, DB.Order("id asc").Find(&logs).Error)
	require.Len(t, logs, 1)
	assert.True(t, strings.Contains(logs[0].Content, "Stripe 在线充值成功"))
}

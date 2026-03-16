package utils

import (
	"github.com/mattermost/mattermost/server/public/model"
)

// NewMockLicense создает лицензию с включенными всеми возможностями
func NewMockLicense() *model.License {
	license := &model.License{}
	license.Features = &model.Features{}

	// Устанавливаем все фичи в true для обеспечения работы всех возможностей
	license.Features.SetDefaults()

	// Перезаписываем значения, чтобы все возможности были активны
	trueVal := true
	userCount := 1000000 // Большое число пользователей

	license.Features.Users = &userCount
	license.Features.LDAP = &trueVal
	license.Features.LDAPGroups = &trueVal
	license.Features.MFA = &trueVal
	license.Features.GoogleOAuth = &trueVal
	license.Features.Office365OAuth = &trueVal
	license.Features.OpenId = &trueVal
	license.Features.Compliance = &trueVal
	license.Features.Cluster = &trueVal
	license.Features.Metrics = &trueVal
	license.Features.MHPNS = &trueVal
	license.Features.SAML = &trueVal
	license.Features.Elasticsearch = &trueVal
	license.Features.Announcement = &trueVal
	license.Features.ThemeManagement = &trueVal
	license.Features.EmailNotificationContents = &trueVal
	license.Features.DataRetention = &trueVal
	license.Features.MessageExport = &trueVal
	license.Features.CustomPermissionsSchemes = &trueVal
	license.Features.CustomTermsOfService = &trueVal
	license.Features.GuestAccounts = &trueVal
	license.Features.GuestAccountsPermissions = &trueVal
	license.Features.IDLoadedPushNotifications = &trueVal
	license.Features.LockTeammateNameDisplay = &trueVal
	license.Features.EnterprisePlugins = &trueVal
	license.Features.AdvancedLogging = &trueVal
	license.Features.Cloud = &trueVal
	license.Features.SharedChannels = &trueVal
	license.Features.RemoteClusterService = &trueVal
	license.Features.OutgoingOAuthConnections = &trueVal
	license.Features.FutureFeatures = &trueVal

	// Устанавливаем SKU на Enterprise для максимальной совместимости
	license.SkuShortName = model.LicenseShortSkuEnterpriseAdvanced

	return license
}

// NewMockLicenseWithSku создает лицензию с заданным SKU
func NewMockLicenseWithSku(sku string) *model.License {
	license := NewMockLicense()
	license.SkuShortName = sku
	return license
}

package plans

import (
	"github.com/mattermost/mattermost/server/public/model"
)

// MockLicense создает лицензию с включенными всеми возможностями
func MockLicense() *model.License {
	license := &model.License{}
	license.Features = &model.Features{}

	// Устанавливаем все фичи в true для обеспечения работы всех возможностей
	license.Features.SetDefaults()

	// Перезаписываем значения, чтобы все возможности были активны
	trueVal := true
	falseVal := false
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

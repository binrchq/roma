// translation.go

package i18n

import (
	"binrc.com/roma/core/constants"
	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle *goi18n.Bundle
var localizer *goi18n.Localizer

func LoadTranslations() {
	bundle = goi18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// 加载语言文件
	bundle.MustLoadMessageFile(constants.BASE_DIR + "/i18n/zh.toml") // Chinese
	bundle.MustLoadMessageFile(constants.BASE_DIR + "/i18n/en.toml") // English
	bundle.MustLoadMessageFile(constants.BASE_DIR + "/i18n/ru.toml") // Russian

	// 设置当前语言环境
	localizer = goi18n.NewLocalizer(bundle, "zh")
}

func T(messageID string, args ...interface{}) string {
	return localizer.MustLocalize(&goi18n.LocalizeConfig{
		DefaultMessage: &goi18n.Message{
			ID:    messageID,
			Other: messageID,
		},
		TemplateData: args,
	})
}

func SetLang(lang string) {
	localizer = goi18n.NewLocalizer(bundle, lang)
}

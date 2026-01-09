// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package i18n

import (
	"context"
	"embed"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed translations/*.toml
var translationFS embed.FS

var bundle *i18n.Bundle

type localeContextKey struct{}
type localizerContextKey struct{}

// Init initializes the i18n bundle with embedded translations.
func Init() error {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	files := []string{
		"translations/active.en.toml",
		"translations/active.de.toml",
	}

	for _, file := range files {
		if _, err := bundle.LoadMessageFileFS(translationFS, file); err != nil {
			return err
		}
	}

	return nil
}

// WithLocale adds the locale to the context.
func WithLocale(ctx context.Context, lang language.Tag) context.Context {
	locale := lang.String()
	ctx = context.WithValue(ctx, localeContextKey{}, locale)
	localizer := i18n.NewLocalizer(bundle, locale)
	return context.WithValue(ctx, localizerContextKey{}, localizer)
}

// GetLocale returns the current locale from context.
func GetLocale(ctx context.Context) string {
	if locale, ok := ctx.Value(localeContextKey{}).(string); ok {
		return locale
	}
	return "en"
}

// T translates a message by ID.
func T(ctx context.Context, messageID string) string {
	localizer := getLocalizer(ctx)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})
	if err != nil {
		return messageID
	}
	return msg
}

// TData translates a message with template data.
func TData(ctx context.Context, messageID string, data map[string]any) string {
	localizer := getLocalizer(ctx)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		return messageID
	}
	return msg
}

// TPlural translates a message with plural support.
func TPlural(ctx context.Context, messageID string, count int) string {
	localizer := getLocalizer(ctx)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		PluralCount:  count,
		TemplateData: map[string]any{"Count": count},
	})
	if err != nil {
		return messageID
	}
	return msg
}

// MatchLanguage matches the best language from Accept-Language header.
func MatchLanguage(acceptLanguage string) language.Tag {
	matcher := language.NewMatcher([]language.Tag{
		language.English,
		language.German,
	})
	tag, _ := language.MatchStrings(matcher, acceptLanguage)
	return tag
}

func getLocalizer(ctx context.Context) *i18n.Localizer {
	if localizer, ok := ctx.Value(localizerContextKey{}).(*i18n.Localizer); ok {
		return localizer
	}
	return i18n.NewLocalizer(bundle, "en")
}

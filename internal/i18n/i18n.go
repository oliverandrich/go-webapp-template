// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

// Package i18n provides internationalization support for the application.
package i18n

import (
	"context"
	"embed"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed translations/*.toml
var translationFS embed.FS

// Supported languages
var (
	DefaultLanguage    = language.English
	SupportedLanguages = []language.Tag{language.English, language.German}
)

// bundle holds all translations
var bundle *i18n.Bundle

// localeContextKey is the context key for the current locale
type localeContextKey struct{}

// localizer context key
type localizerContextKey struct{}

// Init initializes the i18n bundle with all translation files.
func Init() error {
	bundle = i18n.NewBundle(DefaultLanguage)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Load English (default)
	if _, err := bundle.LoadMessageFileFS(translationFS, "translations/active.en.toml"); err != nil {
		return fmt.Errorf("failed to load English translations: %w", err)
	}

	// Load German
	if _, err := bundle.LoadMessageFileFS(translationFS, "translations/active.de.toml"); err != nil {
		return fmt.Errorf("failed to load German translations: %w", err)
	}

	return nil
}

// WithLocale returns a new context with the given locale set.
func WithLocale(ctx context.Context, lang language.Tag) context.Context {
	localizer := i18n.NewLocalizer(bundle, lang.String())
	ctx = context.WithValue(ctx, localeContextKey{}, lang.String())
	ctx = context.WithValue(ctx, localizerContextKey{}, localizer)
	return ctx
}

// GetLocale returns the locale string from the context (e.g., "en", "de").
func GetLocale(ctx context.Context) string {
	if locale, ok := ctx.Value(localeContextKey{}).(string); ok {
		return locale
	}
	return DefaultLanguage.String()
}

// getLocalizer returns the localizer from context
func getLocalizer(ctx context.Context) *i18n.Localizer {
	if loc, ok := ctx.Value(localizerContextKey{}).(*i18n.Localizer); ok {
		return loc
	}
	return i18n.NewLocalizer(bundle, DefaultLanguage.String())
}

// T translates a simple message by its ID.
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

// TPlural translates a message with pluralization.
func TPlural(ctx context.Context, messageID string, count int) string {
	localizer := getLocalizer(ctx)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:   messageID,
		PluralCount: count,
		TemplateData: map[string]any{
			"Count": count,
		},
	})
	if err != nil {
		return messageID
	}
	return msg
}

// MatchLanguage finds the best matching language from Accept-Language header.
func MatchLanguage(acceptLanguage string) language.Tag {
	matcher := language.NewMatcher(SupportedLanguages)
	userPrefs, _, _ := language.ParseAcceptLanguage(acceptLanguage)
	tag, _, _ := matcher.Match(userPrefs...)
	return tag
}

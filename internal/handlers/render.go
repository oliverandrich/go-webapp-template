// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package handlers

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// Render renders a templ component with the given status code.
func Render(c echo.Context, statusCode int, component templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := component.Render(c.Request().Context(), buf); err != nil {
		return err
	}

	return c.HTML(statusCode, buf.String())
}

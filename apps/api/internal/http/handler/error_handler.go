package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSONErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	status := http.StatusInternalServerError
	message := http.StatusText(status)

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		status = httpErr.Code
		message = errorMessage(httpErr.Message, http.StatusText(status))
	} else if err != nil {
		message = err.Error()
	}

	if err := c.JSON(status, ErrorBody{
		Error: ErrorDetail{
			Code:    http.StatusText(status),
			Message: message,
		},
	}); err != nil {
		c.Logger().Error(err)
	}
}

func errorMessage(value any, fallback string) string {
	switch typed := value.(type) {
	case string:
		if typed != "" {
			return typed
		}
	case error:
		return typed.Error()
	default:
		if typed != nil {
			return fmt.Sprint(typed)
		}
	}

	return fallback
}

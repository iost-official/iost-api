package middleware

import (
	"github.com/iost-official/iost-api/controller"
	"github.com/labstack/echo"
	"net/http"
)

func CustomHTTPErrorHandler(err error, c echo.Context) {
	c.JSON(http.StatusOK, controller.ErrorResponse{
		Code:    1,
		Message: err.Error(),
	})
}

package controller

import (
	"net/http"

	"github.com/iost-official/iost-api/model"
	"github.com/labstack/echo"
)

func GetMarket(c echo.Context) error {
	marketInfo, err := model.GetMarketInfo()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, FormatResponse(marketInfo))
}

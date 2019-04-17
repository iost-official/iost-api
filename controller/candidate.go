package controller

import (
	"log"
	"net/http"
	"strconv"

	"github.com/iost-official/iost-api/model"
	"github.com/iost-official/iost-api/model/db"
	"github.com/labstack/echo"
)

type CandidateOutput struct {
	Candidates []*model.Candidate `json:"candidates"`
	TotalCount int                `json:"total_count"`
}

func GetCandidates(c echo.Context) error {
	page := c.QueryParam("page")
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt <= 0 {
		pageInt = 1
	}

	size := c.QueryParam("size")
	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt <= 0 {
		sizeInt = CandidateDefaultCount
	}

	candidates, err := model.GetCandidates(pageInt, sizeInt)
	if err != nil {
		log.Printf("Get candidates failed. err=%v", err)
		return err
	}
	totalCount, err := db.GetCandidateCount()
	if err != nil {
		log.Printf("Get candidate count failed. err=%v", err)
	}
	ret := &CandidateOutput{
		Candidates: candidates,
		TotalCount: totalCount,
	}
	return c.JSON(http.StatusOK, FormatResponse(ret))
}

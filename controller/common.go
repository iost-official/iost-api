package controller

type CommOutput struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}

const (
	ContractTxEachPage = 25
	AccountTxEachPage  = 25
	AccountMaxPage     = 20
)

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func FormatResponse(data interface{}) Response {
	return Response{
		Code: 0,
		Data: data,
	}
}

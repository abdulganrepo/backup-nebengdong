package responses

type OffsetPagination struct {
	TotalDataOnPage int64 `json:"totalDataOnPage"`
	TotalData       int64 `json:"totalData"`
}

func (hrsci *HttpResponseStatusCodesImpl) NewResponsesOffsetPagination(data any, size int64, totalData int64, message string) Responses {
	return &ResponsesImpl{
		Data:    data,
		Code:    hrsci.Code,
		Status:  hrsci.Status,
		Message: message,
		Pagination: OffsetPagination{
			TotalDataOnPage: size,
			TotalData:       totalData,
		},
	}
}

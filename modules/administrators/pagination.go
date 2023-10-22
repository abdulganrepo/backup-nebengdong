package administrators

type Pagination struct {
	TotalDataOnPage int64 `json:"totalDataOnPage"`
	TotalData       int64 `json:"totalData"`
}

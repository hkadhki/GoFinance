package domain

type BulkImportError struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

type BulkImportResult struct {
	Accepted int64             `json:"accepted"`
	Rejected int64             `json:"rejected"`
	Errors   []BulkImportError `json:"errors"`
}

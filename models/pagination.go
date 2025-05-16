package models

type PaginationResult[T any] struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	Page    int  `json:"page"`
	Pages   int  `json:"pages"`
	HasNext bool `json:"has_next"`
	Data    []T  `json:"data"`
}

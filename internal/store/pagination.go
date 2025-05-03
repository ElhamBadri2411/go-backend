package store

import (
	"net/http"
	"strconv"
)

type PaginatedFeedQuery struct {
	Limit  int    `json:"limit" validate:"gte=0"`
	Offset int    `json:"offset" validate:"gte=0"`
	Sort   string `json:"sort" validate:"oneof=asc desc"`
}

const (
	LIMIT  string = "limit"
	OFFSET string = "offset"
	SORT   string = "sort"
)

func (pfq PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {
	query := r.URL.Query()
	limitParam := query.Get(LIMIT)
	offsetParam := query.Get(OFFSET)
	sortParam := query.Get(SORT)

	if limitParam != "" {
		limit, err := strconv.Atoi(limitParam)
		if err != nil {
			return pfq, err
		}

		pfq.Limit = limit
	}
	if offsetParam != "" {
		offset, err := strconv.Atoi(offsetParam)
		if err != nil {
			return pfq, err
		}

		pfq.Offset = offset
	}

	if sortParam != "" {
		pfq.Sort = sortParam
	}

	return pfq, nil
}

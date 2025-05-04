package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=0"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
}

const (
	LIMIT  string = "limit"
	OFFSET string = "offset"
	SORT   string = "sort"
	TAGS   string = "tags"
	SEARCH string = "search"
	SINCE  string = "since"
	UNTIL  string = "until"
)

func (pfq PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {
	query := r.URL.Query()
	limitParam := query.Get(LIMIT)
	offsetParam := query.Get(OFFSET)
	sortParam := query.Get(SORT)
	tags := query.Get(TAGS)
	search := query.Get(SEARCH)
	since := query.Get(SINCE)
	until := query.Get(UNTIL)

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

	if tags != "" {
		pfq.Tags = strings.Split(tags, ",")
	}

	if search != "" {
		pfq.Search = search
	}

	if since != "" {
		pfq.Since = parseTime(since)
	}

	if until != "" {
		pfq.Until = parseTime(until)
	}

	return pfq, nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}

	return t.Format(time.DateTime)
}

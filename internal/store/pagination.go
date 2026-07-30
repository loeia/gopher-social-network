package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
}

func (p *PaginatedFeedQuery) Parse(r *http.Request) (*PaginatedFeedQuery, error) {
	qs := r.URL.Query()

	limit := qs.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return nil, err
		}
		p.Limit = l
	}

	offset := qs.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return nil, err
		}
		p.Offset = o
	}

	sort := qs.Get("sort")
	if sort != "" {
		p.Sort = sort
	}

	tags := qs.Get("tags")
	if tags != "" {
		p.Tags = strings.Split(tags, ",")
	} else {
		p.Tags = []string{}
	}

	search := qs.Get("search")
	if search != "" {
		p.Search = search
	}

	// TODO:
	// The format sent must be "2006-01-02 15:04:05"
	since := qs.Get("since")
	if since != "" {
		p.Since = parseTime(since)
	}

	until := qs.Get("until")
	if until != "" {
		p.Until = parseTime(until)
	}

	return p, nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04:05-07:00")
}

package store

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"omitempty,oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Author string   `json:"author" validate:"max=100"`
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

	author := qs.Get("author")
	if author != "" {
		p.Author = author
	}

	since := qs.Get("since")
	if since != "" {
		s, err := parseTime(since)
		if err != nil {
			return nil, err
		}
		p.Since = s
	}

	until := qs.Get("until")
	if until != "" {
		u, err := parseTime(until)
		if err != nil {
			return nil, err
		}
		p.Until = u
	}

	return p, nil
}

// parseTime parses an RFC 3339 / ISO 8601 absolute time, e.g.
// "2026-08-01T02:00:00Z" or "2026-08-01T10:00:00+08:00", and normalizes it to UTC.
// Offsets/zone are required;
func parseTime(s string) (string, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return "", fmt.Errorf("invalid time %q: expected RFC 3339 (ISO 8601), e.g. 2026-08-01T02:00:00Z", s)
	}
	return t.UTC().Format("2006-01-02 15:04:05-07:00"), nil
}

package gallery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
)

func (self Gallery) Query(ctx context.Context, term string) iter.Seq2[ExtensionMeta, error] {
	return func(yield func(ExtensionMeta, error) bool) {
		const path = "/_apis/public/gallery/extensionquery"

		// Construct the URL
		url := self.BaseURL.JoinPath(path).String()

		nextToken := ""
		for {
			// Construct the extension query request
			queryRequest := defaultQueryRequest
			// Insert the paging token, if we have one from a previous iteration
			if nextToken != "" {
				queryRequest.Filters[0].PagingToken = nextToken
			}
			// Append the filter
			//
			// Seriously, this API design is.. Rough to work with.
			queryRequest.Filters[0].Criteria = append(queryRequest.Filters[0].Criteria, QueryFilterCriteria{
				FilterType: QueryFilterTypeTerm,
				Value:      term,
			})
			// JSON-encode the query request
			data, err := json.Marshal(queryRequest)
			if err != nil {
				yield(ExtensionMeta{}, fmt.Errorf(
					"failed to JSON-encode queryRequest: %w",
					err,
				))
				return
			}

			// Init the HTTP request
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
			if err != nil {
				yield(ExtensionMeta{}, fmt.Errorf("failed to init POST request: %w", err))
				return
			}
			req.Header.Set("user-agent", userAgent)
			req.Header.Set("content-type", "application/json; charset=utf-8")

			// Get the response
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				yield(ExtensionMeta{}, fmt.Errorf(
					"failed to execute POST request to [%s]: %w",
					url, err,
				))
				return
			}
			defer res.Body.Close()

			// Evaluate request failures
			//
			// We include the response body in the error message if the status code is
			// >= 400 (hence this conditional being >1 step from the actual doing of the
			// request)
			if res.StatusCode >= http.StatusBadRequest {
				// Read the response body
				body, err := io.ReadAll(res.Body)
				if err == nil {
					yield(ExtensionMeta{}, fmt.Errorf(
						"received HTTP status code [%d] in POST request to [%s]: %s",
						res.StatusCode, url, string(body),
					))
					return
				}
				if !yield(ExtensionMeta{}, fmt.Errorf(
					"received HTTP status code [%d] in POST request to [%s]",
					res.StatusCode, url,
				)) {
					return
				}
			}

			// Decode the response body
			var queryResponse ExtensionQueryResponse
			err = json.NewDecoder(res.Body).Decode(&queryResponse)
			if err != nil {
				responseBody, _ := io.ReadAll(res.Body)
				yield(ExtensionMeta{}, fmt.Errorf(
					"failed to decode response body [%s]: %w",
					string(responseBody), err,
				))
				return
			}

			// Ship results as we receive them
			for _, result := range queryResponse.Results {
				for _, extension := range result.Extensions {
					if !yield(extension, nil) {
						return
					}
				}
			}

			if queryResponse.PagingToken == "" {
				// If we're out of paginated results, return
				return

			} else {
				// Otherwise, get the next page's results
				nextToken = queryResponse.PagingToken
			}
		}
	}
}

var (
	// Just add a defaultQueryRequest.Filters[0].Criteria of type Term containing
	// a value of the search term
	defaultQueryRequest = QueryRequest{
		AssetTypes: []AssetType{
			Branding,
			Default,
			Small,
		},
		Filters: []QueryFilter{
			{
				Criteria: []QueryFilterCriteria{
					{
						FilterType: QueryFilterTypeProduct,
						Value:      "Microsoft.VisualStudio.Code",
					},
					{
						FilterType: QueryFilterTypeIDK,
						Value:      "37888",
					},
				},
				Direction:  2,
				PageNumber: 1,
				PageSize:   20,
				SortBy:     0,
				SortOrder:  0,
			},
		},
		Flags: QueryRequestFlagsDefault,
	}
)

type AssetType = string

const (
	Default  AssetType = "Microsoft.VisualStudio.Services.Icons.Default"
	Branding AssetType = "Microsoft.VisualStudio.Services.Icons.Branding"
	Small    AssetType = "Microsoft.VisualStudio.Services.Icons.Small"
)

type QueryRequest struct {
	AssetTypes []AssetType   `json:"assetTypes"`
	Filters    []QueryFilter `json:"filters"`
	Flags      QueryRequestFlags
}

type QueryRequestFlags uint16

const (
	QueryRequestFlagsDefault QueryRequestFlags = 870
)

type QueryFilter struct {
	Criteria    []QueryFilterCriteria `json:"criteria"`
	Direction   uint8                 `json:"direction"`
	PageNumber  uint16                `json:"pageNumber"`
	PageSize    uint16                `json:"pageSize"`
	SortBy      uint8                 `json:"sortBy"`
	SortOrder   uint8                 `json:"sortOrder"`
	PagingToken string                `json:"pagingToken"`
}

type QueryFilterCriteria struct {
	FilterType QueryFilterType
	Value      string
}

type QueryFilterType uint8

const (
	QueryFilterTypeTerm    QueryFilterType = 10
	QueryFilterTypeProduct QueryFilterType = 8
	QueryFilterTypeIDK     QueryFilterType = 12
)

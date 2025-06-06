package gallery

import "time"

type ExtensionQueryResponse struct {
	Results     []ExtensionQueryResult `json:"results"`
	PagingToken string                 `json:"pagingToken"`
}

type ExtensionQueryResult struct {
	Extensions []ExtensionMeta `json:"extensions"`
}

type ExtensionMeta struct {
	Publisher   Publisher `json:"publisher"`
	ID          string    `json:"id"`
	Name        string    `json:"extensionName"`
	DisplayName string    `json:"displayName"`
	Flags       string    `json:"flags"`
	LastUpdated time.Time `json:"lastUpdated"`
	Published   time.Time `json:"published"`
	Description string    `json:"description"`
	Versions    []Version `json:"versions"`
}

type Publisher struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Flags       string `json:"flags"`
	Domain      string `json:"domain"`
}

type PublisherFlags uint8

const (
	PublisherVerified PublisherFlags = 1 << iota
)

type Flags uint8

const (
	ExtensionValidated Flags = 1 << iota
	ExtensionPublic
)

type Version struct {
	Version    string    `json:"version"`
	Flags      string    `json:"flags"`
	LastUpated time.Time `json:"lastUpdated"`
}

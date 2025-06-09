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
	Publisher   Publisher   `json:"publisher"`
	ID          string      `json:"id"`
	Name        string      `json:"extensionName"`
	DisplayName string      `json:"displayName"`
	Flags       string      `json:"flags"`
	LastUpdated time.Time   `json:"lastUpdated"`
	Published   time.Time   `json:"published"`
	Description string      `json:"description"`
	Versions    []Version   `json:"versions"`
	Statistics  []Statistic `json:"statistics"`
}

type Publisher struct {
	ID          string `json:"publisherId"`
	Name        string `json:"publisherName"`
	DisplayName string `json:"displayName"`
	Flags       string `json:"flags"`
	Domain      string `json:"domain"`
}

type Version struct {
	Version    string    `json:"version"`
	Flags      string    `json:"flags"`
	LastUpated time.Time `json:"lastUpdated"`
}

type Statistic struct {
	Kind  string
	Value float64
}

const (
	StatisticKindInstall = "install"
)

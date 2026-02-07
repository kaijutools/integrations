package appstore

type AppsResponse struct {
	Data  []App `json:"data"`
	Links Links `json:"links,omitempty"`
	Meta  Meta  `json:"meta,omitempty"`
}

type App struct {
	Type       string        `json:"type"`
	ID         string        `json:"id"`
	Attributes AppAttributes `json:"attributes"`
}

type AppAttributes struct {
	Name               string                 `json:"name"`
	BundleID           string                 `json:"bundleId"`
	Sku                string                 `json:"sku"`
	PrimaryLocale      string                 `json:"primaryLocale"`
	IsOrphaned         bool                   `json:"isOrphaned"`
	ContentRights      string                 `json:"contentRights,omitempty"`
	AssetDeliveryState map[string]interface{} `json:"assetDeliveryState,omitempty"` // simplified
}

type Links struct {
	Self string `json:"self"`
	Next string `json:"next,omitempty"`
}

type Meta struct {
	Paging Paging `json:"paging,omitempty"`
}

type Paging struct {
	Total int `json:"total"`
	Limit int `json:"limit"`
}

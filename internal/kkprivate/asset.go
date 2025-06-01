package kkprivate

type Asset struct {
	AssetID     string `json:"assetId"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
	FileName    string `json:"fileName"`
	Error       string `json:"error"`
}

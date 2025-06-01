package karakeepbot

type AssetType string

type Asset struct {
	ID        string    `json:"id"`
	AssetType AssetType `json:"asset_type"`
}

const (
	AssetTypeScreenshot        AssetType = "screenshot"
	AssetTypeAssetScreenshot   AssetType = "assetScreenshot"
	AssetTypeBannerImage       AssetType = "bannerImage"
	AssetTypeFullPageArchive   AssetType = "fullPageArchive"
	AssetTypeVideo             AssetType = "video"
	AssetTypeBookmarkAsset     AssetType = "bookmarkAsset"
	AssetTypePrecrawledArchive AssetType = "precrawledArchive"
	AssetTypeUnknown           AssetType = "unknown"
)

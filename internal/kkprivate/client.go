package kkprivate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
)

type Client struct {
	URL   string
	Token string

	HTTPClient *http.Client
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func (c *Client) createFormFile(w *multipart.Writer, fieldname, filename string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set(
		"Content-Disposition",
		fmt.Sprintf(
			`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldname),
			escapeQuotes(filename),
		),
	)
	h.Set("Content-Type", "image/png")
	return w.CreatePart(h)
}

func (c *Client) CreateAsset(assetContent io.Reader) (*Asset, error) {
	var requestBody bytes.Buffer

	mw := multipart.NewWriter(&requestBody)
	fw, err := c.createFormFile(mw, "file", "image.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create form field: %w", err)
	}
	if written, err := io.Copy(fw, assetContent); err != nil {
		return nil, fmt.Errorf("failed to copy asset to multipart field: %w", err)
	} else if written == 0 {
		return nil, fmt.Errorf("nothing has been written to the multipart field")
	}
	mw.Close()

	request, err := http.NewRequest("POST", c.URL+"/assets", &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make a request: %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	request.Header.Set("Content-Type", mw.FormDataContentType())
	request.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var asset Asset
	if err := json.Unmarshal(respbody, &asset); err != nil {
		return nil, fmt.Errorf("failed to unmarshal body: %w", err)
	}

	if asset.Error != "" {
		return nil, fmt.Errorf("found error: %s", asset.Error)
	}

	if asset.Size == 0 {
		return nil, fmt.Errorf("the uploaded asset has size=0")
	}

	return &asset, nil
}

package cloudinary

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	cld "github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type Client struct {
	cld       *cld.Cloudinary
	folder    string
	cloudNam  string
	apiKey    string
	apiSecret string
}

func NewClient(cloudName, apiKey, apiSecret, folder string) (*Client, error) {
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("cloudinary credentials are required")
	}
	c, err := cld.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("cloudinary init: %w", err)
	}
	return &Client{
		cld:       c,
		folder:    folder,
		cloudNam:  cloudName,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}, nil
}

type UploadResult struct {
	URL      string
	PublicID string
	Format   string
	Width    int
	Height   int
	Bytes    int64
}

func (c *Client) Upload(ctx context.Context, data []byte, filename string) (*UploadResult, error) {
	params := uploader.UploadParams{
		Folder:       c.folder,
		ResourceType: "image",
	}

	resp, err := c.cld.Upload.Upload(ctx, data, params)
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload: %w", err)
	}
	return &UploadResult{
		URL:      resp.SecureURL,
		PublicID: resp.PublicID,
		Format:   resp.Format,
		Width:    resp.Width,
		Height:   resp.Height,
		Bytes:    int64(resp.Bytes),
	}, nil
}

func (c *Client) Delete(ctx context.Context, publicID string) error {
	_, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	if err != nil {
		return fmt.Errorf("cloudinary delete: %w", err)
	}
	return nil
}

type SignedUploadParams struct {
	UploadURL string            `json:"upload_url"`
	Params    map[string]string `json:"params"`
	Signature string            `json:"signature"`
	APIKey    string            `json:"api_key"`
}

func (c *Client) SignUpload(folder string, validity time.Duration) (*SignedUploadParams, error) {
	if validity <= 0 {
		validity = time.Hour
	}
	ts := time.Now().Unix()
	if folder == "" {
		folder = c.folder
	}

	params := map[string]string{
		"timestamp": fmt.Sprintf("%d", ts),
		"folder":    folder,
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(params[k])
		b.WriteString("&")
	}
	toSign := b.String()
	toSign = toSign[:len(toSign)-1]

	h := sha1.Sum([]byte(toSign + c.apiSecret))
	sig := hex.EncodeToString(h[:])

	return &SignedUploadParams{
		UploadURL: fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", c.cloudNam),
		Params:    params,
		Signature: sig,
		APIKey:    c.apiKey,
	}, nil
}

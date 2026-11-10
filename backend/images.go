package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	_ "golang.org/x/image/webp"
	"golang.org/x/sync/singleflight"
)

const maxImageBytes = 15 << 20
const maxImagePixels = 24_000_000

var imageSizes = map[string]image.Point{
	"collection": {X: 800, Y: 1000},
	"thumbnail":  {X: 160, Y: 160},
	"detail":     {X: 1200, Y: 1500},
}

type imageService struct {
	cacheDir string
	client   *http.Client
	jobs     chan struct{}
	requests singleflight.Group
}

func newImageService(cacheDir string, client *http.Client) *imageService {
	return &imageService{cacheDir: cacheDir, client: client, jobs: make(chan struct{}, 2)}
}

func (s *imageService) handle(c *gin.Context) {
	source := c.Query("url")
	variant := c.DefaultQuery("variant", "collection")
	size, ok := imageSizes[variant]
	if !validURL(source) || !ok || len(source) > 8192 {
		c.JSON(400, gin.H{"error": "画像URLまたはサイズが不正です。"})
		return
	}
	// Include a version so a future conversion change can invalidate cached files.
	key := fmt.Sprintf("%x", sha256.Sum256([]byte("v1:"+variant+":"+source)))
	path := filepath.Join(s.cacheDir, key+".jpg")
	if _, err := os.Stat(path); err != nil {
		// Shared work has its own timeout: one disconnected viewer should not cancel
		// generation for every other viewer waiting for the same image.
		result := s.requests.DoChan(key, func() (any, error) {
			if _, err := os.Stat(path); err == nil {
				return nil, nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			select {
			case s.jobs <- struct{}{}:
				defer func() { <-s.jobs }()
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			return nil, s.generate(ctx, source, path, size)
		})
		select {
		case res := <-result:
			if res.Err != nil {
				c.JSON(422, gin.H{"error": "画像を変換できませんでした。"})
				return
			}
		case <-c.Request.Context().Done():
			return
		}
	}
	c.Header("Content-Type", "image/jpeg")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "public, max-age=86400")
	c.File(path)
}

func (s *imageService) generate(ctx context.Context, source, path string, size image.Point) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "YourSelect/0.1 (Product image preview)")
	req.Header.Set("Accept", "image/jpeg,image/png,image/webp,image/gif")
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("image HTTP %d", res.StatusCode)
	}
	if res.ContentLength > maxImageBytes {
		return errors.New("image too large")
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, maxImageBytes+1))
	if err != nil {
		return err
	}
	if len(data) > maxImageBytes {
		return errors.New("image too large")
	}
	converted, err := convertImage(data, size)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(s.cacheDir, 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(s.cacheDir, "image-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(converted); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

// Fit the entire product onto a fixed white canvas, without stretching or
// cropping. DecodeConfig bounds memory before decoding untrusted image data.
func convertImage(data []byte, size image.Point) ([]byte, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if config.Width <= 0 || config.Height <= 0 || int64(config.Width)*int64(config.Height) > maxImagePixels {
		return nil, errors.New("image dimensions too large")
	}
	src, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	fitted := imaging.Fit(src, size.X, size.Y, imaging.Lanczos)
	canvas := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	offset := image.Pt((size.X-fitted.Bounds().Dx())/2, (size.Y-fitted.Bounds().Dy())/2)
	draw.Draw(canvas, fitted.Bounds().Add(offset), fitted, fitted.Bounds().Min, draw.Over)
	var out bytes.Buffer
	if err = jpeg.Encode(&out, canvas, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

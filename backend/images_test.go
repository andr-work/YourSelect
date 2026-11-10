package main

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
)

func testImage(t *testing.T, w, h int, format string) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 220, A: 255}}, image.Point{}, draw.Src)
	var buf bytes.Buffer
	var err error
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, nil)
	case "gif":
		err = gif.Encode(&buf, img, nil)
	default:
		err = png.Encode(&buf, img)
	}
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
func TestImageCanvas(t *testing.T) {
	for _, shape := range []struct {
		name string
		w, h int
	}{{"portrait", 100, 400}, {"landscape", 400, 100}, {"square", 100, 100}} {
		for variant, size := range imageSizes {
			t.Run(shape.name+"/"+variant, func(t *testing.T) {
				data, err := convertImage(testImage(t, shape.w, shape.h, "png"), size)
				if err != nil {
					t.Fatal(err)
				}
				img, err := jpeg.Decode(bytes.NewReader(data))
				if err != nil {
					t.Fatal(err)
				}
				if img.Bounds().Size() != size {
					t.Fatalf("size = %v; want %v", img.Bounds().Size(), size)
				}
				// The complete red image is centered, with white rather than stretched content.
				center := color.RGBAModel.Convert(img.At(size.X/2, size.Y/2)).(color.RGBA)
				if center.R < 200 || center.G > 25 {
					t.Fatalf("center = %v", center)
				}
				edge := color.RGBAModel.Convert(img.At(0, 0)).(color.RGBA)
				if shape.w*size.Y != shape.h*size.X && (edge.R < 240 || edge.G < 240 || edge.B < 240) {
					t.Fatalf("expected white padding, got %v", edge)
				}
			})
		}
	}
	for _, format := range []string{"jpeg", "gif"} {
		if _, err := convertImage(testImage(t, 20, 40, format), imageSizes["thumbnail"]); err != nil {
			t.Fatalf("%s: %v", format, err)
		}
	}
}
func TestTransparentImageUsesWhite(t *testing.T) {
	var b bytes.Buffer
	png.Encode(&b, image.NewNRGBA(image.Rect(0, 0, 10, 10)))
	data, err := convertImage(b.Bytes(), imageSizes["thumbnail"])
	if err != nil {
		t.Fatal(err)
	}
	img, _ := jpeg.Decode(bytes.NewReader(data))
	r, g, blue, _ := img.At(80, 80).RGBA()
	if r < 65000 || g < 65000 || blue < 65000 {
		t.Fatal("transparency became dark")
	}
}
func TestRejectInvalidAndOversizedImage(t *testing.T) {
	if _, err := convertImage([]byte("<html>not an image</html>"), imageSizes["thumbnail"]); err == nil {
		t.Fatal("accepted non-image")
	}
	data := testImage(t, 1, 1, "png")
	// Change only the PNG header to claim a huge canvas, without allocating it.
	binary.BigEndian.PutUint32(data[16:20], 100_000)
	binary.BigEndian.PutUint32(data[20:24], 100_000)
	binary.BigEndian.PutUint32(data[29:33], crc32.ChecksumIEEE(data[12:29]))
	if _, err := convertImage(data, imageSizes["thumbnail"]); err == nil {
		t.Fatal("accepted excessive pixel count")
	}
}
func TestImageCacheAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	data := testImage(t, 120, 80, "png")
	var fetched atomic.Int32
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetched.Add(1)
		w.Header().Set("Content-Type", "image/png")
		w.Write(data)
	}))
	defer source.Close()
	cache := t.TempDir()
	s := newImageService(cache, source.Client())
	r := gin.New()
	r.GET("/api/images", s.handle)
	request := func(query string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/api/images?"+query, nil))
		return w
	}
	query := url.Values{"url": {source.URL}, "variant": {"collection"}}.Encode()
	for i := 0; i < 2; i++ {
		w := request(query)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(w.Body.Bytes()))
		if err != nil || format != "jpeg" || cfg.Width != 800 || cfg.Height != 1000 {
			t.Fatalf("invalid output: %+v %s %v", cfg, format, err)
		}
		if w.Header().Get("Cache-Control") == "" {
			t.Fatal("missing cache header")
		}
	}
	if fetched.Load() != 1 {
		t.Fatalf("image fetched %d times", fetched.Load())
	}
	files, _ := filepath.Glob(filepath.Join(cache, "*.jpg"))
	if len(files) != 1 {
		t.Fatalf("cache files %v", files)
	}
	// Cached images remain available even when the original server goes offline.
	source.Close()
	if w := request(query); w.Code != 200 {
		t.Fatal("cache unavailable offline")
	}
	for _, query := range []string{"url=file:///etc/passwd", "url=https://example.com/image&variant=huge"} {
		if w := request(query); w.Code != 400 {
			t.Fatalf("bad input returned %d", w.Code)
		}
	}
	blocked := newImageService(t.TempDir(), safeClient())
	r = gin.New()
	r.GET("/api/images", blocked.handle)
	if w := request("url=http://127.0.0.1/private"); w.Code != 422 {
		t.Fatalf("local request not blocked: %d", w.Code)
	}
}
func TestImageEXIFOrientation(t *testing.T) {
	data := testImage(t, 40, 20, "jpeg")
	// JPEG APP1 / Exif TIFF header with orientation=6 (rotate 90 degrees).
	exif := []byte{'E', 'x', 'i', 'f', 0, 0, 'I', 'I', 42, 0, 8, 0, 0, 0, 1, 0, 0x12, 1, 3, 0, 1, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0}
	oriented := append([]byte{}, data[:2]...)
	oriented = append(oriented, 0xff, 0xe1, 0, byte(len(exif)+2))
	oriented = append(oriented, exif...)
	oriented = append(oriented, data[2:]...)
	converted, err := convertImage(oriented, imageSizes["thumbnail"])
	if err != nil {
		t.Fatal(err)
	}
	img, _ := jpeg.Decode(bytes.NewReader(converted))
	// After rotation the portrait fills the height and leaves side padding.
	_, g, _, _ := img.At(80, 65).RGBA()
	if g > 10000 {
		t.Fatal("EXIF orientation was not applied")
	}
	_, g, _, _ = img.At(65, 80).RGBA()
	if g < 60000 {
		t.Fatal("expected side padding after EXIF rotation")
	}
}

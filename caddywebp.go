package github

import (
	"bytes"
	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	"github.com/chai2010/webp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/image/bmp"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"
)

const Quality = 80

func init() {
	log.Println("RegisterPlugin")
	caddy.RegisterPlugin("webp", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	log.Println("setupFunc")
	h := handler{}
	//for c.Next() {
	//
	//}
	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		h.next = next
		return h
	})
	return nil
}

type handler struct {
	next httpserver.Handler
}

func (s handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	ua := r.Header.Get("User-Agent")
	if strings.Contains(ua, "Safari") && !strings.Contains(ua, "Chrome") && !strings.Contains(ua, "Firefox") {
		return s.next.ServeHTTP(w, r) // 对Safari禁用webp
	}
	resp := &response{}
	i, err := s.next.ServeHTTP(resp, r)
	if err != nil {
		return i, err
	}
	ct := http.DetectContentType(resp.Body.Bytes())

	//fmt.Println("file len", resp.Body.Len(), "file type", ct)

	var decoder func(io.Reader) (image.Image, error)
	if strings.Contains(ct, "jpeg") {
		decoder = jpeg.Decode
	} else if strings.Contains(ct, ".png") {
		decoder = png.Decode
	} else if strings.Contains(ct, ".bmp") {
		decoder = bmp.Decode
		// } else if strings.HasSuffix(r.URL.String(), ".gif") { TODO need to support animated webp
		// 	decoder = gif.Decode
	} else {
		return s.next.ServeHTTP(w, r)
	}

	img, err := decoder(bytes.NewReader(resp.Body.Bytes()))
	if err != nil || img == nil {
		log.Error(err)
		return s.next.ServeHTTP(w, r)
	}
	var buf bytes.Buffer
	err = webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: Quality})
	if err != nil {
		log.Error(err)
		return s.next.ServeHTTP(w, r)
	}
	w.Header().Set("Content-Type", "image/webp")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

type response struct {
	header http.Header
	Body   bytes.Buffer
}

func (s *response) Header() http.Header {
	return http.Header{}
}

func (s *response) Write(data []byte) (int, error) {
	s.Body.Write(data)
	return len(data), nil
}
func (s *response) WriteHeader(i int) {
	return
}

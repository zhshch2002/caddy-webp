package github

import (
	"bytes"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/chai2010/webp"
	"golang.org/x/image/bmp"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"strings"
)

const Quality = 80

func init() {
	log.Println("webp plugin")
	err := caddy.RegisterModule(Webp{})
	if err != nil {
		log.Fatal(err)
	}
	httpcaddyfile.RegisterHandlerDirective("webp", parseCaddyfile)
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	return Webp{}, nil
}

type Webp struct {
}

func (Webp) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.webp",
		New: func() caddy.Module { return new(Webp) },
	}
}

func (s Webp) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	ua := r.Header.Get("User-Agent")
	if strings.Contains(ua, "Safari") && !strings.Contains(ua, "Chrome") && !strings.Contains(ua, "Firefox") {
		return next.ServeHTTP(w, r) // 对Safari禁用webp
	}
	resp := &response{}
	err := next.ServeHTTP(resp, r)
	if err != nil {
		return err
	}
	ct := http.DetectContentType(resp.Body.Bytes())

	//fmt.Println("file len", resp.Body.Len(), "file type", ct)

	var decoder func(io.Reader) (image.Image, error)
	if strings.Contains(ct, "jpeg") {
		decoder = jpeg.Decode
	} else if strings.Contains(ct, "png") {
		decoder = png.Decode
	} else if strings.Contains(ct, "bmp") {
		decoder = bmp.Decode
		// } else if strings.HasSuffix(r.URL.String(), ".gif") { TODO need to support animated webp
		// 	decoder = gif.Decode
	} else {
		return next.ServeHTTP(w, r)
	}

	img, err := decoder(bytes.NewReader(resp.Body.Bytes()))
	if err != nil || img == nil {
		log.Println(err)
		return next.ServeHTTP(w, r)
	}
	var buf bytes.Buffer
	err = webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: Quality})
	if err != nil {
		log.Println(err)
		return next.ServeHTTP(w, r)
	}
	w.Header().Set("Content-Type", "image/webp")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
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

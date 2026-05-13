// Package chartshot provides a utility to render an embedded HTML page
// via a headless browser and return a full-page PNG screenshot.
//
// Typical usage:
//
//	//go:embed static
//	var staticFS embed.FS
//
//	subFS, _ := fs.Sub(staticFS, "static")
//	page, err := chartshot.Register(subFS, "chart.html")
//	imgData, err := chartshot.ScreenshotFullPage(page)
//	page.Close() // unregister when no longer needed
package chartshot

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// ─── Global headless browser ──────────────────────────────────────────────────

var (
	globalBrowser   *rod.Browser
	browserInitOnce sync.Once
	browserInitErr  error
)

func ensureBrowser() (*rod.Browser, error) {
	browserInitOnce.Do(func() {
		controlURL, err := launcher.New().
			Headless(true).
			NoSandbox(true).
			Set("disable-gpu", "").
			Launch()
		if err != nil {
			browserInitErr = fmt.Errorf("chartshot: launch browser: %w", err)
			return
		}
		b := rod.New().ControlURL(controlURL)
		if err := b.Connect(); err != nil {
			browserInitErr = fmt.Errorf("chartshot: connect browser: %w", err)
			return
		}
		globalBrowser = b
	})
	return globalBrowser, browserInitErr
}

// ─── Global HTTP server ───────────────────────────────────────────────────────
//
// A single long-lived server handles all requests.
// Each registered Page gets a unique numeric ID; the URL layout is:
//
//	http://127.0.0.1:{port}/{id}/htmlPath

var (
	globalServerPort int
	serverInitOnce   sync.Once
	serverInitErr    error
	fsRegistry       sync.Map // key: string → value: fs.FS
	fsIDCounter      atomic.Int64
)

func ensureServer() (int, error) {
	serverInitOnce.Do(func() {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			serverInitErr = fmt.Errorf("chartshot: listen: %w", err)
			return
		}
		globalServerPort = ln.Addr().(*net.TCPAddr).Port

		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			trimmed := strings.TrimPrefix(r.URL.Path, "/")
			id, _, _ := strings.Cut(trimmed, "/")
			val, ok := fsRegistry.Load(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			http.StripPrefix("/"+id, http.FileServer(http.FS(val.(fs.FS)))).ServeHTTP(w, r)
		})

		go func() {
			_ = (&http.Server{
				Handler:           mux,
				ReadHeaderTimeout: 10 * time.Second,
			}).Serve(ln)
		}()

		slog.Info("globalServerPort", "port", globalServerPort)
	})
	return globalServerPort, serverInitErr
}

// ─── Page ─────────────────────────────────────────────────────────────────────

// Page represents a registered web page that is permanently served by the
// shared HTTP server until Close is called. It may be passed to Screenshot
// any number of times and is safe for concurrent use.
type Page struct {
	id       string // unique registry key
	htmlPath string // path within the fs.FS to the entry-point HTML file
	port     int    // cached server port
}

// Register starts the shared HTTP server (if not already running), mounts
// embeddedFS under a unique URL prefix, and returns a *Page ready for
// screenshotting.
//
// Parameters:
//   - embeddedFS: an fs.FS containing the HTML and any accompanying assets
//     (images, CSS, JS, …). embed.FS and fs.Sub results both satisfy this.
//   - htmlPath:   relative path within embeddedFS to the HTML entry point
//     (e.g. "chart.html" or "subdir/index.html").
//
// Call Close on the returned *Page when it is no longer needed.
func Register(embeddedFS fs.FS, htmlPath string) (*Page, error) {
	port, err := ensureServer()
	if err != nil {
		return nil, err
	}
	id := fmt.Sprintf("fs%d", fsIDCounter.Add(1))
	fsRegistry.Store(id, embeddedFS)
	slog.Info("Register", "id", id)
	return &Page{id: id, htmlPath: htmlPath, port: port}, nil
}

// Close removes the page from the shared HTTP server. After Close returns,
// the *Page must not be used again.
func (p *Page) Close() {
	fsRegistry.Delete(p.id)
}

// URL returns the full URL at which this page is currently being served.
// Useful for debugging.
func (p *Page) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/%s/%s", p.port, p.id, p.htmlPath)
}

// ─── Screenshot ───────────────────────────────────────────────────────────────

// ScreenshotFullPage renders p in a headless browser and returns a full-page
// PNG screenshot.
//
// The page is expected to set  document.body.dataset.chartshotReady = "true"
// after all content (charts, tables, …) has finished rendering. This function
// blocks until that attribute is present, so there is no fixed sleep needed.
//
// Safe to call concurrently on the same or different *Page values.
func ScreenshotFullPage(p *Page) (imgData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("headless browser panic: %v", r)
		}
	}()

	browser, err := ensureBrowser()
	if err != nil {
		return nil, err
	}

	page := browser.MustPage(p.URL())
	defer page.MustClose()

	slog.Debug("chartshot: waiting for chartshotReady", "url", p.URL())

	// 最多等待 30 秒，超时时抛出 panic 被上方 recover 捕获并返回 error
	page.Timeout(30 * time.Second).MustElement("body").MustWait(`
		() => { return this.dataset.chartshotReady === 'true'; }`)

	slog.Debug("chartshot: page ready, taking screenshot")

	// 获取 body 实际渲染宽高，把 viewport 收窄到恰好等于内容，消除右侧空白
	dims := page.MustEval(`() => ({
		w: document.body.scrollWidth,
		h: document.body.scrollHeight,
	})`)
	w := int(dims.Get("w").Num())
	h := int(dims.Get("h").Num())
	if w > 0 && h > 0 {
		_ = proto.EmulationSetDeviceMetricsOverride{
			Width:             w,
			Height:            h,
			DeviceScaleFactor: 1,
		}.Call(page)
	}

	return page.MustScreenshot(), nil
}

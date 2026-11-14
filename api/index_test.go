package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
)

func fixtureData(t *testing.T) []byte {
	t.Helper()
	data, err := json.Marshal(map[string]string{
		"amber":       `<path id="amber"/>`,
		"birch-auto":  `<path id="birch-auto"/>`,
		"birch-dark":  `<path id="birch-dark"/>`,
		"birch-light": `<path id="birch-light"/>`,
		"javascript":  `<path id="javascript"/>`,
	})
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func request(app http.Handler, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func TestIconRequests(t *testing.T) {
	app := newApp(fixtureData(t))
	cases := []struct {
		name        string
		path        string
		status      int
		contentType string
		contains    []string
		absent      string
	}{
		{"explicit", "/api/icons?i=amber,js", 200, "image/svg+xml", []string{`width="104" height="48" viewBox="0 0 556 256"`, `id="amber"`, `id="javascript"`, `translate(300, 0)`}, "<title>"},
		{"auto", "/api/icons?i=birch", 200, "image/svg+xml", []string{`id="birch-auto"`}, `id="birch-dark"`},
		{"dark", "/api/icons?i=birch&theme=dark", 200, "image/svg+xml", []string{`id="birch-dark"`}, ""},
		{"light", "/api/icons?i=birch&theme=light", 200, "image/svg+xml", []string{`id="birch-light"`}, ""},
		{"titles", "/api/icons?i=js&titles=false", 200, "image/svg+xml", []string{"<title>javascript</title>"}, ""},
		{"center", "/api/icons?i=amber&align=center", 200, "image/svg+xml", []string{`width="100%"`}, ""},
		{"right", "/api/icons?i=amber&align=right", 200, "image/svg+xml", []string{`width="calc(200% - 48px)"`}, ""},
		{"alignment fallback", "/api/icons?i=amber&align=other", 200, "image/svg+xml", []string{`width="48"`}, ""},
		{"rows", "/api/icons?i=amber,js&perline=1", 200, "image/svg+xml", []string{`width="48" height="104" viewBox="0 0 256 556"`, `translate(0, 300)`}, ""},
		{"missing", "/api/icons", 400, "text/plain; charset=utf-8", []string{"You didn't specify any icons!"}, ""},
		{"theme error", "/api/icons?i=amber&theme=other", 400, "text/plain; charset=utf-8", []string{"Theme must be either 'light', 'dark' or 'auto'"}, ""},
		{"unknown route", "/other", 404, "text/plain", []string{"404 page not found"}, ""},
		{"unknown icon", "/api/icons?i=absent", 200, "image/svg+xml", []string{`width="48" height="48"`}, "<path"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := request(app, tc.path)

			if w.Code != tc.status || w.Header().Get("Content-Type") != tc.contentType {
				t.Fatalf("response = %d %q, want %d %q", w.Code, w.Header().Get("Content-Type"), tc.status, tc.contentType)
			}
			for _, part := range tc.contains {
				if !strings.Contains(w.Body.String(), part) {
					t.Errorf("body does not contain %q: %s", part, w.Body.String())
				}
			}
			if tc.absent != "" && strings.Contains(w.Body.String(), tc.absent) {
				t.Errorf("body unexpectedly contains %q", tc.absent)
			}
		})
	}
}

func TestInvalidPerLine(t *testing.T) {
	app := newApp(fixtureData(t))
	for _, value := range []string{"0", "-1", "51", "abc"} {
		t.Run(value, func(t *testing.T) {
			w := request(app, "/api/icons?i=amber&perline="+value)

			if w.Code != http.StatusBadRequest || w.Header().Get("Content-Type") != "text/plain; charset=utf-8" || w.Body.String() != "Icons per line must be a number between 1 and 50" {
				t.Fatalf("response = %d %v %q", w.Code, w.Header(), w.Body.String())
			}
		})
	}
}

func TestCatalogStable(t *testing.T) {
	app := newApp(fixtureData(t))
	explicit := request(app, "/api/icons?i=amber,js&titles=true").Body.String()
	wantGroups := map[string]int{"amber": 1, "birch-auto": 3, "javascript": 1}
	pathIDs := regexp.MustCompile(`<path id="([^"]+)"/>`)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 8; j++ {
				w := request(app, "/api/icons?i=amber,js&titles=true")
				if w.Code != 200 || w.Body.String() != explicit {
					t.Error("explicit response changed")
				}
				all := request(app, "/api/icons?i=all")
				groups := map[string]int{}
				for _, match := range pathIDs.FindAllStringSubmatch(all.Body.String(), -1) {
					groups[match[1]]++
				}
				if all.Code != 200 || all.Header().Get("Content-Type") != "image/svg+xml" || !reflect.DeepEqual(groups, wantGroups) || !strings.Contains(all.Body.String(), `width="273" height="48" viewBox="0 0 1456 256"`) {
					t.Errorf("all response changed: %d %s", all.Code, all.Body.String())
				}
			}
		}()
	}
	wg.Wait()
}

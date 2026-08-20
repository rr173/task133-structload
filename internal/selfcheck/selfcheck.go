package selfcheck

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"task133-structload/internal/clock"
	"task133-structload/internal/httpapi"
	"task133-structload/internal/model"
	"task133-structload/internal/service"
	"task133-structload/internal/store"
)

// Run executes all smoke scenarios against an in-memory engine.
// It returns a summary string and a nil error when everything passes.
func Run(dbPath string) (string, error) {
	st, err := store.Open(dbPath)
	if err != nil {
		return "", fmt.Errorf("open store: %w", err)
	}
	defer st.Close()

	clk := clock.NewFake(time.UnixMilli(1740000000000))
	svc := service.New(st, clk)
	handler := httpapi.New(svc)

	var buf bytes.Buffer
	srv := httptest.NewServer(handler)
	defer srv.Close()
	client := &httpClient{base: srv.URL, log: &buf}

	if err := scenarios(client, svc); err != nil {
		return buf.String(), err
	}
	return buf.String(), nil
}

// scenarios runs the full smoke-test sequence.
func scenarios(c *httpClient, svc *service.Service) error {
	ctx := context.Background()
	steps := []struct {
		name string
		fn   func() error
	}{
		{"velocity_pressure", func() error { return scenarioVelocityPressure(c) }},
		{"wind_pressure", func() error { return scenarioWindPressure(c) }},
		{"live_reduction", func() error { return scenarioLiveReduction(c) }},
		{"small_area", func() error { return scenarioSmallAreaNoReduction(c) }},
		{"snow_slope", func() error { return scenarioSnowSlope(c) }},
		{"full_pipeline", func() error { return scenarioFullPipeline(c, svc, ctx) }},
		{"restart", func() error { return scenarioRestartConsistency(c, svc, ctx) }},
		{"override", func() error { return scenarioOverride(c) }},
	}
	for _, s := range steps {
		if err := s.fn(); err != nil {
			return fmt.Errorf("%s: %w", s.name, err)
		}
	}
	return nil
}

// httpClient wraps an httptest server for JSON calls.
type httpClient struct {
	base string
	log  *bytes.Buffer
	srv  *httptest.Server
}

func (c *httpClient) Close() {}

func (c *httpClient) do(method, path string, body any) (int, any, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, c.base+path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var out any
	if len(data) > 0 {
		_ = json.Unmarshal(data, &out)
	}
	fmt.Fprintf(c.log, "%s %s -> %d\n", method, path, resp.StatusCode)
	return resp.StatusCode, out, nil
}

func (c *httpClient) post(path string, body any) (any, error) {
	return c.expect(2, "POST", path, body)
}

func (c *httpClient) get(path string) (any, error) {
	return c.expect(2, "GET", path, nil)
}

func (c *httpClient) expect(family int, method, path string, body any) (any, error) {
	status, out, err := c.do(method, path, body)
	if err != nil {
		return nil, err
	}
	if status/100 != family {
		return out, fmt.Errorf("expected %dxx, got %d: %v", family, status, out)
	}
	return out, nil
}

func mustJSONString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

var errAssertion = errors.New("assertion failed")

func assert(cond bool, msg string) error {
	if !cond {
		return fmt.Errorf("%w: %s", errAssertion, msg)
	}
	return nil
}

func num(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case int:
		return int64(t)
	}
	return 0
}

// helper: create a standard test project.
func (c *httpClient) makeProject(code string, windSpeed, groundSnow int64) (string, error) {
	body, err := c.post("/api/projects", map[string]any{
		"code": code, "name": code + " building", "site": "test site",
		"wind_zone": "I", "snow_zone": "I", "exposure": "C", "importance": "II",
		"wind_speed": windSpeed, "ground_snow": groundSnow,
	})
	if err != nil {
		return "", err
	}
	return str(field(body, "id")), nil
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func arr(v any) []any {
	a, _ := v.([]any)
	return a
}

// objMap asserts v to map[string]any (nil if not an object).
func objMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

// field digs a named key out of an object; returns nil for non-objects/missing.
func field(v any, key string) any {
	return objMap(v)[key]
}

// idx returns the n-th element when v is a JSON array, else nil.
func idx(v any, n int) any {
	a := arr(v)
	if n < 0 || n >= len(a) {
		return nil
	}
	return a[n]
}

// Reference to model package keeps the import used even if narrowed later.
var _ = model.StatusPass

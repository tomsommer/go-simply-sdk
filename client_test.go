package simply

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductsContract(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2/my/products/" || r.Header.Get("Authorization") != "Bearer key" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL)
		}
		w.Header().Set("X-RateLimit-Remaining", "19")
		_, _ = w.Write([]byte(`[{"object":"example.test","type":"domain"}]`))
	}))
	defer ts.Close()
	c, err := NewClient(Config{APIKey: "key", BaseURL: ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	products, response, err := c.Products().List(context.Background())
	if err != nil || len(products) != 1 || response.RateLimitRemaining != "19" {
		t.Fatalf("unexpected result: %+v %+v %v", products, response, err)
	}
}

func TestProductsEnvelopeContract(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"products":[{"object":"ice.horse","name":"ice.horse"}],"status":200,"message":"success"}`))
	}))
	defer ts.Close()
	c, err := NewClient(Config{APIKey: "key", BaseURL: ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	products, response, err := c.Products().List(context.Background())
	if err != nil || response.StatusCode != http.StatusOK || len(products) != 1 || products[0].Object != "ice.horse" {
		t.Fatalf("unexpected result: %+v %+v %v", products, response, err)
	}
}
func TestDNSMutationContract(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/2/my/products/example.test/dns/records/" || r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL)
		}
		_, _ = w.Write([]byte(`{"id":"7","name":"www","type":"CNAME","value":"target.example."}`))
	}))
	defer ts.Close()
	c, err := NewClient(Config{APIKey: "key", BaseURL: ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	record, _, err := c.DNS().AddRecord(context.Background(), "example.test", Record{Name: "www", Type: "CNAME", Value: "target.example."})
	if err != nil || record.ID != "7" {
		t.Fatalf("unexpected result: %+v %v", record, err)
	}
}
func TestAPIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer ts.Close()
	c, err := NewClient(Config{APIKey: "secret", BaseURL: ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = c.Products().List(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !apiErr.IsUnauthorized() || err.Error() == "secret" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeprecatedBasicAuthStillWorks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "acct" || pass != "key" {
			t.Fatalf("unexpected authorization: %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer ts.Close()
	c, err := NewClient(Config{APIKey: "key", Account: "acct", AuthMode: AuthBasic, BaseURL: ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.Products().List(context.Background()); err != nil {
		t.Fatal(err)
	}
}

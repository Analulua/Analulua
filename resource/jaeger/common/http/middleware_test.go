package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"newdemo1/resource/jaeger/health-check/checks/uptime"
)

func newRequest(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	return req
}

func TestCORS(t *testing.T) {
	r := newRequest("GET", "http://www.example.com")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	CORS(testHandler).ServeHTTP(rr, r)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("bad status: got %v want %v", got, want)
	}
}

func TestRecover(t *testing.T) {
	r := newRequest("GET", "http://www.example.com")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	Recover(testHandler).ServeHTTP(rr, r)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("bad status: got %v want %v", got, want)
	}
}

func TestHealthCheck(t *testing.T) {
	r := newRequest("GET", "http://www.example.com")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	HealthCheck(testHandler, "/", uptime.Process(), uptime.System()).ServeHTTP(rr, r)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("bad status: got %v want %v", got, want)
	}
}

func TestWithRecovery(t *testing.T) {
	r := newRequest("GET", "http://www.example.com")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	NewHandler(testHandler, WithRecovery()).ServeHTTP(rr, r)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("bad status: got %v want %v", got, want)
	}
}

func TestWithCompression(t *testing.T) {
	r := newRequest("GET", "http://www.example.com")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	NewHandler(testHandler, WithCompression()).ServeHTTP(rr, r)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("bad status: got %v want %v", got, want)
	}
}

func TestWithCORS(t *testing.T) {
	r := newRequest("GET", "http://www.example.com")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	NewHandler(testHandler, WithCORS()).ServeHTTP(rr, r)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("bad status: got %v want %v", got, want)
	}
}

func TestWithHealthCheck(t *testing.T) {
	r := newRequest("GET", "http://www.example.com")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	NewHandler(testHandler, WithHealthCheck(uptime.Process(), uptime.System())).ServeHTTP(rr, r)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("bad status: got %v want %v", got, want)
	}
}

func TestWithHealthCheckPath(t *testing.T) {
	r := newRequest("GET", "http://www.example.com")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	NewHandler(testHandler, WithHealthCheckPath("/", uptime.Process(), uptime.System())).ServeHTTP(rr, r)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("bad status: got %v want %v", got, want)
	}
}

func TestWithDefault(t *testing.T) {
	r := newRequest("GET", "http://www.example.com")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	NewHandler(testHandler, WithDefault()).ServeHTTP(rr, r)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("bad status: got %v want %v", got, want)
	}
}

func TestNewHandler(t *testing.T) {
	r := newRequest("GET", "http://www.example.com")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	NewHandler(testHandler).ServeHTTP(rr, r)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("bad status: got %v want %v", got, want)
	}
}

func TestDefaultHandler(t *testing.T) {
	r := newRequest("GET", "http://www.example.com")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	DefaultHandler(testHandler).ServeHTTP(rr, r)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("bad status: got %v want %v", got, want)
	}
}

func TestWithHealthCheckPathFilter(t *testing.T) {
	r := newRequest("GET", "http://www.example.com")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	NewHandler(testHandler, WithHealthCheckPath("/", uptime.Process(), uptime.System())).ServeHTTP(rr, r)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("bad status: got %v want %v", got, want)
	}
}

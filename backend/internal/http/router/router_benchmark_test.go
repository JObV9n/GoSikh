package router_test

import (
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

var markLessonProgressPayload = []byte(`{"completed":true}`)

func silenceLogsForBenchmark(b *testing.B) {
	b.Helper()
	log.SetOutput(io.Discard)
	b.Cleanup(func() {
		log.SetOutput(os.Stderr)
	})
}

func BenchmarkGetCourseProgressParallel(b *testing.B) {
	silenceLogsForBenchmark(b)
	app := setupTestApp(b)
	token := registerAndGetToken(b, app.engine, "bench-progress@example.com")

	warm := doJSONBytesRequest(b, app.engine, http.MethodPut, "/api/courses/go-basics/lessons/hello-go/progress", markLessonProgressPayload, token)
	if warm.Code != http.StatusOK {
		b.Fatalf("warmup mark progress failed: %d body=%s", warm.Code, warm.Body.String())
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rr := doJSONRequest(b, app.engine, http.MethodGet, "/api/courses/go-basics/progress", nil, token)
			if rr.Code != http.StatusOK {
				b.Fatalf("get progress failed: %d body=%s", rr.Code, rr.Body.String())
			}
		}
	})
}

func BenchmarkMarkLessonProgressParallel(b *testing.B) {
	silenceLogsForBenchmark(b)
	app := setupTestApp(b)
	token := registerAndGetToken(b, app.engine, "bench-mark@example.com")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rr := doJSONBytesRequest(b, app.engine, http.MethodPut, "/api/courses/go-basics/lessons/hello-go/progress", markLessonProgressPayload, token)
			if rr.Code != http.StatusOK {
				b.Fatalf("mark progress failed: %d body=%s", rr.Code, rr.Body.String())
			}
		}
	})
}

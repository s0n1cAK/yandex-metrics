package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/s0n1cAK/yandex-metrics/internal/hash"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	filestorage "github.com/s0n1cAK/yandex-metrics/internal/storage/fileStorage"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
		wroteHeader  bool
	}

	compressWriter struct {
		w  http.ResponseWriter
		zw *gzip.Writer
	}

	compressReader struct {
		r  io.ReadCloser
		zr *gzip.Reader
	}
)

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	if !w.wroteHeader {
		w.responseData.status = statusCode
		w.ResponseWriter.WriteHeader(statusCode)
		w.wroteHeader = true
	}
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size += size
	return size, err
}

func Logging(l *zap.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			h.ServeHTTP(&lw, r)

			duration := time.Since(start)

			l.Info("",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Int("status", responseData.status),
				zap.Duration("duration", duration),
				zap.Int("size", responseData.size),
			)
		}
		return http.HandlerFunc(logFn)
	}
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func gzipCompession() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		compressFn := func(w http.ResponseWriter, r *http.Request) {

			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				w.Header().Set("Content-Encoding", "gzip")
				cw := newCompressWriter(w)
				ow = cw
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer cr.Close()
			}

			h.ServeHTTP(ow, r)

		}
		return http.HandlerFunc(compressFn)
	}
}

func writeMetrics(p *filestorage.Producer) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "unable to read body", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			var metric models.Metrics
			if err := json.Unmarshal(body, &metric); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}

			h.ServeHTTP(w, r)

			if err := p.WriteMetric(metric); err != nil {
				fmt.Printf("failed to write metric: %v\n", err)
			}
		})
	}
}

func checkHash(key string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		hash := func(w http.ResponseWriter, r *http.Request) {
			gHash := r.Header.Get("HashSHA256")
			gBody, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "unable to read body", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(gBody))

			bHash := hash.GetHashHex(gBody, key)

			if !strings.EqualFold(gHash, bHash) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("HashSHA256", bHash)
			h.ServeHTTP(w, r)

		}
		return http.HandlerFunc(hash)
	}
}

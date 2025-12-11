package router

import (
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/klauspost/compress/zstd"
)

func compress(lvl int, types ...string) func(next http.Handler) http.Handler {
	comp := middleware.NewCompressor(lvl, types...)
	comp.SetEncoder("br", func(w io.Writer, lvl int) io.Writer {
		return brotli.NewWriterOptions(w, brotli.WriterOptions{
			Quality: lvl,
		})
	})
	comp.SetEncoder("zstd", func(w io.Writer, lvl int) io.Writer {
		writer, err := zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.EncoderLevel(lvl)))
		if err != nil {
			panic("Failed to create zstd writer")
		}
		return writer
	})

	return comp.Handler
}

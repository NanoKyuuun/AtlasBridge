package server

import (
	"log"
	"net/http"
)

type commitWriter struct {
	http.ResponseWriter
	committed bool
}

func (w *commitWriter) WriteHeader(code int) {
	if !w.committed {
		w.committed = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *commitWriter) Write(b []byte) (int, error) {
	if !w.committed {
		w.committed = true
	}
	return w.ResponseWriter.Write(b)
}

func (w *commitWriter) Committed() bool {
	return w.committed
}

func (w *commitWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func writeJSONAfterCommit(cw *commitWriter, status int, data interface{}) {
	if cw.Committed() {
		log.Printf("DROPPED response (headers already committed, status would have been %d)", status)
		return
	}
	writeJSON(cw.ResponseWriter, status, data)
}

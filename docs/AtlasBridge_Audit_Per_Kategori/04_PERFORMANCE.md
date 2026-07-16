# AtlasBridge — Performance

**Catatan:** File ini adalah pecahan laporan audit utama agar temuan dapat dikerjakan per kategori. Tidak ada perubahan source code yang dilakukan.

---

## 10. Temuan Detail — Performance

### PERF-01 — Frontend bundle masih aman, tetapi ada warning dynamic/static import

**Severity:** Low/Medium  
**Evidence:** hasil `npm run build`; warning Vite pada `web/src/stores/auth.ts`  
**Temuan:** Vite memperingatkan bahwa `auth.ts` di-dynamic import di beberapa file tetapi juga static import di `Layout.vue` dan `Login.vue`, sehingga tidak dipindah ke chunk terpisah.  
**Dampak:** bukan blocker, tetapi code-splitting tidak optimal/kurang bersih.  
**Rekomendasi audit:** konsistenkan import store; store kecil boleh static, tetapi hapus dynamic import yang tidak perlu.

### PERF-02 — Tidak ada field performance nyata di observability log

**Severity:** Medium  
**Evidence:** `internal/server/observability.go`, `internal/server/server.go` record observation  
**Temuan:** struktur log punya `latency_ms`, `status_code`, `bytes`, tetapi record utama tidak mengisi seluruh metrik final secara konsisten.  
**Dampak:** UI performance/observability tidak benar-benar mengukur performa request end-to-end.  
**Rekomendasi audit:** instrumentasi response writer untuk capture status, bytes, latency, route, dan error outcome.

### PERF-03 — Downstream request non-stream dibaca penuh ke memory

**Severity:** Medium  
**Evidence:** `internal/forwarder/forwarder.go:140-146`  
**Temuan:** response downstream non-stream dibaca dengan `io.ReadAll(io.LimitReader(...))`. Sudah ada limit, tetapi tetap full-buffer.  
**Dampak:** untuk response mendekati limit, memory spike per request dapat terjadi.  
**Rekomendasi audit:** pertimbangkan streaming copy dengan cap dan backpressure untuk non-stream juga.

### PERF-04 — Tidak ada load/perf test operasional yang tampak sebagai gate CI

**Severity:** Medium  
**Evidence:** `.github/workflows/ci.yml`, `internal/server/perf_test.go` hanya test unit/perf lokal  
**Temuan:** CI menjalankan Go test dan E2E, tetapi tidak ada benchmark threshold, smoke latency, atau load test sederhana.  
**Dampak:** regresi performa dapat lolos.  
**Rekomendasi audit:** tambahkan benchmark trend atau smoke test untuk request classification/forwarding.

---

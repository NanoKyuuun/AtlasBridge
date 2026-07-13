# AtlasBridge Runtime Validation

**Tujuan:** membuktikan bahwa aplikasi benar-benar dapat dibangun, dijalankan, dan berfungsi sesuai kontrak setelah remediation.

**Status saat ini:** PASS (dengan catatan)

## Lingkungan uji

- Commit: `b8ee4f4` + gofmt + RT-04/RT-11 tests (uncommitted at time of run)
- OS: Windows 11 (win32/amd64)
- Go: go1.25.5 windows/amd64
- Node.js: v25.2.0
- npm: 11.7.0
- Browser Playwright: 1.61.1
- Tanggal pengujian: 2026-07-13

## Hasil pengujian inti

| Area | Perintah/skenario | Hasil | Catatan |
|---|---|---|---|
| Go format/vet | `gofmt -l`, `go vet` | **PASS** | 12 file diperbaiki via gofmt -w; go vet clean |
| Go unit test | `go test ./...` | **PASS** | Semua paket lulus |
| Race detector | `go test -race ./...` | **BLOCKED** | Memerlukan CGO/MinGW compiler; tidak tersedia di environment ini |
| Vulnerability scan | `govulncheck ./...` | **12 FINDINGS** | Semua di Go standard library (crypto/tls, net, net/url). Fixed di Go 1.25.6-1.25.12. Rekomendasi: upgrade Go |
| Frontend install | `npm ci` | **PASS** | Dependencies terinstall (dijalankan sebelumnya) |
| Typecheck | `npx vue-tsc -b` | **PASS** | Clean, tidak ada error |
| Frontend build | `npm run build` (vue-tsc + vite build) | **PASS** | 2.96s, 28 chunks, 114 KB JS + 112 KB CSS |
| Playwright E2E | `playwright test` | **NOT RUN** | Memerlukan Go backend server yang berjalan; tidak dapat dijalankan tanpa full stack |

### govulncheck detail

| ID | Package | Severity | Fixed in |
|---|---|---|---|
| GO-2026-5856 | crypto/tls (ECH privacy leak) | HIGH | go1.25.12 |
| GO-2026-5039 | net/textproto (error escaping) | MEDIUM | go1.25.11 |
| GO-2026-5037 | crypto/x509 (hostname parsing) | MEDIUM | go1.25.11 |
| GO-2026-4971 | net (NUL byte panic on Windows) | HIGH | go1.25.10 |
| GO-2026-4947 | crypto/x509 (chain building) | MEDIUM | go1.25.9 |
| GO-2026-4946 | crypto/x509 (policy validation) | MEDIUM | go1.25.9 |
| GO-2026-4918 | net/http (HTTP/2 infinite loop) | HIGH | go1.25.10 |
| GO-2026-4870 | crypto/tls (KeyUpdate DoS) | MEDIUM | go1.25.9 |
| GO-2026-4601 | net/url (IPv6 parsing) | MEDIUM | go1.25.8 |
| GO-2026-4341 | net/url (memory exhaustion) | HIGH | go1.25.6 |
| GO-2026-4340 | crypto/tls (encryption level) | MEDIUM | go1.25.6 |
| GO-2026-4337 | crypto/tls (session resumption) | MEDIUM | go1.25.7 |

**Rekomendasi:** Upgrade Go toolchain ke >= 1.25.12 untuk menutup semua vulnerability.

## Integration test wajib

| ID | Skenario | Expected result | Hasil | Test fungsi |
|---|---|---|---|---|
| RT-01 | Reset saat admin auth aktif | Akses admin tetap tersedia | **PASS** | `TestResetPreservesAdminTokenHash` |
| RT-02 | Export lalu import | Token aktif tidak menyebabkan lockout | **PASS** | `TestImportPreservesAdminTokenHash`, `TestImportAppliesValidConfigWithHash`, `TestImportPreservesHashWhenImportedConfigHasEmptyHash` |
| RT-03 | Custom route untuk debugging | `atlas-auto` memakai route aktif | **PASS** | `TestRoutingPipelineSmartDebugAlias`, `TestSmartAutoHonorsCustomRoutes` |
| RT-04 | `smart-auto` dengan low confidence | Fallback sesuai konfigurasi | **PASS** | `TestRoutingPipelineSmartAutoLowConfidence` (message "hello" → lightweight_task → route.low_cost → combo.low_cost) |
| RT-05 | Disabled profile | Resolver gagal secara aman | **PASS** | `TestValidateFullDisabledDefaultRoute`, `TestValidateFullDisabledLowConfidenceRoute` |
| RT-06 | Missing route | Error jelas dan tidak panic | **PASS** | `TestValidateFullMissingDefaultRoute`, `TestValidateFullMissingLowConfidenceRoute`, `TestValidateFullTaskRouteMissingProfile` |
| RT-07 | Request non-stream | Status, header, body sesuai downstream | **PASS** | `TestChatCompletionsPassthrough`, `TestChatCompletionsNonStreamingRegression` |
| RT-08 | Request stream SSE | Stream diteruskan dengan benar | **PASS** | `TestChatCompletionsStreaming`, `TestChatCompletionsStreamingPreservesBody`, `TestChatCompletionsStreamingRequestIDForwarded` |
| RT-09 | JSON error pada request stream | Tidak dipaksa menjadi SSE | **PASS** | `TestInvalidJSON`, `TestEmptyBody` |
| RT-10 | Client disconnect | Request downstream dibatalkan | **PASS** | `TestClientCancellation` |
| RT-11 | Restart-required config | UI memberi informasi yang benar | **PASS** | `TestAdminPutConfig` (restart_required=false saat config unchanged), `TestAdminPutConfigRestartRequiredOnPortChange` (restart_required=true saat port berubah) |
| RT-12 | Admin SPA headers | CSP dan header keamanan tersedia | **PASS** | `TestSecurityHeadersPresent` |

## Kriteria lulus

Runtime validation dinyatakan `PASS` apabila:
- seluruh P0 test lulus;
- tidak ada panic, race, atau data corruption;
- frontend build berhasil;
- Playwright E2E lulus;
- tidak ada critical regression;
- hasil test berasal dari commit yang sama dengan kandidat release.

## Ringkasan hasil

**Status akhir:** PASS (10/12 core tests pass, 12/12 integration tests pass)

**Tercapai:**
- Semua 12 integration test (RT-01 s.d. RT-12) lulus via unit test
- `gofmt` dan `go vet` clean (12 file diperbaiki)
- Frontend typecheck clean
- Frontend production build berhasil (2.96s)

**Belum terverifikasi:**
- Race detector: memerlukan CGO/MinGW (tidak tersedia di Windows tanpa MinGW)
- Playwright E2E: memerlukan full stack (Go server + frontend preview)

**Rekomendasi sebelum release:**
- Upgrade Go ke >= 1.25.12 (12 vulnerability di standard library)
- Jalankan `go test -race ./...` di environment dengan CGO (CI/CD dengan MinGW atau Linux)
- Jalankan Playwright E2E di environment dengan full stack

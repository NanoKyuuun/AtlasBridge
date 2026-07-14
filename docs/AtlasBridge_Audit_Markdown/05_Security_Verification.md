# AtlasBridge Security Verification

**Tujuan:** memverifikasi kontrol keamanan secara dinamis setelah perbaikan P0 selesai.

**Status saat ini:** Belum dijalankan

## A. Admin authentication

| ID | Skenario | Expected result | Hasil |
|---|---|---|---|
| SEC-01 | Token admin valid | Request diterima | NOT RUN |
| SEC-02 | Token salah | `401` | NOT RUN |
| SEC-03 | Token kosong | `401` | NOT RUN |
| SEC-04 | Reset config | Tidak terjadi lockout | NOT RUN |
| SEC-05 | Export-import | Credential tetap valid atau dirotasi aman | NOT RUN |
| SEC-06 | Rotasi token | Token lama dicabut | NOT RUN |

## B. Data-plane authentication dan LAN

| ID | Skenario | Expected result | Hasil |
|---|---|---|---|
| SEC-07 | LAN tanpa data-plane token | Ditolak | NOT RUN |
| SEC-08 | LAN dengan token valid | Diterima | NOT RUN |
| SEC-09 | Token client dicabut | Ditolak | NOT RUN |
| SEC-10 | Banyak request satu client | Concurrency limit bekerja (WeightedBulkhead) | NOT RUN |
| SEC-11 | Banyak stream satu client | Concurrency limit bekerja | NOT RUN |

## C. SSRF dan outbound policy

| ID | Skenario | Expected result | Hasil |
|---|---|---|---|
| SEC-12 | Loopback 9Router yang diizinkan | Diterima | NOT RUN |
| SEC-13 | Metadata IP | Ditolak | NOT RUN |
| SEC-14 | IPv4 private tidak diizinkan | Ditolak | NOT RUN |
| SEC-15 | IPv6 loopback | Diterima (loopback diizinkan) | NOT RUN |
| SEC-15b | IPv6 link-local/unique local | Ditolak sesuai policy (`fe80::/10`, `fc00::/7`) | NOT RUN |
| SEC-16 | IPv6 global yang diizinkan | Diterima sesuai policy | NOT RUN |
| SEC-17 | Redirect ke blocked IP | Ditolak | NOT RUN |
| SEC-18 | Host dengan beberapa A/AAAA record | ⚠️ Partial: `ResolveAndValidateIP` ada tapi belum dipanggil saat forwarding | NOT RUN |
| SEC-19 | DNS re-resolution/rebinding | Tidak melewati policy | NOT RUN |
| SEC-20 | Credential/query/fragment pada URL | Ditolak | NOT RUN |

## D. Web security

| ID | Skenario | Expected result | Hasil |
|---|---|---|---|
| SEC-21 | Admin HTML | CSP tersedia | NOT RUN |
| SEC-22 | Admin HTML | `nosniff` tersedia | NOT RUN |
| SEC-23 | Admin HTML | frame protection tersedia | NOT RUN |
| SEC-24 | Admin HTML | `Referrer-Policy` tersedia | NOT RUN |
| SEC-25 | Admin HTML | ❌ NOT IMPLEMENTED: X-Robots-Tag belum tersedia | NOT RUN |
| SEC-26 | Token di browser | Tidak muncul di URL/log | NOT RUN |

## Kriteria lulus

Security verification dinyatakan `PASS` apabila:
- seluruh pengujian P0 lulus;
- tidak ada bypass autentikasi;
- seluruh outbound request melewati network policy;
- redirect dan DNS tidak dapat melewati policy;
- LAN tidak dapat digunakan tanpa credential;
- tidak ada critical/high finding terbuka.

**Catatan known gaps:**
- SEC-18: `ResolveAndValidateIP` belum dipanggil saat forwarding (DNS validation hanya di URL construction)
- SEC-25: X-Robots-Tag/noindex belum diimplementasi (medium risk, hanya admin UI)

## Keputusan per deployment

| Deployment | Status |
|---|---|
| Localhost single-user | PENDING |
| Trusted LAN | NO-GO sampai seluruh test LAN lulus |
| Remote downstream | NO-GO sampai seluruh test SSRF lulus |
| Multi-user/enterprise | OUT OF SCOPE untuk versi saat ini |

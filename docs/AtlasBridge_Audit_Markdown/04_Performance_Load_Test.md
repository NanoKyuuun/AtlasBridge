# AtlasBridge Performance and Load Test

**Tujuan:** mengukur overhead proxy, penggunaan memory, goroutine, streaming, dan perilaku saat beban meningkat.

**Status saat ini:** Belum dijalankan

## Lingkungan benchmark

- Commit:
- OS:
- CPU:
- RAM:
- Go:
- Mode build:
- Downstream mock:
- Durasi test:

## Skenario minimum

| ID | Skenario | Beban |
|---|---|---|
| PF-01 | Non-stream request kecil | 1× |
| PF-02 | Non-stream request kecil | 5× |
| PF-03 | Non-stream request kecil | 10× |
| PF-04 | Request 1 MiB | concurrency bertahap |
| PF-05 | Request mendekati body limit | concurrency rendah |
| PF-06 | Streaming panjang | beberapa stream aktif |
| PF-07 | Downstream lambat | latency buatan |
| PF-08 | Downstream tidak merespons | timeout verification |
| PF-09 | Client disconnect saat stream | cleanup verification |
| PF-10 | Beban berulang | memory/goroutine leak check |

## Metrik

| Metrik | Target awal | Hasil |
|---|---:|---|
| Routing computation p95 | < 5 ms | NOT TESTED |
| Proxy overhead p95 non-stream | < 25 ms di luar latency downstream | NOT TESTED |
| Internal error rate | < 0,1% | NOT TESTED |
| Bulkhead rejection pada normal load | 0% | NOT TESTED |
| Memory setelah test | kembali mendekati baseline | NOT TESTED |
| Goroutine setelah stream selesai | tidak terus meningkat | NOT TESTED |

> Target di atas adalah sasaran awal, bukan hasil audit.

## Hal yang wajib diperiksa

- Request body hanya dibaca satu kali.
- Response non-stream memiliki size limit.
- Stream memiliki idle timeout.
- Maximum stream lifetime dapat dikonfigurasi.
- Client disconnect membatalkan downstream request.
- Heap tidak tumbuh terus setelah beban berhenti.
- Goroutine tidak bocor.
- Bulkhead bekerja sesuai batas.

## Keputusan

| Kondisi | Status |
|---|---|
| Belum ada benchmark | NOT READY |
| Semua target awal tercapai | PASS |
| Ada memory/goroutine leak | FAIL |
| Error meningkat tajam pada 10× | FAIL atau perlu optimasi |

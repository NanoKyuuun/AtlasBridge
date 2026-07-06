# Blueprint Project: AtlasBridge

## 1. Executive Summary

AtlasBridge adalah OpenAI-compatible proxy server yang berada di antara AI coding assistant seperti OpenCode, Cursor, Cline, Continue, atau aplikasi lain yang kompatibel dengan OpenAI API, dan 9Router sebagai downstream router.

Tujuan utama AtlasBridge adalah menganalisis isi request, mengklasifikasikan jenis pekerjaan, lalu memilih AI combo atau routing tujuan yang paling sesuai sebelum request diteruskan ke 9Router. AtlasBridge tidak menggantikan peran 9Router. Seluruh fungsi seperti failover, load balancing, rotasi akun, fallback model, dan rate limit handling tetap menjadi tanggung jawab 9Router.

Dengan pendekatan ini, developer tidak perlu terus-menerus mengganti model AI secara manual. Sistem dapat memilih model atau kombinasi model yang paling relevan berdasarkan konteks pekerjaan, seperti coding, debugging, refactoring, reasoning, documentation, testing, code review, planning, atau task ringan.

---

## 2. Project Vision

Visi AtlasBridge adalah menjadi intelligent routing layer untuk ekosistem AI coding assistant yang mampu menghubungkan developer dengan model AI terbaik untuk setiap jenis pekerjaan secara otomatis, efisien, dan tetap kompatibel dengan standar OpenAI API.

AtlasBridge dirancang untuk menjadi lapisan pengambil keputusan, bukan lapisan eksekusi model. Sistem ini harus ringan, fleksibel, mudah diintegrasikan, dan dapat digunakan bersama berbagai router, terutama 9Router.

Dalam jangka panjang, AtlasBridge dapat berkembang menjadi AI orchestration layer yang mampu memahami konteks proyek, kebiasaan developer, performa model, biaya penggunaan, kualitas respons, serta strategi routing berbasis evaluasi historis.

---

## 3. Problem Statement

Developer modern sering menggunakan banyak model AI untuk kebutuhan coding, reasoning, debugging, review, dokumentasi, dan eksperimen. Setiap model memiliki keunggulan berbeda. Ada model yang lebih kuat untuk reasoning, ada yang lebih baik untuk coding cepat, ada yang murah untuk task ringan, dan ada yang cocok untuk konteks panjang.

Masalahnya, pemilihan model biasanya masih dilakukan secara manual. Developer harus mengetahui kekuatan masing-masing model, mengganti konfigurasi secara berkala, dan menyesuaikan model dengan jenis pekerjaan. Proses ini tidak efisien, rawan salah pilih model, dan mengganggu workflow.

AtlasBridge menyelesaikan masalah ini dengan menyediakan intelligent routing layer yang otomatis menganalisis request dan memilih jalur model paling relevan sebelum diteruskan ke 9Router.

---

## 4. Existing Problems

### 4.1 Manual Model Switching

Developer harus mengganti model secara manual berdasarkan kebutuhan. Misalnya menggunakan model tertentu untuk debugging, model lain untuk planning, dan model lain untuk generate dokumentasi. Hal ini memperlambat workflow.

### 4.2 Model Capability Awareness Rendah

Tidak semua developer mengetahui model mana yang paling baik untuk task tertentu. Akibatnya, model yang digunakan sering tidak optimal.

### 4.3 Overuse Model Mahal atau Berat

Task sederhana seperti formatting, penjelasan pendek, atau generate komentar kode sering dikirim ke model besar yang sebenarnya tidak diperlukan. Ini dapat meningkatkan biaya dan memperlambat respons.

### 4.4 Underuse Model Kuat

Sebaliknya, task kompleks seperti architectural reasoning, debugging mendalam, atau multi-file refactoring terkadang dikirim ke model yang kurang mampu, sehingga hasilnya tidak maksimal.

### 4.5 Tidak Ada Task-Aware Routing

Router seperti 9Router fokus pada failover, load balancing, rotasi akun, fallback model, dan rate limit handling. Namun routing berbasis jenis pekerjaan atau intent request membutuhkan lapisan terpisah.

### 4.6 Fragmentasi Tool AI Coding

OpenCode, Cursor, Cline, Continue, dan tool sejenis menggunakan format OpenAI-compatible API, tetapi tidak selalu memiliki intelligence layer untuk memilih model terbaik berdasarkan task.

---

## 5. Proposed Solution

AtlasBridge menjadi proxy OpenAI-compatible yang menerima request dari AI coding assistant, menganalisis request tersebut, menentukan klasifikasi task, memilih routing target atau AI combo, lalu meneruskan request ke 9Router.

AtlasBridge tidak memanggil provider AI secara langsung. Semua request tetap diteruskan ke 9Router. Dengan demikian, AtlasBridge tetap memiliki tanggung jawab yang jelas dan tidak mengambil alih fungsi downstream router.

Tanggung jawab utama AtlasBridge:

- Menerima request OpenAI-compatible.
- Mengekstrak metadata request.
- Menganalisis prompt, system message, user message, tool call, atau context.
- Mengklasifikasikan jenis pekerjaan.
- Menentukan routing intent.
- Memilih AI combo atau target routing profile.
- Menambahkan metadata routing yang dapat dipahami oleh 9Router.
- Meneruskan request ke 9Router.
- Mencatat telemetry dasar untuk analisis performa routing.

Tanggung jawab yang tetap berada di 9Router:

- Failover.
- Load balancing.
- Rotasi akun.
- Fallback model.
- Rate limit handling.
- Provider abstraction.
- Provider credential management.
- Model availability handling.

---

## 6. High Level Architecture

### 6.1 Logical Architecture

```text
AI Coding Assistant
(OpenCode, Cursor, Cline, Continue, etc.)
        ↓
OpenAI-Compatible Request
        ↓
AtlasBridge
        ↓
Request Analyzer
        ↓
Task Classifier
        ↓
Routing Policy Engine
        ↓
AI Combo Selector
        ↓
9Router-Compatible Forwarder
        ↓
9Router
        ↓
Multiple AI Providers
```

### 6.2 Responsibility Boundary

AtlasBridge berada pada lapisan decision-making. 9Router berada pada lapisan execution-routing.

AtlasBridge menjawab pertanyaan:

> “Jenis pekerjaan apa ini, dan model atau combo apa yang paling cocok?”

9Router menjawab pertanyaan:

> “Bagaimana request ini dikirim secara andal ke provider yang tersedia?”

### 6.3 Main Components

#### API Gateway Layer

Menerima request dari client yang kompatibel dengan OpenAI API. Layer ini bertugas menjaga kompatibilitas endpoint seperti chat completion, streaming, headers, authentication forwarding, dan format response.

#### Request Analyzer

Menganalisis struktur request, termasuk system prompt, user prompt, message history, model yang diminta, temperature, max tokens, tools, dan metadata tambahan.

#### Task Classifier

Mengklasifikasikan request ke dalam kategori pekerjaan. Contoh kategori:

- Code generation.
- Debugging.
- Refactoring.
- Test generation.
- Documentation.
- Code review.
- Architecture planning.
- Multi-step reasoning.
- Shell command assistance.
- Simple Q&A.
- Translation or rewriting.
- Long-context analysis.
- Security-sensitive review.

#### Routing Policy Engine

Menentukan aturan routing berdasarkan hasil klasifikasi, konfigurasi user, prioritas biaya, performa, latency, dan preferensi model.

#### AI Combo Selector

Memilih model target, model group, atau combo routing. Contoh:

- Fast coding combo.
- Deep reasoning combo.
- Cheap/simple task combo.
- Long context combo.
- Documentation combo.
- Debugging combo.
- Security review combo.

#### Forwarder to 9Router

Meneruskan request ke 9Router dengan format OpenAI-compatible. Forwarder dapat menambahkan metadata routing melalui header, model alias, routing profile, atau mekanisme konfigurasi yang disepakati.

#### Observability Layer

Mencatat informasi non-sensitif seperti task type, selected route, latency, error class, dan routing decision untuk evaluasi sistem.

---

## 7. Core Concepts

### 7.1 OpenAI Compatibility

AtlasBridge harus mempertahankan kompatibilitas dengan OpenAI API agar dapat digunakan oleh OpenCode, Cursor, Cline, Continue, dan aplikasi lain tanpa perubahan besar.

Compatibility harus mencakup:

- Request format.
- Response format.
- Streaming behavior.
- Error format.
- Authentication forwarding.
- Model naming convention.
- Tool call compatibility.

### 7.2 Task Classification

Task classification adalah proses memahami tujuan request. Sistem tidak hanya melihat nama model yang diminta, tetapi juga isi prompt, konteks, dan pola instruksi.

Contoh klasifikasi:

- “Fix this error” → debugging.
- “Refactor this component” → refactoring.
- “Explain this code” → code explanation.
- “Create unit tests” → test generation.
- “Design architecture” → system design.
- “Review for security issues” → security review.

### 7.3 Routing Policy

Routing policy adalah kumpulan aturan yang menentukan model atau combo berdasarkan klasifikasi task.

Contoh policy:

- Task ringan → model cepat dan murah.
- Task coding umum → coding-optimized model.
- Task reasoning kompleks → reasoning-optimized model.
- Task dengan konteks panjang → long-context model.
- Task security review → model dengan reasoning kuat dan kecenderungan analitis.
- Task dokumentasi → model dengan kemampuan writing kuat.

### 7.4 AI Combo

AI Combo adalah konsep abstraksi routing yang merepresentasikan satu atau beberapa pilihan model yang cocok untuk jenis pekerjaan tertentu.

Contoh AI Combo:

- `combo.fast_coding`
- `combo.deep_reasoning`
- `combo.debugging`
- `combo.refactor`
- `combo.documentation`
- `combo.long_context`
- `combo.low_cost`
- `combo.security_review`

AtlasBridge memilih combo. 9Router kemudian menjalankan detail provider, fallback, failover, dan routing aktual.

### 7.5 Model Alias

Model alias adalah nama model virtual yang dapat dimengerti oleh AtlasBridge dan/atau 9Router.

Contoh:

- `smart-auto`
- `smart-code`
- `smart-debug`
- `smart-architect`
- `smart-cheap`
- `smart-long-context`

Client dapat menggunakan satu model seperti `smart-auto`, lalu AtlasBridge menentukan route aktual.

### 7.6 Policy-Driven Routing

Routing tidak boleh hardcoded secara permanen. Semua aturan idealnya dapat dikonfigurasi melalui file, dashboard, database, atau environment configuration.

Policy perlu mendukung:

- Default route.
- Task-based route.
- User-specific preference.
- Project-specific rule.
- Cost priority.
- Latency priority.
- Quality priority.
- Manual override.
- Safe fallback behavior.

---

## 8. User Flow

### 8.1 Basic User Flow

1. Developer mengatur AI coding assistant agar menggunakan endpoint AtlasBridge.
2. Developer memilih model virtual, misalnya `smart-auto`.
3. Developer mengirim prompt dari OpenCode, Cursor, Cline, Continue, atau tool lainnya.
4. AtlasBridge menerima request.
5. AtlasBridge menganalisis isi request.
6. AtlasBridge mengklasifikasikan jenis task.
7. AtlasBridge memilih AI combo atau routing profile.
8. AtlasBridge meneruskan request ke 9Router.
9. 9Router menangani failover, load balancing, fallback, provider, dan rate limit.
10. Provider AI menghasilkan response.
11. Response dikembalikan melalui 9Router ke AtlasBridge.
12. AtlasBridge meneruskan response ke client tanpa mengubah format utama.

### 8.2 Example Flow: Debugging

Developer mengirim prompt:

> “Tolong cari penyebab error pada fungsi ini dan perbaiki.”

AtlasBridge mendeteksi:

- Intent: debugging.
- Context: source code.
- Complexity: medium.
- Recommended route: debugging combo.

AtlasBridge meneruskan request ke 9Router dengan routing target:

- `combo.debugging`

9Router menentukan provider dan model aktual berdasarkan konfigurasi internalnya.

### 8.3 Example Flow: Simple Documentation

Developer mengirim prompt:

> “Buatkan komentar singkat untuk fungsi ini.”

AtlasBridge mendeteksi:

- Intent: documentation.
- Complexity: low.
- Recommended route: low-cost documentation combo.

AtlasBridge meneruskan request ke 9Router dengan routing target:

- `combo.documentation_fast` atau `combo.low_cost`

### 8.4 Example Flow: Architecture Design

Developer mengirim prompt:

> “Rancang arsitektur backend untuk sistem multi-tenant dengan audit log dan role-based access.”

AtlasBridge mendeteksi:

- Intent: architecture planning.
- Complexity: high.
- Needs reasoning: high.
- Recommended route: deep reasoning combo.

AtlasBridge meneruskan request ke 9Router dengan routing target:

- `combo.deep_reasoning` atau `combo.architect`

---

## 9. Functional Scope

### 9.1 OpenAI-Compatible API Proxy

AtlasBridge harus dapat menerima request dari client yang menggunakan OpenAI-compatible API.

Fitur utama:

- Endpoint chat completion.
- Streaming response.
- Non-streaming response.
- Header forwarding.
- Request body forwarding.
- Error forwarding.
- Model alias support.

### 9.2 Request Analysis

Sistem harus dapat membaca dan menganalisis request sebelum diteruskan.

Analisis mencakup:

- System message.
- User message.
- Assistant message history.
- Tool definitions.
- Function calling intent.
- Code block presence.
- File/context size.
- Prompt length.
- Task keywords.
- Requested output format.
- Complexity signal.

### 9.3 Task Classification

Sistem harus mengklasifikasikan request ke task type tertentu.

Minimum task type:

- General chat.
- Coding.
- Debugging.
- Refactoring.
- Code explanation.
- Test generation.
- Documentation.
- Code review.
- Architecture design.
- Planning.
- Security review.
- Long-context analysis.
- Data transformation.
- Lightweight task.

### 9.4 Routing Decision

Sistem harus memilih routing target berdasarkan task classification dan policy.

Routing decision dapat berbentuk:

- Model alias.
- AI combo name.
- Header metadata.
- Route profile.
- Custom downstream parameter.
- Original model passthrough.

### 9.5 Manual Override

User harus dapat melakukan override routing.

Contoh:

- Memaksa route tertentu.
- Menonaktifkan auto-routing.
- Mengunci model tertentu.
- Memilih mode murah.
- Memilih mode kualitas tinggi.
- Memilih mode cepat.

### 9.6 Policy Configuration

Routing policy harus dapat dikonfigurasi tanpa mengubah kode inti.

Contoh konfigurasi:

- Default combo.
- Mapping task ke combo.
- Cost preference.
- Latency preference.
- Quality preference.
- Project-specific rules.
- User-specific rules.

### 9.7 Logging and Observability

AtlasBridge perlu mencatat informasi teknis untuk evaluasi.

Contoh data observability:

- Request ID.
- Timestamp.
- Client type.
- Task classification.
- Selected route.
- Routing reason.
- Latency.
- Error status.
- Token estimation.
- Streaming or non-streaming mode.

Data sensitif seperti source code, secrets, API key, dan prompt penuh sebaiknya tidak disimpan secara default.

### 9.8 Safe Passthrough Mode

Jika AtlasBridge gagal menganalisis request, sistem harus tetap dapat meneruskan request ke 9Router menggunakan default route.

Prinsip penting:

> Routing intelligence boleh gagal, tetapi request flow tidak boleh berhenti tanpa alasan yang jelas.

---

## 10. Non Functional Scope

### 10.1 Performance

AtlasBridge harus menambahkan latency seminimal mungkin. Karena proxy ini berada di jalur utama request AI, proses klasifikasi harus cepat dan efisien.

Target ideal:

- Classification overhead rendah.
- Tidak memanggil LLM tambahan untuk setiap request pada versi awal.
- Menggunakan rule-based atau lightweight classifier sebagai baseline.
- Mendukung caching keputusan untuk pola request serupa.

### 10.2 Reliability

Sistem harus tetap stabil meskipun classifier gagal, policy invalid, atau downstream 9Router mengalami error.

Reliability requirement:

- Safe fallback.
- Request timeout handling.
- Error propagation.
- Graceful degradation.
- Health check endpoint.
- Retry terbatas hanya jika sesuai boundary tanggung jawab.

Catatan penting: retry terhadap provider tidak menjadi tanggung jawab utama AtlasBridge karena itu masuk wilayah 9Router.

### 10.3 Compatibility

AtlasBridge harus menjaga kompatibilitas tinggi dengan OpenAI-compatible clients.

Compatibility requirement:

- Tidak merusak request body.
- Tidak mengubah response format secara tidak perlu.
- Mendukung streaming.
- Mendukung tool calling.
- Mendukung model alias.
- Mendukung header forwarding.

### 10.4 Security

AtlasBridge harus dirancang dengan prinsip secure-by-default.

Requirement:

- Tidak menyimpan API key dalam log.
- Tidak menyimpan prompt penuh secara default.
- Mendukung redaction untuk secrets.
- Mendukung allowlist downstream endpoint.
- Mendukung authentication layer.
- Mendukung rate control internal jika diperlukan untuk melindungi proxy.
- Mendukung audit log non-sensitif.

### 10.5 Privacy

Karena request dapat berisi source code, data bisnis, credential, atau informasi sensitif, AtlasBridge perlu membatasi penyimpanan data.

Privacy principle:

- Log metadata, bukan isi prompt.
- Prompt sampling hanya jika user mengaktifkan.
- Redaction sebelum logging.
- Konfigurasi retention period.
- Mode private untuk enterprise atau local deployment.

### 10.6 Scalability

Sistem harus dapat berjalan dari skala personal developer hingga tim kecil atau organisasi.

Scalability consideration:

- Stateless proxy design.
- Horizontal scaling.
- Externalized policy storage.
- Centralized observability.
- Lightweight routing engine.
- Low memory footprint.

### 10.7 Maintainability

Sistem harus mudah dikembangkan, diuji, dan dikonfigurasi.

Maintainability requirement:

- Modular architecture.
- Clear responsibility boundary.
- Policy engine terpisah dari API layer.
- Classifier dapat diganti.
- Routing strategy dapat ditambah.
- Observability tidak bercampur dengan business logic.

---

## 11. Future Expansion

### 11.1 Learning-Based Routing

Pada tahap lanjutan, AtlasBridge dapat belajar dari hasil historis untuk meningkatkan keputusan routing.

Contoh data:

- Response latency.
- Error rate.
- User retry.
- Manual override.
- Acceptance signal.
- Cost estimation.
- Task success feedback.

### 11.2 Project-Aware Routing

AtlasBridge dapat memahami konteks proyek tertentu.

Contoh:

- Bahasa pemrograman utama.
- Framework yang digunakan.
- Struktur repository.
- Coding style.
- Testing framework.
- Preferred architecture.
- Project rules.

### 11.3 Developer Profile

Sistem dapat mendukung preferensi per developer.

Contoh:

- Prefer fast response.
- Prefer high quality.
- Prefer low cost.
- Prefer open-source models.
- Prefer paid premium models.
- Prefer local models.

### 11.4 Multi-Agent Preprocessing

Pada versi lanjutan, AtlasBridge dapat melakukan preprocessing ringan sebelum request dikirim ke 9Router.

Namun fitur ini harus hati-hati agar tidak mengambil alih fungsi utama client atau 9Router.

Contoh:

- Prompt normalization.
- Context trimming.
- Sensitive data redaction.
- Task summary generation.
- Model-specific prompt adaptation.

### 11.5 Admin Dashboard

Dashboard dapat digunakan untuk melihat:

- Routing statistics.
- Task distribution.
- Model usage.
- Cost trend.
- Latency trend.
- Error trend.
- Policy configuration.
- User/project settings.

### 11.6 Evaluation Framework

Sistem dapat memiliki benchmark internal untuk menguji kualitas routing.

Contoh:

- Apakah debugging task masuk ke debugging combo?
- Apakah lightweight task masuk ke low-cost combo?
- Apakah architecture task masuk ke reasoning combo?
- Apakah long-context task masuk ke long-context combo?

### 11.7 Cost-Aware Optimization

Routing dapat mempertimbangkan estimasi biaya.

Contoh:

- Gunakan model murah untuk task ringan.
- Gunakan model premium hanya untuk task kompleks.
- Gunakan long-context model hanya jika prompt memang panjang.
- Gunakan routing berbeda berdasarkan budget harian.

### 11.8 Team and Enterprise Mode

Untuk tim atau organisasi, AtlasBridge dapat dikembangkan dengan fitur:

- Shared policy.
- Role-based access.
- Team-level budget.
- Audit log.
- Centralized API key management.
- Environment-based routing.
- Workspace isolation.

---

## 12. Success Metrics

### 12.1 Product Metrics

- Penurunan kebutuhan manual model switching.
- Peningkatan kepuasan developer.
- Jumlah request yang berhasil diklasifikasikan.
- Jumlah override manual oleh user.
- Jumlah client/tool yang berhasil terintegrasi.
- Retention pengguna aktif.

### 12.2 Technical Metrics

- Routing decision latency.
- Total proxy overhead latency.
- Classification accuracy.
- Error rate.
- Streaming compatibility success rate.
- Passthrough success rate.
- Policy resolution success rate.

### 12.3 Cost Metrics

- Penurunan penggunaan model mahal untuk task ringan.
- Rasio task ringan yang diarahkan ke low-cost combo.
- Estimasi penghematan biaya.
- Distribusi penggunaan model berdasarkan task type.

### 12.4 Quality Metrics

- User acceptance rate terhadap hasil model.
- Retry rate setelah response buruk.
- Manual reroute frequency.
- Task completion success.
- Feedback score per combo.

---

## 13. Risks

### 13.1 Salah Klasifikasi Task

AtlasBridge dapat salah mengklasifikasikan request, sehingga memilih model atau combo yang kurang tepat.

Mitigasi:

- Gunakan safe default route.
- Sediakan manual override.
- Simpan metadata routing untuk evaluasi.
- Gunakan confidence score.
- Gunakan fallback policy untuk request ambigu.

### 13.2 Latency Tambahan

Karena AtlasBridge berada di jalur request, proses analisis dapat menambah latency.

Mitigasi:

- Gunakan classifier ringan.
- Hindari LLM classification untuk setiap request pada fase awal.
- Gunakan caching.
- Optimalkan parsing request.

### 13.3 Compatibility Issue

Beberapa OpenAI-compatible client memiliki variasi implementasi. Jika proxy terlalu banyak mengubah request atau response, integrasi dapat terganggu.

Mitigasi:

- Pertahankan passthrough behavior.
- Minimalkan perubahan response.
- Buat compatibility test untuk OpenCode, Cursor, Cline, Continue.
- Dukung streaming secara benar sejak awal.

### 13.4 Boundary Creep dengan 9Router

Ada risiko AtlasBridge berkembang terlalu jauh dan mulai mengambil alih fungsi 9Router.

Mitigasi:

- Tetapkan responsibility boundary secara eksplisit.
- Hindari implementasi provider failover di AtlasBridge.
- Hindari provider credential management.
- Hindari fallback model internal yang bertabrakan dengan 9Router.

### 13.5 Security and Privacy Risk

Request dapat berisi source code, secrets, token, API key, atau data rahasia.

Mitigasi:

- Redact sensitive data.
- Jangan log prompt penuh secara default.
- Batasi telemetry ke metadata.
- Sediakan private mode.
- Gunakan secure configuration management.

### 13.6 Policy Misconfiguration

Routing policy yang salah dapat menyebabkan request diarahkan ke combo yang tidak sesuai.

Mitigasi:

- Validasi policy saat startup.
- Gunakan schema untuk konfigurasi.
- Sediakan default policy.
- Sediakan dry-run mode.
- Catat alasan routing decision.

---

## 14. Assumptions

- Client menggunakan OpenAI-compatible API.
- 9Router dapat menerima request OpenAI-compatible dari AtlasBridge.
- 9Router tetap menjadi layer utama untuk provider routing dan reliability.
- AtlasBridge tidak perlu menyimpan prompt penuh untuk menjalankan fungsi dasarnya.
- Model atau combo dapat direpresentasikan melalui model alias, header, atau konfigurasi yang disepakati.
- Developer bersedia menggunakan model virtual seperti `smart-auto`.
- Task classification awal dapat dilakukan menggunakan rule-based atau heuristic-based system.
- Kebutuhan utama pengguna adalah routing otomatis, bukan mengganti keseluruhan AI workflow.
- Sistem perlu mendukung penggunaan personal, tim kecil, dan kemungkinan enterprise di masa depan.

---

## 15. Constraints

### 15.1 Tidak Menggantikan 9Router

AtlasBridge tidak boleh mengambil alih fungsi:

- Failover provider.
- Load balancing.
- Rotasi akun.
- Fallback model.
- Rate limit handling.
- Credential provider.
- Provider abstraction utama.

### 15.2 Harus OpenAI-Compatible

Sistem harus tetap kompatibel dengan OpenAI API agar mudah digunakan oleh tool existing.

### 15.3 Minimal Latency

Karena digunakan dalam coding workflow, latency tambahan harus sangat rendah.

### 15.4 Tidak Bergantung pada LLM untuk Semua Klasifikasi

Menggunakan LLM untuk mengklasifikasi setiap request dapat mahal dan lambat. Versi awal sebaiknya menggunakan rule-based, heuristic, atau lightweight classifier.

### 15.5 Privacy by Default

Prompt dan source code tidak boleh disimpan secara penuh kecuali user secara eksplisit mengaktifkan fitur tersebut.

### 15.6 Konfigurasi Harus Fleksibel

Karena model AI dan provider sering berubah, routing policy tidak boleh terkunci pada satu daftar model tertentu.

### 15.7 Harus Mendukung Streaming

Banyak AI coding assistant membutuhkan streaming response. AtlasBridge harus meneruskan streaming dengan benar dan tidak merusak format chunk.

---

## 16. Technical Challenges

### 16.1 Accurate Task Classification

Membedakan task coding, debugging, refactoring, review, documentation, dan reasoning tidak selalu mudah. Prompt developer sering pendek, ambigu, atau bercampur beberapa intent.

Contoh:

> “Fix this and make it cleaner.”

Prompt tersebut dapat berarti debugging sekaligus refactoring.

Solusi yang dibutuhkan:

- Multi-label classification.
- Confidence score.
- Priority rule.
- Fallback route.
- Manual override.

### 16.2 Routing Without Breaking API Compatibility

AtlasBridge harus menambahkan intelligence tanpa merusak request atau response OpenAI-compatible.

Tantangan:

- Streaming chunk forwarding.
- Tool call preservation.
- Function calling compatibility.
- Error format compatibility.
- Header propagation.
- Timeout behavior.

### 16.3 Designing the AI Combo Abstraction

AI Combo harus cukup fleksibel untuk merepresentasikan routing strategy, tetapi tidak terlalu kompleks sampai meniru fungsi 9Router.

AI Combo harus fokus pada intent, bukan provider execution.

Contoh yang benar:

- `combo.debugging`
- `combo.deep_reasoning`
- `combo.low_cost`

Contoh yang perlu dihindari di AtlasBridge:

- Menentukan akun provider spesifik.
- Melakukan rotasi provider.
- Mengelola fallback provider secara langsung.

### 16.4 Policy Engine Complexity

Routing policy dapat menjadi kompleks jika harus mempertimbangkan task type, biaya, latency, kualitas, project, user, dan override.

Tantangan:

- Menjaga policy tetap mudah dipahami.
- Menjaga precedence rule jelas.
- Menghindari konflik antar rule.
- Menyediakan default behavior saat rule tidak cocok.

### 16.5 Observability Without Privacy Violation

AtlasBridge membutuhkan telemetry untuk meningkatkan routing, tetapi tidak boleh membocorkan data sensitif.

Tantangan:

- Mencatat metadata yang cukup berguna.
- Menghindari penyimpanan prompt penuh.
- Mendeteksi secrets.
- Menyediakan mode audit yang aman.
- Mengatur retention data.

### 16.6 Streaming Performance

Streaming response harus diteruskan secepat mungkin dari 9Router ke client. Proxy tidak boleh melakukan buffering berlebihan.

Tantangan:

- Backpressure handling.
- Connection timeout.
- Chunk transformation minimal.
- Client disconnect handling.
- Error propagation saat streaming berjalan.

### 16.7 Client Diversity

OpenCode, Cursor, Cline, Continue, dan aplikasi lain mungkin memiliki perbedaan kecil dalam cara mengirim request.

Tantangan:

- Perbedaan headers.
- Perbedaan model field.
- Perbedaan tool schema.
- Perbedaan streaming expectation.
- Perbedaan timeout.

### 16.8 Maintaining Clear System Boundaries

AtlasBridge harus tetap menjadi intelligent decision layer. Jika terlalu banyak fitur ditambahkan, sistem dapat menjadi tumpang tindih dengan 9Router.

Tantangan:

- Menolak fitur yang seharusnya berada di 9Router.
- Menjaga desain tetap modular.
- Membuat dokumentasi boundary.
- Memisahkan decision routing dari provider routing.

---

## Final Positioning

AtlasBridge adalah intelligent routing proxy untuk AI coding workflow. Sistem ini bukan pengganti 9Router, melainkan lapisan tambahan di atas 9Router yang bertugas memahami request dan memilih routing intent terbaik.

Dengan desain ini, developer cukup menggunakan satu endpoint dan satu model virtual seperti `smart-auto`, sementara AtlasBridge menentukan jenis pekerjaan dan meneruskan request ke routing combo yang paling sesuai.

Nilai utama AtlasBridge adalah:

- Mengurangi manual model switching.
- Meningkatkan kecocokan model dengan task.
- Memaksimalkan pemanfaatan model gratis dan berbayar.
- Menjaga kompatibilitas OpenAI API.
- Tetap memanfaatkan kekuatan 9Router untuk reliability dan provider management.

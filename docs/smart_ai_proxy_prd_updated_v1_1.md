# Product Requirements Document (PRD)
# Smart AI Proxy

**Document Type:** Product Requirements Document  
**Project Name:** Smart AI Proxy  
**Status:** Draft v1.1  
**Primary Role:** OpenAI-Compatible Intelligent Routing Proxy with Local Web Control Panel  
**Primary Downstream:** 9Router  
**Target Users:** Developers, AI coding assistant users, engineering teams, technical admins  
**Prepared For:** Product, Engineering, Architecture, QA, DevOps, and UX Teams  
**Revision Notes:** Updated to include Web UI settings, user-configurable task-to-route mapping, route profile management, auto-start behavior, runtime control, and local configuration management.  

---

## 1. Executive Summary

Smart AI Proxy adalah OpenAI-compatible proxy server yang berada di antara AI coding assistant dan 9Router. Produk ini bertugas melakukan analisis request, klasifikasi task, pemilihan routing intent, serta meneruskan request ke 9Router menggunakan route profile atau AI combo yang paling sesuai.

Pada revisi v1.1, Smart AI Proxy juga diposisikan sebagai **local AI routing control center**. Artinya, user tidak hanya menggunakan proxy secara teknis, tetapi juga dapat mengatur perilaku routing melalui **Web UI lokal**. User dapat menentukan kategori pekerjaan tertentu harus diarahkan ke route tertentu, misalnya:

| Task Category | Example Route Profile |
|---|---|
| Design | `route.design` |
| Backend Engineering | `route.backend` |
| Frontend Engineering | `route.frontend` |
| Debugging | `route.debugging` |
| Documentation | `route.documentation` |
| Architecture Planning | `route.architect` |
| Low Cost Task | `route.low_cost` |

Smart AI Proxy tidak bertanggung jawab terhadap provider-level execution seperti failover, load balancing, rotasi akun, fallback model, rate limit handling, atau manajemen akun provider. Seluruh tanggung jawab tersebut tetap berada pada 9Router.

Produk ini dirancang untuk menyelesaikan masalah manual model switching pada workflow developer. Dengan Smart AI Proxy, pengguna cukup mengarahkan OpenCode, Cursor, Cline, Continue, atau aplikasi OpenAI-compatible lain ke satu endpoint proxy dan menggunakan model virtual seperti `smart-auto`. Proxy kemudian menentukan route terbaik berdasarkan jenis pekerjaan dan preferensi konfigurasi user.

---

## 2. Product Vision

Smart AI Proxy bertujuan menjadi intelligent decision layer untuk workflow AI coding modern. Produk ini memungkinkan developer menggunakan banyak model AI gratis maupun berbayar secara lebih efisien tanpa harus mengganti model secara manual.

Visi produk ini adalah:

> Menyediakan lapisan routing cerdas yang kompatibel dengan OpenAI API, mampu memahami jenis pekerjaan developer, menyediakan Web UI untuk pengaturan routing, memilih route profile terbaik, dan tetap memanfaatkan 9Router sebagai routing infrastructure utama.

Dalam jangka panjang, Smart AI Proxy dapat berkembang menjadi local AI orchestration control plane yang mampu mempertimbangkan kualitas model, latency, biaya, konteks proyek, preferensi user, telemetry historis, feedback pengguna, dan policy tim.

---

## 3. Background and Context

Developer saat ini sering menggunakan banyak AI model untuk berbagai kebutuhan. Beberapa model unggul untuk coding, beberapa lebih kuat untuk reasoning, beberapa lebih murah untuk task ringan, dan beberapa lebih cocok untuk long-context analysis.

Namun, pemilihan model masih sering dilakukan secara manual. Developer perlu memahami karakteristik setiap model dan mengganti konfigurasi berdasarkan kebutuhan. Hal ini menambah cognitive load, mengganggu workflow, dan sering menyebabkan penggunaan model yang tidak optimal.

9Router sudah menangani routing teknis di sisi provider, seperti failover, fallback, load balancing, rotasi akun, dan rate limit handling. Namun 9Router tidak diposisikan sebagai task-aware intelligent classifier ataupun user-facing configuration UI.

Oleh karena itu, Smart AI Proxy dibutuhkan sebagai lapisan tambahan yang fokus pada:

- Pemahaman request.
- Klasifikasi task.
- Pemilihan routing intent.
- Pengaturan task-to-route melalui Web UI.
- Kontrol runtime proxy.
- Auto-start ketika laptop atau komputer restart.
- Tetap meneruskan eksekusi provider-level kepada 9Router.

---

## 4. Problem Statement

Developer membutuhkan cara otomatis dan mudah dikonfigurasi untuk memilih model, route profile, atau AI combo terbaik berdasarkan jenis pekerjaan tanpa harus mengganti konfigurasi model secara manual.

Masalah utama yang ingin diselesaikan:

1. Developer terlalu sering melakukan manual model switching.
2. Developer tidak selalu mengetahui model terbaik untuk setiap jenis task.
3. Task sederhana sering dikirim ke model mahal atau berat.
4. Task kompleks sering dikirim ke model yang kurang kuat.
5. OpenAI-compatible tools belum tentu memiliki intelligent routing layer.
6. 9Router menangani provider routing, tetapi bukan task-aware routing UI.
7. Tidak ada abstraction layer yang jelas untuk memilih AI combo berdasarkan intent pekerjaan.
8. Pengaturan routing berbasis task sulit jika hanya dilakukan melalui config file.
9. User membutuhkan kontrol sederhana untuk menentukan apakah proxy hidup otomatis saat laptop restart.
10. User membutuhkan pilihan apakah proxy harus selalu aktif, manual, atau dimatikan sementara.

---

## 5. Goals

### 5.1 Product Goals

- Mengurangi kebutuhan developer mengganti model secara manual.
- Menyediakan intelligent routing berbasis task classification.
- Memungkinkan pengguna memakai satu endpoint dan satu model virtual seperti `smart-auto`.
- Memaksimalkan penggunaan model AI gratis dan berbayar sesuai spesialisasi task.
- Menjaga kompatibilitas dengan OpenAI API agar mudah digunakan oleh OpenCode, Cursor, Cline, Continue, dan tools sejenis.
- Menjaga batas tanggung jawab dengan 9Router secara jelas.
- Menyediakan Web UI lokal agar user dapat mengatur routing tanpa mengedit file manual.
- Menyediakan runtime control untuk start, stop, restart, dan status proxy.
- Menyediakan startup behavior setting agar proxy dapat hidup otomatis setelah laptop restart.

### 5.2 Technical Goals

- Menyediakan OpenAI-compatible proxy endpoint.
- Melakukan request analysis sebelum forwarding.
- Mengklasifikasikan task secara cepat dan ringan.
- Memilih route profile atau AI combo berdasarkan policy yang dapat dikonfigurasi.
- Meneruskan request ke 9Router tanpa merusak format request dan response.
- Mendukung streaming response.
- Menyediakan observability dasar tanpa menyimpan prompt penuh secara default.
- Menyediakan safe passthrough mode jika klasifikasi gagal.
- Menyediakan local persistent configuration store.
- Menyediakan Web UI yang berjalan secara lokal.
- Menyediakan mekanisme auto-start sesuai sistem operasi target.

### 5.3 Business/User Value Goals

- Meningkatkan produktivitas developer.
- Mengurangi biaya akibat penggunaan model besar untuk task ringan.
- Meningkatkan kualitas response karena model lebih sesuai dengan task.
- Menjadikan penggunaan multi-model lebih mudah bagi pengguna non-expert.
- Mengurangi kebutuhan edit config manual.
- Memberikan pengalaman setup yang lebih mudah melalui Web UI.
- Menyediakan fondasi untuk fitur routing cerdas tingkat lanjut.

---

## 6. Non-Goals

Smart AI Proxy tidak dirancang untuk menggantikan 9Router.

Produk ini tidak akan menangani:

- Provider failover.
- Provider load balancing.
- Rotasi akun provider.
- Provider fallback model.
- Provider rate limit handling.
- Provider credential management.
- Direct provider integration.
- Billing provider.
- Fine-tuning model.
- Training model AI.
- Menjalankan multi-agent execution secara langsung pada MVP.
- Menyimpan prompt penuh secara default.
- Menjadi IDE atau AI coding assistant sendiri.
- Membuka Web UI ke public network secara default.
- Mengelola akun provider langsung dari Smart AI Proxy.

---

## 7. Target Users and Personas

### 7.1 Individual Developer

Developer individu yang menggunakan AI coding assistant untuk coding harian, debugging, refactoring, dan dokumentasi. Mereka ingin workflow sederhana tanpa harus memikirkan model mana yang harus digunakan.

**Needs:**

- Satu endpoint untuk semua task.
- Model otomatis berdasarkan kebutuhan.
- Setup sederhana.
- Latency rendah.
- Bisa override manual saat diperlukan.
- Web UI untuk mengatur routing.
- Proxy bisa hidup otomatis setelah laptop restart.

### 7.2 Power User AI Coding Assistant

Pengguna tingkat lanjut yang memakai OpenCode, Cursor, Cline, Continue, atau tools lain dan memiliki banyak API key atau model route melalui 9Router.

**Needs:**

- Routing policy fleksibel.
- Mode kualitas tinggi, murah, cepat, atau auto.
- Konfigurasi task-to-route.
- Observability routing.
- Manual override.
- Route profile management.
- Import/export configuration.

### 7.3 Small Engineering Team

Tim kecil yang ingin menstandarkan penggunaan AI model untuk workflow development.

**Needs:**

- Shared configuration.
- Policy berbasis proyek.
- Logging metadata.
- Audit ringan.
- Konsistensi routing antar developer.
- Pengaturan default route untuk kategori kerja tim.

### 7.4 Technical Admin / DevOps

Pengelola infrastructure yang menjalankan Smart AI Proxy dan 9Router.

**Needs:**

- Deployment mudah.
- Health check.
- Config management.
- Observability.
- Security controls.
- Error monitoring.
- Runtime control.
- Startup management.

---

## 8. User Journey

### 8.1 First-Time Setup Journey

1. User menjalankan atau mengaktifkan Smart AI Proxy.
2. User membuka Web UI lokal Smart AI Proxy.
3. User mengatur downstream 9Router endpoint.
4. User mengatur proxy host dan port.
5. User mengatur API key atau authentication mode untuk proxy.
6. User memilih default route profile.
7. User mengatur task-to-route mapping, misalnya Design ke `route.design`, Backend ke `route.backend`, dan Documentation ke `route.documentation`.
8. User memilih startup mode: Always On, Manual, atau Disabled.
9. User mengatur AI coding assistant agar menggunakan endpoint Smart AI Proxy.
10. User memilih model virtual, misalnya `smart-auto`.
11. User mulai mengirim request dari OpenCode, Cursor, Cline, Continue, atau tool lain.
12. Smart AI Proxy menganalisis request dan meneruskan ke 9Router dengan routing intent yang sesuai.

### 8.2 Daily Developer Journey

1. Developer membuka laptop atau komputer.
2. Jika auto-start aktif, Smart AI Proxy berjalan otomatis di background.
3. Developer membuka AI coding assistant.
4. Developer meminta bantuan coding, debugging, refactoring, testing, desain, atau dokumentasi.
5. Developer tetap memakai model virtual yang sama.
6. Smart AI Proxy memilih route profile berdasarkan task dan setting user.
7. 9Router menangani provider-level routing.
8. Developer menerima response seperti menggunakan OpenAI-compatible API biasa.

### 8.3 Manual Override Journey

1. Developer merasa task tertentu membutuhkan route spesifik.
2. Developer menggunakan model alias khusus atau mengubah setting melalui Web UI.
3. Smart AI Proxy melewati auto classification atau memberi prioritas pada override.
4. Request diteruskan ke 9Router sesuai override.

### 8.4 Runtime Control Journey

1. User membuka Web UI.
2. User melihat status proxy: Running, Stopped, Error, atau Disabled.
3. User memilih Start, Stop, atau Restart.
4. Smart AI Proxy menjalankan tindakan runtime yang dipilih.
5. UI menampilkan status terbaru dan endpoint aktif.

### 8.5 Startup Mode Journey

1. User membuka halaman Startup Settings.
2. User memilih apakah proxy otomatis hidup saat laptop restart.
3. User memilih mode:
   - Always On.
   - Manual.
   - Disabled.
4. Setting disimpan ke local configuration store.
5. Setelah device restart, behavior proxy mengikuti setting tersebut.

---

## 9. Core Product Concepts

### 9.1 Smart Model Alias

Smart model alias adalah nama model virtual yang digunakan client. Alias ini tidak selalu merepresentasikan satu model nyata, tetapi merepresentasikan routing behavior.

Contoh:

| Alias | Purpose |
|---|---|
| `smart-auto` | Auto-routing berdasarkan klasifikasi task |
| `smart-code` | Routing prioritas untuk coding umum |
| `smart-debug` | Routing prioritas untuk debugging |
| `smart-refactor` | Routing prioritas untuk refactoring |
| `smart-architect` | Routing prioritas untuk architecture planning |
| `smart-docs` | Routing prioritas untuk dokumentasi |
| `smart-cheap` | Routing prioritas biaya rendah |
| `smart-fast` | Routing prioritas latency rendah |
| `smart-long-context` | Routing prioritas context panjang |

### 9.2 Task Classification

Task classification adalah proses menentukan jenis pekerjaan dari request.

Task type minimum:

| Task Type | Description |
|---|---|
| `general_chat` | Percakapan umum atau pertanyaan umum |
| `design_task` | UI/UX design, product design, layout, visual direction, design copy |
| `backend_engineering` | API, database, server, service, authentication, backend logic |
| `frontend_engineering` | UI implementation, component, CSS, state management, browser behavior |
| `fullstack_engineering` | Task yang melibatkan frontend dan backend sekaligus |
| `code_generation` | Membuat kode baru |
| `debugging` | Mencari dan memperbaiki error |
| `refactoring` | Merapikan atau mengubah struktur kode |
| `code_explanation` | Menjelaskan kode |
| `test_generation` | Membuat unit test, integration test, atau test case |
| `documentation` | Membuat README, komentar, docstring, atau dokumentasi teknis |
| `code_review` | Meninjau kualitas kode |
| `architecture_design` | Mendesain sistem, modul, service, atau technical blueprint |
| `planning` | Membuat rencana implementasi atau task breakdown |
| `security_review` | Meninjau risiko keamanan |
| `long_context_analysis` | Menganalisis konteks panjang atau multi-file |
| `data_transformation` | Mengubah format data atau struktur teks |
| `lightweight_task` | Task sederhana, pendek, atau low-complexity |

### 9.3 AI Combo

AI Combo adalah abstraksi tujuan routing. Smart AI Proxy memilih combo, sementara 9Router tetap menangani model/provider aktual.

Contoh AI Combo:

| AI Combo | Intended Use |
|---|---|
| `combo.fast_coding` | Coding cepat dan task implementasi ringan |
| `combo.deep_reasoning` | Reasoning kompleks dan architecture design |
| `combo.debugging` | Debugging dan root cause analysis |
| `combo.refactor` | Refactoring dan code cleanup |
| `combo.documentation` | Dokumentasi, README, komentar kode |
| `combo.low_cost` | Task ringan dengan prioritas biaya rendah |
| `combo.long_context` | Request dengan konteks panjang |
| `combo.security_review` | Review keamanan dan risiko |
| `combo.test_generation` | Pembuatan test dan test scenario |
| `combo.design` | Task desain, UI/UX, visual/product direction |
| `combo.backend` | Backend engineering dan API/server-side work |
| `combo.frontend` | Frontend engineering dan UI implementation |
| `combo.fullstack` | Task campuran frontend dan backend |

### 9.4 Route Profile

Route profile adalah abstraction layer yang lebih user-friendly dibanding AI combo. Route profile merepresentasikan tujuan routing yang dapat dikonfigurasi melalui Web UI.

Contoh route profile:

| Route Profile | Example Use |
|---|---|
| `route.design` | Untuk task desain, UI/UX, visual, dan product surface |
| `route.backend` | Untuk backend engineering, API, database, auth, service logic |
| `route.frontend` | Untuk frontend engineering, component, UI behavior, CSS |
| `route.fullstack` | Untuk task yang mencakup frontend dan backend |
| `route.debugging` | Untuk error analysis dan debugging |
| `route.refactor` | Untuk refactoring dan code cleanup |
| `route.documentation` | Untuk dokumentasi teknis |
| `route.architect` | Untuk architecture planning dan system design |
| `route.reasoning` | Untuk reasoning kompleks |
| `route.low_cost` | Untuk task ringan dengan prioritas biaya rendah |
| `route.long_context` | Untuk konteks panjang |
| `route.security` | Untuk security review |

Smart AI Proxy memilih route profile. 9Router tetap menerjemahkan route profile tersebut menjadi provider/model aktual sesuai konfigurasi downstream.

### 9.5 Task-to-Route Mapping

Task-to-route mapping adalah konfigurasi utama yang menghubungkan hasil klasifikasi task dengan route profile.

Contoh:

| Detected Task | User-Configured Route |
|---|---|
| `design_task` | `route.design` |
| `backend_engineering` | `route.backend` |
| `frontend_engineering` | `route.frontend` |
| `debugging` | `route.debugging` |
| `documentation` | `route.documentation` |
| `architecture_design` | `route.architect` |
| `lightweight_task` | `route.low_cost` |

Mapping ini harus bisa diubah dari Web UI tanpa mengubah kode inti.

### 9.6 Routing Policy

Routing policy adalah aturan yang menentukan mapping dari task type, user preference, project setting, dan override ke route profile atau AI combo tertentu.

Policy harus dapat dikonfigurasi tanpa mengubah kode inti.

### 9.7 Web Control Panel

Web Control Panel adalah UI lokal untuk mengatur Smart AI Proxy.

Fungsi utama:

- Melihat status proxy.
- Mengatur endpoint 9Router.
- Mengatur proxy port.
- Mengatur task-to-route mapping.
- Mengelola route profile.
- Mengatur startup behavior.
- Mengatur logging dan privacy mode.
- Mengimpor atau mengekspor konfigurasi.

### 9.8 Startup Mode

Startup mode menentukan apakah Smart AI Proxy otomatis berjalan setelah laptop atau komputer restart.

Mode utama:

| Mode | Description |
|---|---|
| Always On | Proxy otomatis aktif saat device menyala dan terus berjalan di background |
| Manual | Proxy hanya aktif ketika user menjalankannya secara manual |
| Disabled | Proxy tidak aktif sampai user mengaktifkannya kembali |

### 9.9 Safe Passthrough

Safe passthrough adalah mode ketika Smart AI Proxy tidak dapat menentukan route dengan confidence yang cukup. Dalam kondisi ini, request tetap diteruskan ke 9Router menggunakan default route.

---

## 10. Product Scope

### 10.1 MVP Scope

MVP Smart AI Proxy harus mencakup:

1. OpenAI-compatible chat completion proxy.
2. Streaming response passthrough.
3. Non-streaming response passthrough.
4. Model alias `smart-auto`.
5. Rule-based task classifier.
6. Basic routing policy engine.
7. Route profile selection.
8. AI combo selection.
9. Forwarding ke 9Router.
10. Manual override melalui model alias.
11. Basic metadata logging.
12. Safe passthrough fallback.
13. Health check endpoint.
14. Local persistent configuration store.
15. Web UI lokal untuk settings dasar.
16. Routing settings page untuk task-to-route mapping.
17. Startup settings untuk Always On, Manual, dan Disabled.
18. Runtime control untuk start, stop, restart, dan status proxy.
19. Basic privacy/logging settings.
20. Config import/export sederhana.

### 10.2 Post-MVP Scope

Post-MVP dapat mencakup:

1. Dashboard observability lebih lengkap.
2. Project-specific routing.
3. User-specific routing preference.
4. Cost-aware routing.
5. Latency-aware routing.
6. Feedback-based routing improvement.
7. Prompt redaction engine.
8. Policy validation UI.
9. Dry-run routing mode.
10. Multi-label classification.
11. Confidence score visualization.
12. Team workspace mode.
13. Shared team policy.
14. Advanced route profile templates.
15. Auto-update configuration package.

### 10.3 Out of Scope for MVP

MVP tidak mencakup:

1. Direct AI provider integration.
2. Provider account rotation.
3. Provider fallback.
4. Provider billing.
5. Complex enterprise dashboard.
6. Machine learning classifier.
7. Multi-agent orchestration.
8. Full prompt storage.
9. Enterprise RBAC.
10. Repository indexing.
11. Cloud-hosted management panel.
12. Public network Web UI by default.

---

## 11. Functional Requirements

### 11.1 OpenAI-Compatible Proxy

| ID | Requirement | Priority | Description | Acceptance Criteria |
|---|---|---|---|---|
| FR-001 | Accept OpenAI-compatible requests | Must Have | Proxy harus menerima request dengan format OpenAI-compatible. | Request dari client OpenAI-compatible dapat diterima tanpa perubahan besar. |
| FR-002 | Forward request to 9Router | Must Have | Request harus diteruskan ke downstream 9Router. | Request berhasil mencapai 9Router dan response diteruskan kembali ke client. |
| FR-003 | Preserve request structure | Must Have | Proxy tidak boleh merusak body request utama. | Field utama seperti `messages`, `model`, `temperature`, `tools`, dan `stream` tetap valid. |
| FR-004 | Preserve response format | Must Have | Response ke client harus tetap OpenAI-compatible. | Client dapat membaca response tanpa error format. |
| FR-005 | Support streaming | Must Have | Proxy harus mendukung streaming response. | Streaming chunk diteruskan ke client secara bertahap. |
| FR-006 | Support non-streaming | Must Have | Proxy harus mendukung non-streaming response. | Response non-streaming diterima sebagai full JSON response. |

### 11.2 Request Analysis

| ID | Requirement | Priority | Description | Acceptance Criteria |
|---|---|---|---|---|
| FR-007 | Parse message content | Must Have | Proxy harus dapat membaca system, user, dan assistant messages. | Message dapat dianalisis untuk klasifikasi task. |
| FR-008 | Detect code presence | Must Have | Proxy harus mendeteksi keberadaan code block atau pola source code. | Request dengan code block diberi signal coding-related. |
| FR-009 | Detect task keywords | Must Have | Proxy harus mengenali keyword seperti fix, debug, refactor, test, explain, review, document. | Keyword memengaruhi task classification. |
| FR-010 | Estimate request complexity | Should Have | Proxy perlu menilai kompleksitas sederhana, sedang, atau tinggi. | Request panjang atau multi-step diberi complexity lebih tinggi. |
| FR-011 | Detect long context | Should Have | Proxy harus mendeteksi prompt/context panjang. | Request di atas threshold tertentu diarahkan ke long-context route. |
| FR-012 | Detect domain category | Must Have | Proxy harus mengenali kategori design, backend, frontend, fullstack, docs, architecture, security, dan task ringan. | Task domain diarahkan ke route profile yang sesuai. |

### 11.3 Task Classification

| ID | Requirement | Priority | Description | Acceptance Criteria |
|---|---|---|---|---|
| FR-013 | Classify task type | Must Have | Proxy harus mengklasifikasikan request ke task type. | Setiap request memiliki task type atau fallback `unknown`. |
| FR-014 | Support coding task classification | Must Have | Proxy harus mengenali coding, debugging, refactoring, testing, documentation, review. | Prompt coding umum diklasifikasikan dengan benar pada test cases utama. |
| FR-015 | Support architecture task classification | Should Have | Proxy harus mengenali architecture planning dan system design. | Prompt architecture diarahkan ke reasoning/architect combo. |
| FR-016 | Support lightweight task classification | Should Have | Proxy harus mengenali task ringan. | Task pendek/sederhana dapat diarahkan ke low-cost/fast combo. |
| FR-017 | Provide classification confidence | Should Have | Classifier sebaiknya menghasilkan confidence score. | Routing engine dapat memakai confidence untuk fallback. |
| FR-018 | Support design/backend/frontend classification | Must Have | Proxy harus mengenali task design, backend, frontend, dan fullstack. | Task tersebut dapat dipetakan ke route profile yang diatur user di Web UI. |

### 11.4 Routing Policy Engine

| ID | Requirement | Priority | Description | Acceptance Criteria |
|---|---|---|---|---|
| FR-019 | Map task to route profile | Must Have | Task type harus dipetakan ke route profile. | Debugging diarahkan ke `route.debugging`, documentation ke `route.documentation`, dan seterusnya. |
| FR-020 | Support default route | Must Have | Jika task tidak dikenali, proxy harus memakai default route. | Request tetap diteruskan ke 9Router. |
| FR-021 | Support configurable policy | Must Have | Mapping task ke route profile harus bisa dikonfigurasi. | Perubahan policy tidak membutuhkan perubahan kode inti. |
| FR-022 | Support priority mode | Should Have | Policy dapat mempertimbangkan mode cepat, murah, atau kualitas tinggi. | User dapat memilih preference routing. |
| FR-023 | Support policy validation | Should Have | Policy harus divalidasi saat startup atau reload. | Policy invalid menghasilkan error konfigurasi yang jelas. |
| FR-024 | Support UI-based policy update | Must Have | User dapat mengubah routing policy melalui Web UI. | Perubahan disimpan dan dipakai pada request berikutnya. |

### 11.5 Route Profile and AI Combo Selection

| ID | Requirement | Priority | Description | Acceptance Criteria |
|---|---|---|---|---|
| FR-025 | Select route profile | Must Have | Proxy harus memilih route profile berdasarkan policy. | Setiap request memiliki selected route profile atau default route. |
| FR-026 | Select AI combo | Must Have | Proxy dapat memilih combo berdasarkan route profile. | Route profile dapat diterjemahkan ke AI combo atau metadata downstream. |
| FR-027 | Add routing metadata | Must Have | Proxy harus mengirim routing target ke 9Router melalui mekanisme yang disepakati. | 9Router dapat menerima atau memahami route intent. |
| FR-028 | Preserve original request intent | Must Have | Pemilihan route tidak boleh mengubah instruksi utama user. | Isi prompt utama tidak dimodifikasi secara agresif. |
| FR-029 | Support route alias | Should Have | Route dapat direpresentasikan sebagai model alias atau header. | Route dapat dikirim sebagai `model`, header, atau metadata sesuai konfigurasi. |

### 11.6 Manual Override

| ID | Requirement | Priority | Description | Acceptance Criteria |
|---|---|---|---|---|
| FR-030 | Support model alias override | Must Have | User dapat memilih alias seperti `smart-debug` atau `smart-cheap`. | Proxy menghormati alias dan memilih combo terkait. |
| FR-031 | Support auto-routing disable | Should Have | User dapat menonaktifkan auto-routing untuk request tertentu. | Request diteruskan sesuai model asli atau default passthrough. |
| FR-032 | Support explicit route selection | Should Have | User dapat memilih route profile tertentu. | Proxy tidak melakukan klasifikasi ulang jika override eksplisit valid. |
| FR-033 | Support UI override | Should Have | User dapat mengunci default route dari Web UI. | Semua request dapat diarahkan ke route tertentu jika mode override aktif. |

### 11.7 Observability and Logging

| ID | Requirement | Priority | Description | Acceptance Criteria |
|---|---|---|---|---|
| FR-034 | Generate request ID | Must Have | Setiap request harus memiliki request ID. | Log dan response tracing memiliki ID unik. |
| FR-035 | Log routing metadata | Must Have | Proxy mencatat task type, selected route, selected combo, confidence, dan latency. | Metadata tersedia untuk debugging dan analisis. |
| FR-036 | Avoid full prompt logging by default | Must Have | Prompt penuh tidak boleh disimpan secara default. | Log default tidak berisi source code/prompt penuh. |
| FR-037 | Log errors | Must Have | Error harus dicatat dengan class dan context non-sensitif. | Error dapat dianalisis tanpa membocorkan data rahasia. |
| FR-038 | Provide health check | Must Have | Proxy harus menyediakan health check. | Monitoring dapat mengecek status proxy. |
| FR-039 | Show basic metrics in Web UI | Should Have | Web UI menampilkan request count, selected routes, dan status proxy. | User dapat melihat ringkasan aktivitas tanpa membuka log manual. |

### 11.8 Security and Privacy

| ID | Requirement | Priority | Description | Acceptance Criteria |
|---|---|---|---|---|
| FR-040 | Forward authentication safely | Must Have | Proxy harus menangani authentication ke 9Router secara aman. | API key tidak bocor ke log. |
| FR-041 | Redact sensitive headers | Must Have | Header sensitif harus disembunyikan dari log. | Authorization dan token tidak muncul di log. |
| FR-042 | Support downstream allowlist | Should Have | Proxy hanya boleh meneruskan request ke endpoint downstream yang dikonfigurasi. | Endpoint tidak dapat diubah sembarangan oleh client. |
| FR-043 | Support privacy mode | Should Have | User dapat menjalankan mode privacy tinggi. | Logging dibatasi hanya ke metadata minimum. |
| FR-044 | Local-only Web UI by default | Must Have | Web UI tidak boleh terbuka ke public network secara default. | UI hanya listen di localhost kecuali user mengaktifkan opsi lanjutan. |
| FR-045 | Protect Web UI settings | Must Have | Web UI harus memiliki proteksi dasar seperti local token/password. | Setting tidak dapat diubah tanpa otorisasi lokal. |

### 11.9 Web Admin UI

| ID | Requirement | Priority | Description | Acceptance Criteria |
|---|---|---|---|---|
| FR-046 | Provide local Web UI | Must Have | Smart AI Proxy harus menyediakan Web UI lokal. | User dapat membuka halaman settings melalui browser lokal. |
| FR-047 | Show dashboard status | Must Have | UI harus menampilkan status proxy, endpoint, mode, dan downstream 9Router. | User dapat melihat apakah proxy running dan tersambung ke 9Router. |
| FR-048 | Provide routing settings page | Must Have | UI harus menyediakan halaman task-to-route mapping. | User dapat mengubah mapping kategori task ke route profile. |
| FR-049 | Provide route profile page | Should Have | UI menyediakan pengelolaan route profile. | User dapat melihat, membuat, mengedit, dan menonaktifkan route profile. |
| FR-050 | Provide startup settings page | Must Have | UI menyediakan pengaturan auto-start dan runtime mode. | User dapat memilih Always On, Manual, atau Disabled. |
| FR-051 | Provide privacy/logging page | Should Have | UI menyediakan pengaturan logging dan privacy mode. | User dapat mengatur metadata logs, strict mode, dan clear logs. |
| FR-052 | Provide advanced settings page | Should Have | UI menyediakan pengaturan port, downstream endpoint, timeout, dan import/export config. | User dapat mengubah setting teknis dengan validasi. |

### 11.10 Runtime and Startup Management

| ID | Requirement | Priority | Description | Acceptance Criteria |
|---|---|---|---|---|
| FR-053 | Start proxy from UI | Must Have | User dapat menjalankan proxy dari Web UI. | Status berubah menjadi Running dan endpoint aktif. |
| FR-054 | Stop proxy from UI | Must Have | User dapat menghentikan proxy dari Web UI. | Status berubah menjadi Stopped dan request baru tidak diterima. |
| FR-055 | Restart proxy from UI | Must Have | User dapat restart proxy dari Web UI. | Proxy berhenti lalu berjalan kembali dengan config terbaru. |
| FR-056 | Enable auto-start on boot | Must Have | User dapat mengaktifkan proxy agar hidup saat laptop restart. | Setelah restart, proxy berjalan sesuai setting. |
| FR-057 | Disable auto-start on boot | Must Have | User dapat menonaktifkan auto-start. | Setelah restart, proxy tidak otomatis berjalan jika mode Manual/Disabled. |
| FR-058 | Support Always On mode | Must Have | Proxy berjalan otomatis dan tetap aktif di background. | Proxy aktif setelah device startup. |
| FR-059 | Support Manual mode | Must Have | Proxy hanya berjalan ketika user menyalakannya. | Proxy tidak otomatis aktif saat startup. |
| FR-060 | Support Disabled mode | Must Have | Proxy tidak menerima request sampai user mengaktifkan ulang. | Status Disabled terlihat jelas di UI. |

### 11.11 Local Configuration Management

| ID | Requirement | Priority | Description | Acceptance Criteria |
|---|---|---|---|---|
| FR-061 | Persist user settings locally | Must Have | Semua setting utama harus tersimpan lokal. | Setting tetap ada setelah restart. |
| FR-062 | Validate configuration before save | Must Have | UI harus memvalidasi config sebelum disimpan. | Config invalid ditolak dengan pesan error jelas. |
| FR-063 | Support config import | Should Have | User dapat mengimpor config. | Config valid dapat dipakai setelah import. |
| FR-064 | Support config export | Should Have | User dapat mengekspor config. | File config dapat digunakan untuk backup atau sharing. |
| FR-065 | Support reset configuration | Should Have | User dapat mengembalikan default config. | Setting kembali ke default aman. |

---

## 12. Non-Functional Requirements

### 12.1 Performance

| ID | Requirement | Target |
|---|---|---|
| NFR-001 | Routing decision overhead | Minimal dan tidak terasa pada workflow coding |
| NFR-002 | Classifier execution | Harus ringan dan tidak bergantung pada LLM untuk setiap request di MVP |
| NFR-003 | Streaming overhead | Proxy tidak boleh melakukan buffering berlebihan |
| NFR-004 | Startup time | Cepat dan cocok untuk local/self-host deployment |
| NFR-005 | Web UI responsiveness | Settings page harus terasa ringan dan cepat untuk local use |

### 12.2 Reliability

| ID | Requirement | Target |
|---|---|---|
| NFR-006 | Safe fallback | Request tetap diteruskan meskipun classifier gagal |
| NFR-007 | Error propagation | Error dari 9Router diteruskan dengan format kompatibel |
| NFR-008 | Config failure handling | Policy invalid harus menghasilkan pesan error jelas |
| NFR-009 | Health check | Proxy dapat dipantau oleh external monitor |
| NFR-010 | Startup recovery | Proxy dapat berjalan kembali setelah restart jika Always On aktif |
| NFR-011 | Runtime mode consistency | Status proxy di UI harus sesuai kondisi runtime sebenarnya |

### 12.3 Compatibility

| ID | Requirement | Target |
|---|---|---|
| NFR-012 | OpenAI API compatibility | Kompatibel dengan OpenAI-compatible clients |
| NFR-013 | Tool calling compatibility | Tidak merusak struktur tools/function calling |
| NFR-014 | Client compatibility | Mendukung OpenCode, Cursor, Cline, Continue, dan tools sejenis |
| NFR-015 | Response compatibility | Client tidak perlu parser khusus |
| NFR-016 | Local browser compatibility | Web UI dapat digunakan di browser modern umum |

### 12.4 Security

| ID | Requirement | Target |
|---|---|---|
| NFR-017 | Secret protection | API key dan token tidak masuk log |
| NFR-018 | Data minimization | Log default hanya metadata |
| NFR-019 | Endpoint protection | Downstream 9Router endpoint dikontrol oleh config |
| NFR-020 | Auditability | Routing decision dapat ditelusuri tanpa menyimpan prompt penuh |
| NFR-021 | Local UI security | Web UI hanya localhost secara default dan dilindungi token/password lokal |
| NFR-022 | Safe config storage | Konfigurasi sensitif tidak ditampilkan penuh dan tidak bocor ke log |

### 12.5 Scalability

| ID | Requirement | Target |
|---|---|---|
| NFR-023 | Stateless proxy path | Request forwarding path sebaiknya tetap ringan dan scalable |
| NFR-024 | Externalized config | Policy dapat dipisah dari runtime jika dibutuhkan |
| NFR-025 | Low memory usage | Cocok untuk personal dan team deployment |
| NFR-026 | Concurrent request handling | Mampu menangani banyak request paralel |
| NFR-027 | Config reload behavior | Perubahan setting tidak boleh mengganggu request yang sedang berjalan secara tidak perlu |

---

## 13. Routing Behavior Requirements

### 13.1 Default Routing Matrix

| Detected Task | Recommended Route Profile | Recommended Combo | Priority Logic |
|---|---|---|---|
| General chat | `route.low_cost` or default | `combo.low_cost` or default | Hemat biaya jika tidak kompleks |
| Design task | `route.design` | `combo.design` | Cocok untuk UI/UX, product design, visual, dan design copy |
| Backend engineering | `route.backend` | `combo.backend` | Cocok untuk API, database, service, auth, backend logic |
| Frontend engineering | `route.frontend` | `combo.frontend` | Cocok untuk component, UI behavior, CSS, browser/client logic |
| Fullstack engineering | `route.fullstack` | `combo.fullstack` | Cocok untuk task lintas frontend dan backend |
| Code generation | `route.backend`, `route.frontend`, or `route.fullstack` | `combo.fast_coding` | Model coding cepat dan akurat |
| Debugging | `route.debugging` | `combo.debugging` | Analisis error dan root cause |
| Refactoring | `route.refactor` | `combo.refactor` | Kualitas kode dan maintainability |
| Code explanation | `route.low_cost` or `route.backend` | `combo.low_cost` or `combo.fast_coding` | Hemat biaya kecuali konteks kompleks |
| Test generation | `route.testing` | `combo.test_generation` | Test logic dan coverage |
| Documentation | `route.documentation` | `combo.documentation` | Kualitas writing dan struktur |
| Code review | `route.reasoning` or `route.security` | `combo.deep_reasoning` or `combo.security_review` | Review kualitas atau risiko |
| Architecture design | `route.architect` | `combo.deep_reasoning` | Reasoning dan desain sistem |
| Planning | `route.reasoning` or `route.backend` | `combo.deep_reasoning` or `combo.fast_coding` | Tergantung kompleksitas |
| Security review | `route.security` | `combo.security_review` | Analisis risiko |
| Long context analysis | `route.long_context` | `combo.long_context` | Context window besar |
| Lightweight task | `route.low_cost` | `combo.low_cost` or `combo.fast` | Latency dan biaya rendah |
| Unknown | default route | default combo | Safe passthrough |

### 13.2 Example User Configuration

| Task Category | User Preference Example | Route Profile |
|---|---|---|
| Design | DeepSeek-based route through 9Router | `route.design` |
| Backend | Gemini-based route through 9Router | `route.backend` |
| Frontend | Claude/Gemini-based route through 9Router | `route.frontend` |
| Debugging | Strong reasoning model route through 9Router | `route.debugging` |
| Documentation | Fast writing route through 9Router | `route.documentation` |
| Lightweight Task | Cheap/free model route through 9Router | `route.low_cost` |

Catatan penting: Smart AI Proxy tidak memilih provider akhir secara langsung. Smart AI Proxy hanya meneruskan route profile atau routing intent ke 9Router. Konfigurasi provider/model aktual tetap diatur oleh 9Router.

### 13.3 Routing Decision Precedence

Urutan prioritas routing:

1. Explicit user override.
2. Model alias override.
3. UI-configured locked route.
4. Project-specific policy.
5. User-specific preference.
6. Task-to-route mapping dari Web UI.
7. Task classification result.
8. Complexity and context signal.
9. Default route.
10. Safe passthrough.

### 13.4 Classification Confidence Behavior

| Confidence Level | Behavior |
|---|---|
| High | Gunakan route profile hasil klasifikasi |
| Medium | Gunakan route profile hasil klasifikasi dengan fallback aman |
| Low | Gunakan default route atau safe passthrough |
| Unknown | Jangan blokir request; teruskan ke 9Router |

---

## 14. Configuration Requirements

Smart AI Proxy harus mendukung konfigurasi yang fleksibel dan dapat diubah melalui Web UI.

### 14.1 Required Configuration

- Smart AI Proxy listen host/port.
- 9Router downstream endpoint.
- Authentication mode.
- Default route.
- Mapping model alias ke route profile atau combo.
- Mapping task type ke route profile.
- Logging level.
- Privacy mode.
- Startup mode.
- Runtime mode.
- Web UI access token/password.

### 14.2 Optional Configuration

- User preference.
- Project preference.
- Cost priority.
- Latency priority.
- Quality priority.
- Long-context threshold.
- Classifier confidence threshold.
- Custom headers to forward.
- Headers to redact.
- Telemetry enable/disable.
- Config import/export.
- Debug mode.
- UI theme.

### 14.3 Configuration Principles

- Konfigurasi harus mudah dibaca manusia.
- Web UI harus menjadi cara utama konfigurasi untuk user umum.
- Config file tetap dapat digunakan untuk power user.
- Policy harus dapat divalidasi.
- Default configuration harus aman.
- Perubahan model/provider sebaiknya tidak membutuhkan perubahan kode inti.
- Routing policy harus dapat berkembang mengikuti perubahan ekosistem AI model.
- Setting harus persisten setelah device restart.

---

## 15. API Compatibility Requirements

### 15.1 Client Compatibility

Smart AI Proxy harus kompatibel dengan client yang menggunakan OpenAI-compatible API, terutama:

- OpenCode.
- Cursor.
- Cline.
- Continue.
- Custom AI coding assistant.
- Script atau aplikasi internal yang memakai OpenAI-compatible chat completion.

### 15.2 Endpoint Compatibility

MVP harus memprioritaskan chat completion compatible endpoint.

Compatibility yang perlu dijaga:

- Request body structure.
- Response body structure.
- Streaming events.
- Error response format.
- Tool/function calling payload.
- Model field behavior.
- Header forwarding behavior.

### 15.3 Model Field Behavior

Smart AI Proxy harus dapat menerima model field seperti:

- Model nyata yang ingin dipassthrough.
- Model virtual seperti `smart-auto`.
- Model alias seperti `smart-debug`.
- Route alias jika didukung.
- Combo alias jika didukung.

---

## 16. UX Requirements

Smart AI Proxy memiliki dua area UX utama:

1. Developer experience saat digunakan dari AI coding assistant.
2. Web UI experience saat user mengatur proxy.

### 16.1 Setup Experience

User harus dapat memahami:

- Endpoint mana yang dipakai.
- Model alias apa yang tersedia.
- Bagaimana menghubungkan OpenCode/Cursor/Cline/Continue.
- Bagaimana memilih mode auto/manual.
- Bagaimana mengatur task-to-route mapping.
- Bagaimana mengaktifkan auto-start.
- Bagaimana melihat status proxy.
- Bagaimana melihat log routing.

### 16.2 Runtime Experience

User harus merasakan:

- Tidak perlu sering mengganti model.
- Response tetap kompatibel dengan tool yang digunakan.
- Streaming tetap berjalan normal.
- Override tersedia ketika dibutuhkan.
- Error mudah dipahami.
- Proxy dapat hidup otomatis setelah restart jika user mengaktifkannya.

### 16.3 Web UI Main Pages

#### 16.3.1 Dashboard

Dashboard menampilkan:

- Proxy status: Running, Stopped, Error, atau Disabled.
- Current OpenAI-compatible endpoint.
- Downstream 9Router endpoint.
- 9Router connection status.
- Startup mode.
- Current default route.
- Request count hari ini.
- Most used task type.
- Most used route profile.

#### 16.3.2 Routing Settings

Routing Settings menampilkan:

- Task category list.
- Route profile dropdown untuk setiap task.
- Default route setting.
- Auto-routing toggle.
- Manual override mode.
- Save button.
- Reset to default button.

#### 16.3.3 Route Profiles

Route Profiles menampilkan:

- Nama route profile.
- Deskripsi route profile.
- Target metadata untuk 9Router.
- Priority mode: speed, quality, cost, balanced.
- Active/inactive status.
- Create/edit/delete route profile.

#### 16.3.4 Startup Settings

Startup Settings menampilkan:

- Auto-start on boot toggle.
- Always On mode.
- Manual mode.
- Disabled mode.
- Restart after crash option.
- Run in background option.

#### 16.3.5 Privacy and Logs

Privacy and Logs menampilkan:

- Enable metadata logs.
- Disable prompt logging.
- Strict privacy mode.
- Redact secrets.
- Log retention period.
- Clear logs.
- Export diagnostic report.

#### 16.3.6 Advanced Settings

Advanced Settings menampilkan:

- Proxy host and port.
- 9Router base URL.
- API key forwarding mode.
- Timeout setting.
- Streaming mode.
- Debug mode.
- Config import/export.
- Reset configuration.

### 16.4 Debugging Experience

Saat terjadi routing yang tidak sesuai, user/admin harus dapat melihat:

- Request ID.
- Task type yang terdeteksi.
- Confidence score.
- Selected route profile.
- Selected combo.
- Routing reason.
- Downstream status.
- Latency proxy dan downstream.

---

## 17. Observability Requirements

### 17.1 Required Metrics

- Total requests.
- Requests by task type.
- Requests by selected route profile.
- Requests by selected combo.
- Classification confidence distribution.
- Proxy latency.
- Downstream latency.
- Error rate.
- Streaming request count.
- Non-streaming request count.
- Safe passthrough count.
- Manual override count.
- Auto-start success/failure count.
- Runtime status changes.

### 17.2 Required Logs

Log minimum:

- Timestamp.
- Request ID.
- Client identifier jika tersedia.
- Model requested.
- Task type.
- Selected route profile.
- Selected combo.
- Confidence score jika tersedia.
- Routing decision reason.
- Status code.
- Latency.
- Error class jika ada.
- Runtime action jika user melakukan start/stop/restart.

### 17.3 Data That Should Not Be Logged by Default

- Full prompt.
- Source code penuh.
- API keys.
- Authorization headers.
- Secrets.
- Personal data.
- Provider credentials.
- Password Web UI.

---

## 18. Security and Privacy Requirements

### 18.1 Security Principles

- Secure by default.
- Minimum data retention.
- No full prompt logging by default.
- Sensitive header redaction.
- Downstream endpoint allowlist.
- Clear authentication boundary.
- Safe error messages.
- Web UI localhost-only by default.
- Web UI protected by local admin token or password.

### 18.2 Privacy Modes

| Mode | Description |
|---|---|
| Standard | Log metadata routing tanpa prompt penuh |
| Strict | Log hanya request ID, status, latency, dan selected route |
| Debug | Dapat menampilkan informasi tambahan, tetapi harus eksplisit diaktifkan |

### 18.3 Secret Handling

Proxy harus menghindari pencatatan:

- API key.
- Bearer token.
- Password.
- Private key.
- Environment secret.
- Credential dalam prompt jika terdeteksi.
- Web UI access token.

### 18.4 Web UI Security

- Web UI harus listen di localhost secara default.
- Public network access harus disabled secara default.
- Jika public access diaktifkan, user harus menerima warning keamanan.
- Sensitive values harus di-mask di UI.
- Perubahan konfigurasi kritikal harus divalidasi.
- Reset configuration harus meminta konfirmasi.

---

## 19. MVP Release Criteria

MVP dianggap siap dirilis jika memenuhi kriteria berikut:

1. Dapat menerima request OpenAI-compatible dari minimal satu AI coding assistant.
2. Dapat meneruskan request ke 9Router dan menerima response.
3. Streaming response berjalan stabil.
4. Model alias `smart-auto` tersedia.
5. Rule-based classifier dapat mengenali task dasar.
6. Routing policy dapat memetakan task ke route profile.
7. User dapat mengubah task-to-route mapping melalui Web UI.
8. User dapat mengatur downstream 9Router endpoint melalui Web UI.
9. User dapat melihat status proxy melalui Web UI.
10. User dapat start, stop, dan restart proxy melalui Web UI.
11. User dapat memilih startup mode: Always On, Manual, atau Disabled.
12. Safe passthrough berjalan saat klasifikasi gagal.
13. Log metadata routing tersedia.
14. Prompt penuh tidak tersimpan secara default.
15. Health check tersedia.
16. Dokumentasi setup dasar tersedia.
17. Tidak ada fitur provider-level routing yang mengambil alih peran 9Router.

---

## 20. User Stories

### 20.1 Auto Routing

**As a** developer,  
**I want** menggunakan satu model virtual seperti `smart-auto`,  
**so that** saya tidak perlu mengganti model secara manual untuk coding, debugging, refactoring, atau dokumentasi.

Acceptance criteria:

- User dapat mengatur client ke endpoint Smart AI Proxy.
- User dapat memilih model `smart-auto`.
- Proxy memilih route berdasarkan request.
- Response tetap muncul normal di client.

### 20.2 Routing Settings via Web UI

**As a** developer,  
**I want** mengatur task tertentu diarahkan ke route tertentu melalui Web UI,  
**so that** saya tidak perlu mengedit config file secara manual.

Acceptance criteria:

- User dapat membuka Routing Settings.
- User dapat memilih route profile untuk setiap task category.
- Setting tersimpan.
- Request berikutnya memakai mapping terbaru.

### 20.3 Design Route Configuration

**As a** developer,  
**I want** task desain diarahkan ke route khusus design,  
**so that** task UI/UX atau visual direction menggunakan AI yang saya anggap paling cocok melalui 9Router.

Acceptance criteria:

- Task design dikenali oleh classifier.
- Task design diarahkan ke `route.design`.
- `route.design` diteruskan ke 9Router sebagai route intent.

### 20.4 Backend Route Configuration

**As a** backend developer,  
**I want** task backend diarahkan ke route backend,  
**so that** API, database, auth, dan server logic memakai model yang saya pilih melalui 9Router.

Acceptance criteria:

- Task backend dikenali oleh classifier.
- Task backend diarahkan ke `route.backend`.
- User dapat mengubah route backend dari Web UI.

### 20.5 Debugging Routing

**As a** developer,  
**I want** prompt debugging otomatis diarahkan ke route debugging,  
**so that** hasil analisis error lebih relevan.

Acceptance criteria:

- Prompt berisi error, stack trace, atau kata debug/fix diklasifikasikan sebagai debugging.
- Selected route adalah `route.debugging` atau policy equivalent.
- Request tetap diteruskan ke 9Router.

### 20.6 Low-Cost Routing

**As a** developer,  
**I want** task ringan diarahkan ke route murah atau cepat,  
**so that** penggunaan model mahal dapat dikurangi.

Acceptance criteria:

- Prompt sederhana diklasifikasikan sebagai lightweight task.
- Selected route adalah `route.low_cost` atau equivalent.
- User tetap bisa override jika ingin route lebih kuat.

### 20.7 Auto-Start on Boot

**As a** developer,  
**I want** Smart AI Proxy otomatis hidup saat laptop restart,  
**so that** saya tidak perlu menyalakan proxy manual setiap kali mulai bekerja.

Acceptance criteria:

- User dapat mengaktifkan Always On mode dari Web UI.
- Setting tersimpan.
- Setelah restart, proxy aktif otomatis.
- UI menampilkan status Running.

### 20.8 Manual Mode

**As a** user,  
**I want** proxy hanya hidup ketika saya jalankan manual,  
**so that** saya dapat mengontrol kapan proxy aktif.

Acceptance criteria:

- User dapat memilih Manual mode.
- Proxy tidak otomatis hidup setelah restart.
- User dapat menyalakan proxy dari Web UI.

### 20.9 Disabled Mode

**As a** user,  
**I want** dapat mematikan Smart AI Proxy sementara,  
**so that** tidak ada request yang lewat proxy saat saya tidak membutuhkannya.

Acceptance criteria:

- User dapat memilih Disabled mode.
- Proxy menolak atau tidak menerima request baru dengan status yang jelas.
- UI menampilkan status Disabled.

### 20.10 Safe Passthrough

**As a** user,  
**I want** request tetap berjalan walaupun classifier gagal,  
**so that** workflow saya tidak terganggu.

Acceptance criteria:

- Jika classifier error, request diteruskan ke default route.
- Error classifier dicatat sebagai metadata.
- Client tetap menerima response dari 9Router jika downstream berhasil.

---

## 21. Acceptance Test Scenarios

### 21.1 Basic Passthrough Test

**Scenario:** Client mengirim request biasa ke Smart AI Proxy.  
**Expected Result:** Request diteruskan ke 9Router dan response dikembalikan ke client dalam format OpenAI-compatible.

### 21.2 Streaming Test

**Scenario:** Client mengirim request dengan `stream: true`.  
**Expected Result:** Proxy meneruskan streaming chunk dari 9Router ke client tanpa buffering berlebihan.

### 21.3 Debugging Classification Test

**Scenario:** User mengirim stack trace dan meminta perbaikan.  
**Expected Result:** Task diklasifikasikan sebagai debugging dan diarahkan ke `route.debugging`.

### 21.4 Documentation Classification Test

**Scenario:** User meminta README atau docstring.  
**Expected Result:** Task diklasifikasikan sebagai documentation dan diarahkan ke `route.documentation`.

### 21.5 Architecture Classification Test

**Scenario:** User meminta desain sistem multi-tenant.  
**Expected Result:** Task diklasifikasikan sebagai architecture design dan diarahkan ke `route.architect` atau `route.reasoning`.

### 21.6 Design Classification Test

**Scenario:** User meminta desain UI, wireframe, UX flow, atau product page.  
**Expected Result:** Task diklasifikasikan sebagai design task dan diarahkan ke `route.design`.

### 21.7 Backend Classification Test

**Scenario:** User meminta pembuatan API, database schema, auth flow, atau service backend.  
**Expected Result:** Task diklasifikasikan sebagai backend engineering dan diarahkan ke `route.backend`.

### 21.8 Frontend Classification Test

**Scenario:** User meminta component, styling, browser behavior, atau UI implementation.  
**Expected Result:** Task diklasifikasikan sebagai frontend engineering dan diarahkan ke `route.frontend`.

### 21.9 Web UI Routing Save Test

**Scenario:** User mengubah Design dari `route.design` ke route lain melalui Web UI.  
**Expected Result:** Setting tersimpan dan request design berikutnya memakai route baru.

### 21.10 Startup Always On Test

**Scenario:** User mengaktifkan Always On mode lalu device restart.  
**Expected Result:** Proxy aktif otomatis setelah restart.

### 21.11 Manual Mode Test

**Scenario:** User memilih Manual mode lalu device restart.  
**Expected Result:** Proxy tidak otomatis berjalan sampai user menyalakannya.

### 21.12 Disabled Mode Test

**Scenario:** User memilih Disabled mode.  
**Expected Result:** Proxy tidak menerima request baru dan UI menampilkan status Disabled.

### 21.13 Manual Override Test

**Scenario:** User memilih model alias `smart-cheap`.  
**Expected Result:** Proxy memilih route low-cost tanpa auto-routing penuh.

### 21.14 Classifier Failure Test

**Scenario:** Classifier gagal karena input tidak valid atau internal error.  
**Expected Result:** Proxy memakai safe passthrough dan request tetap diteruskan ke 9Router.

### 21.15 Privacy Logging Test

**Scenario:** User mengirim source code dalam prompt.  
**Expected Result:** Log default tidak menyimpan prompt/source code penuh.

---

## 22. Product Metrics

### 22.1 Adoption Metrics

- Jumlah active users.
- Jumlah connected clients.
- Jumlah request harian.
- Jumlah request melalui `smart-auto`.
- Jumlah manual override.
- Jumlah user yang membuka Web UI.
- Jumlah user yang menyelesaikan setup melalui Web UI.

### 22.2 Routing Metrics

- Classification success rate.
- Safe passthrough rate.
- Task distribution.
- Route profile distribution.
- Combo distribution.
- Low-confidence classification rate.
- Manual override frequency.
- Task-to-route mapping customization rate.

### 22.3 Quality Metrics

- User retry rate setelah response.
- User manual reroute rate.
- Feedback positif/negatif terhadap routing.
- Task completion satisfaction.
- Routing accuracy berdasarkan evaluation set.

### 22.4 Performance Metrics

- Proxy overhead latency.
- End-to-end latency.
- Streaming first-token delay.
- Error rate.
- Timeout rate.
- Web UI load time.

### 22.5 Cost Efficiency Metrics

- Persentase lightweight task yang diarahkan ke low-cost route.
- Penurunan penggunaan model mahal untuk task sederhana.
- Estimasi cost saving per user/team.

### 22.6 Runtime Metrics

- Auto-start success rate.
- Auto-start failure rate.
- Start/stop/restart actions.
- Crash/recovery count.
- Active mode distribution: Always On, Manual, Disabled.

---

## 23. Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Salah klasifikasi task | Model/combo kurang sesuai | Confidence score, fallback, manual override |
| Latency tambahan | Workflow terasa lambat | Classifier ringan, caching, minimal processing |
| Streaming rusak | Client tidak kompatibel | Compatibility test dan minimal transformation |
| Boundary creep dengan 9Router | Arsitektur tumpang tindih | Dokumentasi responsibility boundary |
| Prompt/source code bocor di log | Risiko keamanan tinggi | No full prompt logging by default, redaction |
| Policy salah konfigurasi | Routing buruk | Policy validation dan default config |
| Client berbeda-beda implementasi | Compatibility issue | Test matrix untuk OpenCode, Cursor, Cline, Continue |
| 9Router metadata tidak sesuai | Route intent tidak terbaca | Gunakan mekanisme route yang disepakati dan fallback |
| Web UI terbuka ke jaringan publik | Risiko keamanan tinggi | Localhost-only by default, warning untuk public bind |
| Auto-start tidak konsisten antar OS | UX buruk | OS-specific startup manager dan acceptance test per OS |
| User salah mengatur route profile | Response tidak sesuai | Default templates, validation, reset config |
| Runtime status tidak sinkron dengan service | User bingung | Health polling dan single source of truth untuk status |

---

## 24. Assumptions

- Client target menggunakan OpenAI-compatible API.
- 9Router dapat menerima request dari Smart AI Proxy.
- 9Router tetap menjadi downstream utama untuk provider routing.
- Smart AI Proxy tidak perlu memanggil provider AI langsung.
- User bersedia menggunakan model virtual seperti `smart-auto`.
- Task classification MVP cukup menggunakan rule-based atau heuristic-based classifier.
- Prompt penuh tidak diperlukan untuk telemetry default.
- Routing combo dapat direpresentasikan melalui model alias, header, atau metadata lain yang kompatibel dengan 9Router.
- User membutuhkan Web UI agar konfigurasi lebih mudah dibanding edit file manual.
- User akan menjalankan Smart AI Proxy secara lokal pada laptop/komputer development.
- Auto-start behavior akan berbeda per sistem operasi dan perlu abstraction layer.

---

## 25. Constraints

- Smart AI Proxy tidak boleh menggantikan fungsi 9Router.
- Harus kompatibel dengan OpenAI API.
- Harus mendukung streaming.
- Harus menambahkan latency seminimal mungkin.
- Harus menjaga privacy by default.
- Policy harus fleksibel dan tidak hardcoded secara permanen.
- MVP tidak boleh terlalu kompleks.
- Tidak boleh menyimpan API key atau prompt penuh di log default.
- Web UI harus local-only secara default.
- Startup manager harus mengikuti batasan OS target.
- Runtime control tidak boleh memutus request aktif secara sembarangan tanpa behavior yang jelas.

---

## 26. Dependencies

### 26.1 External Dependencies

- 9Router downstream service.
- OpenAI-compatible clients.
- AI provider availability melalui 9Router.
- Model alias atau routing profile yang dikenali oleh downstream.
- Operating system startup mechanism.
- Browser lokal untuk mengakses Web UI.

### 26.2 Internal Dependencies

- Routing policy configuration.
- Task classifier rules.
- Local configuration store.
- Web UI frontend.
- Runtime controller.
- Startup manager.
- Logging and observability setup.
- Deployment environment.
- Documentation setup.

---

## 27. Release Plan

### Phase 1: MVP Foundation

Focus:

- OpenAI-compatible proxy.
- Basic forwarding ke 9Router.
- Streaming support.
- `smart-auto` alias.
- Rule-based classifier.
- Basic task-to-route policy.
- Route profile abstraction.
- Safe passthrough.
- Basic logs.
- Local config store.

### Phase 2: Web UI and Runtime Control

Focus:

- Local Web UI.
- Dashboard status.
- Routing settings page.
- Route profile management dasar.
- Startup settings.
- Start/stop/restart control.
- Config import/export.
- Privacy/logging setting.

### Phase 3: Routing Quality Improvement

Focus:

- More task types.
- Confidence score.
- Better policy precedence.
- Manual override improvements.
- Evaluation dataset.
- Compatibility test matrix.

### Phase 4: Observability and Configuration

Focus:

- Dashboard observability lebih detail.
- Policy validation UI.
- Dry-run mode.
- Metrics export.
- Project-specific routing.

### Phase 5: Optimization and Intelligence

Focus:

- Cost-aware routing.
- Latency-aware routing.
- Feedback loop.
- Historical routing analysis.
- Learning-assisted classification.

### Phase 6: Team/Enterprise Readiness

Focus:

- Team policy.
- Workspace isolation.
- RBAC.
- Audit log.
- Centralized settings.
- Enterprise privacy controls.

---

## 28. Open Questions

1. Bagaimana format terbaik agar Smart AI Proxy mengirim route intent ke 9Router: model alias, header, atau metadata khusus?
2. Apakah 9Router akan mengenali `route.*` dan `combo.*` secara langsung atau perlu mapping menjadi model alias?
3. Apakah manual override lebih baik melalui model field, header, request metadata, atau Web UI lock mode?
4. Apakah Web UI wajib masuk MVP penuh, atau MVP cukup minimal settings page terlebih dahulu?
5. Apakah task classifier cukup rule-based pada versi awal?
6. Apakah perlu menyimpan sample prompt untuk evaluasi jika user mengaktifkan debug mode?
7. Bagaimana standard compatibility test untuk OpenCode, Cursor, Cline, dan Continue?
8. Apakah setiap user/team membutuhkan policy berbeda sejak MVP?
9. Sistem operasi mana yang diprioritaskan untuk auto-start: Windows, macOS, Linux, atau semua sejak awal?
10. Apakah Smart AI Proxy akan didistribusikan sebagai CLI, desktop app, background service, Docker container, atau kombinasi?
11. Apakah Web UI perlu authentication sejak MVP meskipun berjalan di localhost?
12. Bagaimana UX terbaik untuk menjelaskan boundary antara Smart AI Proxy dan 9Router kepada user?

---

## 29. Definition of Done

Sebuah fitur dianggap selesai jika:

- Requirement terkait sudah diimplementasikan sesuai acceptance criteria.
- Tidak merusak OpenAI-compatible request/response format.
- Streaming tetap berjalan jika relevan.
- Logging tidak membocorkan data sensitif.
- Error handling jelas.
- Safe fallback tersedia jika terjadi kegagalan.
- Dokumentasi konfigurasi diperbarui.
- Test scenario utama lulus.
- Responsibility boundary dengan 9Router tetap terjaga.
- Untuk fitur Web UI, perubahan setting tersimpan dan diterapkan ke routing behavior.
- Untuk fitur startup, behavior setelah restart sesuai mode yang dipilih user.

---

## 30. Final Product Positioning

Smart AI Proxy adalah product-level intelligent routing layer untuk AI coding workflow. Produk ini membantu developer menggunakan banyak model AI secara otomatis dan efisien tanpa mengganti model secara manual.

Pada revisi v1.1, Smart AI Proxy juga diposisikan sebagai **local Web UI-based control center** untuk mengatur routing AI berdasarkan kategori pekerjaan. User dapat menentukan sendiri task seperti design, backend, frontend, debugging, documentation, architecture, dan low-cost task harus diarahkan ke route profile tertentu.

Smart AI Proxy mengambil keputusan berdasarkan request intent, task classification, dan konfigurasi user. 9Router tetap menjadi downstream infrastructure yang menangani provider-level reliability, failover, load balancing, rotasi akun, fallback model, dan rate limit handling.

Positioning utama:

> Smart AI Proxy memilih route terbaik dan menyediakan control panel untuk user. 9Router mengeksekusi routing provider dengan andal.

Nilai utama produk:

- Satu endpoint untuk banyak AI model.
- Auto-routing berdasarkan jenis pekerjaan.
- Web UI untuk mengatur routing tanpa edit config manual.
- Auto-start saat laptop restart jika user mengaktifkan.
- Runtime mode: Always On, Manual, atau Disabled.
- Kompatibel dengan OpenAI API.
- Cocok untuk OpenCode, Cursor, Cline, Continue, dan tools sejenis.
- Mengurangi manual model switching.
- Mengoptimalkan biaya, kualitas, dan latency.
- Tidak menggantikan 9Router, tetapi memperkuatnya dengan intelligence layer dan control plane.

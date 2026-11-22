# Runbook Operasional

## Prasyarat

- PostgreSQL client tools: `pg_dump`, `pg_restore`, `psql`.
- `jq` dan `curl` untuk smoke check observability.
- File `.env` berisi `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, dan `DB_NAME`.

## Backup

Jalankan dari direktori `Golang-Backend-StarterKit`:

```bash
./scripts/backup-database.sh
```

Backup dibuat dalam format PostgreSQL custom dan diverifikasi dengan
`pg_restore --list`. Simpan file backup ke storage terpisah sebelum deployment
atau perubahan schema.

## Restore

Restore selalu membutuhkan konfirmasi eksplisit. Untuk restore ke database
validasi yang sudah dibuat:

```bash
CONFIRM_RESTORE=YES \
TARGET_DB_NAME=ujian-online-restore-check \
BACKUP_FILE=backups/ujian-online_<timestamp>.dump \
./scripts/restore-database.sh
```

Restore langsung ke database sumber hanya boleh dilakukan pada rollback window:

```bash
CONFIRM_RESTORE=YES \
ALLOW_DESTRUCTIVE_RESTORE=YES \
BACKUP_FILE=backups/ujian-online_<timestamp>.dump \
./scripts/restore-database.sh
```

## Observability Smoke Check

Saat API berjalan, jalankan:

```bash
./scripts/check-observability.sh
```

Check memverifikasi:

- `GET /health` mengembalikan status `ok`.
- `GET /ready` mengembalikan status `ready` dan koneksi database aktif.
- Header dan body membawa `request_id`.
- `logs/server.log` memiliki structured `http_request` log dengan status dan latency.

## Alert Minimum Sebelum Pilot

- Alert jika `/ready` bukan HTTP `200` selama dua pemeriksaan berturut-turut.
- Alert jika HTTP `5xx` atau kegagalan submit melewati threshold kapasitas pilot.
- Alert jika backup gagal, file backup tidak terbentuk, atau restore validation gagal.
- Simpan request ID dari error untuk korelasi dengan `logs/server.log`.

Routing ke Slack, email, Sentry, atau sistem monitoring eksternal masih menjadi
konfigurasi deployment; script smoke check ini menjadi probe lokal atau health
check scheduler.

## Preflight Pilot

Jalankan preflight dari direktori backend sebelum rehearsal atau pilot. Secara
default, script menolak environment `development`, placeholder secret, wildcard
CORS, konfigurasi database placeholder, dan interval expiry yang tidak valid.

Untuk staging:

```bash
EXPECTED_ENVIRONMENT=staging \
REQUIRE_BACKUP=true \
BACKUP_FILE=backups/<backup>.dump \
./scripts/check-pilot-readiness.sh
```

Untuk rehearsal lokal dengan fixture non-produksi:

```bash
ALLOW_DEVELOPMENT_ENV=YES \
./scripts/check-pilot-readiness.sh
```

Preflight memanggil observability smoke check, sehingga `/health`, `/ready`,
request ID, structured log, dan database readiness tetap diverifikasi bersama.

## Rehearsal Pilot Teknis

Rehearsal tidak membuat fixture baru dan tidak menghapus data. Gunakan assignment
khusus rehearsal yang belum pernah disubmit. Script menolak assignment yang sudah
`submitted` atau `expired`.

Dengan token kandidat yang sudah tersedia:

```bash
CONFIRM_PILOT=YES \
ASSIGNMENT_ID=<assignment-id> \
CANDIDATE_TOKEN=<candidate-jwt> \
./scripts/pilot-rehearsal.sh
```

Atau biarkan script melakukan login kandidat:

```bash
CONFIRM_PILOT=YES \
ASSIGNMENT_ID=<assignment-id> \
CANDIDATE_EXAM_NUMBER=<nomor-ujian> \
CANDIDATE_PASSWORD=<password> \
./scripts/pilot-rehearsal.sh
```

Script memverifikasi start berdasarkan assignment, GET ulang untuk resume,
autosave jawaban, submit, duplicate submit, dan result. Rehearsal bisnis tetap
memerlukan persetujuan pemilik ujian serta drill reconnect dan timeout terpisah.

## Export Hasil Admin

Export hasil dilakukan dari halaman admin **Nilai peserta** melalui tombol
**Export CSV**. Filter ujian, sesi, status, dan pencarian yang sedang aktif ikut
diterapkan ke file yang diunduh.

Endpoint langsungnya adalah:

```text
GET /api/v1/admin/results/export
```

Parameter query opsional:

- `exam_id`: batasi ke satu ujian.
- `exam_session_id`: batasi ke satu sesi.
- `status`: `submitted` atau `expired`.
- `search`: nomor peserta, nama peserta, kode/nama ujian.

Export hanya menyertakan attempt yang sudah selesai (`submitted`) atau otomatis
berakhir (`expired`); attempt aktif dan assignment yang belum dikerjakan tidak
ikut diekspor. Kolom yang tersedia meliputi identitas attempt, peserta, ujian,
sesi, waktu mulai/kirim, jumlah benar, dan nilai. Token, password, serta
jawaban plaintext tidak pernah dimasukkan ke CSV.

Export jawaban tersedia melalui tombol **Export Jawaban** pada halaman yang
sama atau endpoint langsung:

```text
GET /api/v1/admin/results/answers/export
```

Endpoint ini memakai filter yang sama dan menghasilkan satu baris per soal,
termasuk identitas peserta, soal, jawaban terpilih, kunci jawaban, status benar,
serta timestamp jawaban. Konten HTML soal dan opsi dipertahankan agar format
rich text tidak hilang saat diproses lebih lanjut.

## Import Peserta CSV

Halaman admin **Peserta ujian** menyediakan tombol **Import CSV**. File harus
memiliki header berikut:

```text
exam_number,password,name,gender,school,class_name
```

Endpoint langsungnya adalah:

```text
POST /api/v1/admin/candidates/import
```

Kirim file sebagai field multipart bernama `file`. Ukuran file dibatasi 5 MiB.
Nilai `gender` harus `L` atau `P`; nomor ujian harus unik; password harus
memiliki panjang 6 sampai 100 karakter. Validasi dilakukan terhadap seluruh
file sebelum batch disimpan, sehingga satu baris invalid tidak meninggalkan
sebagian peserta di database. Password di-hash dan tidak pernah dimasukkan ke
response atau log.

Format Excel belum diaktifkan; konversi workbook ke CSV dapat digunakan sebagai
workaround operasional sementara.

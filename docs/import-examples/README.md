# Contoh CSV Import

Folder ini berisi contoh file untuk fitur import admin:

- `candidates.csv` untuk import peserta.
- `questions.csv` untuk import soal.

## Import peserta

Upload `candidates.csv` dari halaman admin **Peserta ujian** melalui tombol
**Import CSV**. Password pada contoh hanya untuk development dan harus diganti
sebelum digunakan pada data sebenarnya.

Nomor ujian harus unik. Jangan meng-upload file contoh yang sama dua kali tanpa
mengganti `exam_number`, karena peserta duplikat akan ditolak.

## Import soal

Sebelum upload `questions.csv`, ganti semua nilai
`REPLACE_WITH_EXAM_VERSION_ID` dengan ID versi ujian yang aktif. ID tersebut
dapat dilihat dari response daftar ujian admin pada field `current_version_id`
atau dari detail ujian di aplikasi.

Upload file dari halaman admin **Soal dan pilihan jawaban** melalui tombol
**Import CSV**. Setiap baris membutuhkan opsi A dan B, serta tepat satu nilai
`correct_option_key` yang cocok dengan opsi terisi.

Contoh endpoint langsung:

```text
POST /api/v1/admin/candidates/import
POST /api/v1/admin/questions/import
```

Kedua endpoint menerima file sebagai field multipart bernama `file` dan
membatasi ukuran file sampai 5 MiB.

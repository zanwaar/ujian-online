# Load test submit massal

`api_load.js` mendukung tiga mode: `login`, `autosave`, dan `submit`.

Untuk menguji submit pada beberapa attempt secara paralel, siapkan file JSON
dengan satu pasangan token dan attempt per virtual user:

```json
[
  {"token":"candidate-token-1","attemptID":"attempt-id-1"},
  {"token":"candidate-token-2","attemptID":"attempt-id-2"}
]
```

Jalankan k6 dari direktori backend:

```bash
k6 run \
  -e TEST=submit \
  -e SUBMIT_CASES_FILE=tests/load/submit_cases.json \
  -e RATE=20 \
  -e DURATION=30s \
  -e VUS=20 \
  -e MAX_VUS=40 \
  tests/load/api_load.js
```

Jumlah item dalam file sebaiknya sama atau lebih besar dari jumlah VU agar
setiap VU memakai attempt yang berbeda. Jangan menggunakan token atau attempt
produksi; buat fixture khusus load test dan hapus setelah pengujian.

Secara default seluruh VU terlihat berasal dari satu IP sehingga rate limiter
per-IP ikut diuji. Untuk menguji kapasitas submit tanpa berbenturan dengan
limiter satu alamat IP pada lingkungan lokal, tambahkan `-e FORWARD_CLIENT_IP=true`.

# TODO — Integrasi modem live 4G Monitor

Diperbarui: 31 Agustus 2026.

Status: implementasi source selesai; verifikasi dibatasi pemeriksaan statis.
Aplikasi hasil integrasi belum dikompilasi atau dijalankan. Tidak membuat unit
test dan tidak menjalankan build/test sesuai instruksi pengguna.

## Selesai

- [x] GET kedua endpoint saat perencanaan: HTTP 200, JSON dengan Content-Type text/html.
- [x] Hardcode kedua URL lengkap pada internal/modem/client.go.
- [x] Client GET tanpa proxy/redirect/cookie; body dibatasi 1 MiB per endpoint.
- [x] Poll pertama langsung, interval 2 detik, dua GET paralel, timeout siklus 5 detik.
- [x] Lewati tick saat request aktif; batalkan dan tunggu worker saat widget ditutup.
- [x] Store snapshot bersama, publikasi satu siklus utuh, history maksimal 30 sampel.
- [x] Penanganan Loading, Online, No Signal, Disconnected, API Error, dan stale.
- [x] Parsing angka/string, nilai opsional, placeholder —, dan nol yang tetap valid.
- [x] Signal Strength memakai bar modem 0–5; hapus score dan label kualitas fixture.
- [x] RSRQ/SINR diberi label raw pada metrik/grafik; throughput tiga desimal Mbps.
- [x] Ping tetap —; tidak memakai durasi HTTP atau menjalankan speed test.
- [x] Hapus package mock dan seluruh pemilih/label Demo; gunakan Live.
- [x] Pertahankan singleton dashboard, always-on-top, ukuran/DPI, serta Coming soon.
- [x] Perbarui README dan checklist manual untuk implementasi live.
- [x] Pemeriksaan statis source/import, formatting gofmt, dan whitespace Git.

## Belum diverifikasi — pekerjaan pengguna berikutnya

- [ ] Jalankan aplikasi secara manual ketika diizinkan, lalu catat hasil runtime.
- [ ] Periksa ukuran/DPI, keterbacaan widget, scroll, dan singleton dashboard.
- [ ] Bandingkan data live serta konversi throughput dengan respons modem terbaru.
- [ ] Periksa stale, partial failure, No Signal, timeout, dan pemulihan koneksi.
- [ ] Periksa grafik, batas history, dan pembatalan request saat aplikasi keluar.
- [ ] Selesaikan item yang relevan di MANUAL-CHECKLIST.md.

## Batasan yang sengaja dipertahankan

- [ ] Verifikasi makna/konversi nv_rsrq dan nv_sinr dari sumber modem yang tepercaya
      sebelum mengganti label raw menjadi dB. Tidak menebak konversi.
- Ping tidak tersedia pada kedua endpoint; tetap kosong tanpa tambahan probe.
- Pengaturan URL, autentikasi otomatis, persistence, dan kontrol modem di luar scope.
- Tidak melakukan push ke remote kecuali diminta.

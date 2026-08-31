# TODO — Integrasi modem live 4G Monitor

Diperbarui: 31 Agustus 2026.

Status: implementasi source dan cross-build Windows selesai. Aplikasi belum
dijalankan pada Windows dan verifikasi visual/live masih menunggu lingkungan
yang dapat mengakses modem. Tidak menambahkan unit test.

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
- [x] Verifikasi checksum dependency dengan `go mod verify`.
- [x] Cross-build Windows amd64 biasa dan `windowsgui` dengan CGO nonaktif.
- [x] Jalankan `go vet ./...` untuk target Windows amd64 tanpa temuan.
- [x] Type-check package non-window melalui `go test`; tidak ada file unit test.

## Belum diverifikasi — pekerjaan pengguna berikutnya

- [ ] Jalankan executable pada Windows, lalu catat hasil runtime.
- [ ] Periksa ukuran/DPI, keterbacaan widget, scroll, dan singleton dashboard.
- [ ] Bandingkan data live serta konversi throughput dengan respons modem terbaru.
- [ ] Periksa stale, partial failure, No Signal, timeout, dan pemulihan koneksi.
- [ ] Periksa grafik, batas history, dan pembatalan request saat aplikasi keluar.
- [ ] Selesaikan item yang relevan di MANUAL-CHECKLIST.md.

## Catatan verifikasi 31 Agustus 2026

- Cross-build menghasilkan executable PE32+ x86-64 dengan Go 1.27.0.
- Build target Linux tidak dapat diselesaikan karena image tidak menyediakan
  development package `xkbcommon` dan `wayland-client` yang dibutuhkan Gio.
  Ini merupakan batasan environment Linux, bukan error kompilasi source target
  Windows.
- Endpoint `192.168.100.1:80` tidak dapat dijangkau dari container verifikasi,
  sehingga respons modem terbaru dan perilaku polling live belum diuji ulang.
- Aplikasi GUI Windows tidak dapat dijalankan di container Linux ini; seluruh
  item visual, DPI, lifecycle window, dan shutdown tetap membutuhkan UAT.

## Batasan yang sengaja dipertahankan

- [ ] Verifikasi makna/konversi nv_rsrq dan nv_sinr dari sumber modem yang tepercaya
      sebelum mengganti label raw menjadi dB. Tidak menebak konversi.
- Ping tidak tersedia pada kedua endpoint; tetap kosong tanpa tambahan probe.
- Pengaturan URL, autentikasi otomatis, persistence, dan kontrol modem di luar scope.
- Tidak melakukan push ke remote kecuali diminta.

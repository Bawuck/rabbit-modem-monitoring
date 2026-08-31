# TODO — Integrasi modem live 4G Monitor

Diperbarui: 31 Agustus 2026.

Status: implementasi source dan cross-build Windows selesai. Aplikasi belum
dijalankan pada Windows dan verifikasi visual/live masih menunggu lingkungan
yang dapat mengakses modem. Tidak menambahkan unit test.

## Selesai

- [x] Toggle Start on startup per-user HKCU Run, perubahan langsung, tanpa restart polling.
- [ ] Buat EXE portable setelah izin build eksplisit; auto-review menolak build karena larangan build/test.
- [ ] Verifikasi runtime toggle dan login Windows; registry startup belum diubah pada sesi implementasi.

- [x] Form Koneksi Modem, host tervalidasi, password masked, tombol Pengaturan tanpa sidebar.
- [x] Profil versi 1 dengan DPAPI current-user dan penggantian file atomik; gagal simpan mempertahankan koneksi.
- [x] Pergantian worker serial, reset snapshot/history, cookie jar baru, tanpa restart.
- [ ] UAT form, DPAPI save/load, pergantian koneksi, dan shutdown; tanpa build/test pada implementasi ini.

- [x] Card detail modem: operator, roaming, counter perangkat terpisah, total RX/TX, dan durasi raw; satu GET 20 field.
- [ ] Verifikasi satuan realtime_time dan pembagian sta_count/m_sta_count sebelum memberi label durasi atau jumlah perangkat gabungan.

- [x] GET kedua endpoint saat perencanaan: HTTP 200, JSON dengan Content-Type text/html.
- [x] Satu URL monitoring dari konfigurasi dengan query gabungan 20 field; GET live berhasil HTTP 200.
- [x] Client tanpa proxy/redirect; cookie jar in-memory untuk login; body dibatasi 1 MiB per endpoint.
- [x] Poll pertama langsung, interval 2 detik, satu GET gabungan, timeout request 5 detik (pemulihan login maksimal 15 detik).
- [x] Lewati tick saat request aktif; batalkan dan tunggu worker saat widget ditutup.
- [x] Store snapshot bersama, publikasi satu siklus utuh, history maksimal 30 sampel.
- [x] Penanganan Loading, Online, No Signal, Disconnected, API Error, dan stale.
- [x] Parsing angka/string, nilai opsional, placeholder —, dan nol yang tetap valid.
- [x] Signal Strength memakai bar modem 0–5; hapus score dan label kualitas fixture.
- [x] RSRQ/SINR diberi label raw pada metrik/grafik; throughput tiga desimal Mbps.
- [x] Hapus Ping dari widget/dashboard; trafik hanya Download dan Upload.
- [x] Hapus package mock dan seluruh pemilih/label Demo; gunakan Live.
- [x] Pertahankan singleton dashboard, always-on-top, ukuran/DPI; hapus sidebar agar Overview memenuhi lebar halaman.
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
- Tidak menambahkan probe ping atau speed test.
- Profil host/password disimpan via DPAPI; history/cookie tidak dipersist dan kontrol modem tetap di luar scope.
- Tidak melakukan push ke remote kecuali diminta.

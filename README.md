# 4G Monitor

Aplikasi desktop Windows dengan Go + Gio: widget always-on-top, satu dashboard
Overview, dan data modem live dari `192.168.100.1`. Semua UI ditulis dalam Go;
tidak menggunakan HTML, CSS, JavaScript, Electron, atau WebView.

**Status: implementasi source dan pemeriksaan statis. Build dan aplikasi hasil
integrasi belum dijalankan; tidak membuat atau menjalankan unit test.**

## Menjalankan

Prasyarat: Windows 10/11, driver grafis yang mendukung Gio, Go 1.27.0 atau lebih
baru, dan akses LAN ke modem. Internet diperlukan untuk mengunduh dependency
pertama kali. Gio tetap dikunci ke `v0.10.2`.

Jalankan sendiri dari PowerShell:

```powershell
cd C:\Go\rabbit-monitoring
go run ./cmd/4g-monitor
```

Untuk executable GUI tanpa console, gunakan opsi linker
`go run -ldflags="-H windowsgui" ./cmd/4g-monitor`.
Perintah tersebut mengompilasi dan menjalankan aplikasi; tidak dijalankan selama
implementasi ini.

## Window dan tampilan

- Widget: konten tetap 300 × 380 dp, always-on-top, title bar native.
- Dashboard: konten awal 900 × 650 dp, minimum 760 × 520 dp, dapat diperbesar
  dan digulir; grafik tersusun vertikal pada area konten yang lebih sempit.
- Klik konten widget membuka dashboard; klik berikutnya memfokuskan/memulihkan
  instance yang sama. Enter/Space pada widget yang fokus juga dapat membukanya.
- Menutup dashboard tidak menghentikan widget atau polling. Menutup widget
  membatalkan request aktif, menghentikan worker/ticker, dan menutup dashboard.
- Kedua window berlabel Live. Tidak ada mode atau pemilih skenario Demo.
- Signal, Cell, History, dan Settings tetap nonaktif dengan Coming soon.
- Setiap window memiliki shaper, state interaksi, dan event loop sendiri.
  Window mengambil satu salinan snapshot per frame; repaint tidak harus serentak.

## Endpoint dan pemetaan

Dua URL GET lengkap hardcoded sebagai `SignalURL` dan `StatusURL` di
[internal/modem/client.go](internal/modem/client.go). Keduanya mempertahankan
host, path `/goform/goform_get_cmd_process`, dan seluruh query dari pengguna.
Tidak ada pengaturan URL, autentikasi otomatis, cookie, atau proxy sistem;
redirect tidak diikuti. Tidak ada request untuk mengubah konfigurasi modem.

| Tampilan | Field | Perlakuan |
| --- | --- | --- |
| Network | network_type, sub_network_type | FDD_LTE/TDD_LTE menjadi LTE; LTE-A hanya bila dinyatakan eksplisit |
| Signal Strength | signalbar | Integer 0–5, ditampilkan sebagai bar dan n/5; tanpa score buatan |
| RSRP / RSSI | lte_rsrp / rssi | dBm; tidak ada konversi tanda tanpa bukti |
| RSRQ / SINR | nv_rsrq / nv_sinr | Raw, termasuk unit sumbu grafik; konversi dB belum terverifikasi |
| Band / PCI | lte_band / nv_pci | Band numerik menjadi Bn; PCI LTE 0–503 |
| Download / Upload | realtime_rx_thrpt / realtime_tx_thrpt | Bytes/detik × 8 ÷ 1.000.000; Mbps dengan tiga desimal |
| Ping | Tidak disediakan | Selalu —, dengan keterangan No API data |
| Updated | Waktu penerimaan siklus sukses | Dibekukan selama data stale |

Trafik adalah aktivitas modem saat itu, bukan speed test. Durasi GET bukan ping.
Metrik khusus LTE disembunyikan bila jenis jaringan bukan LTE/LTE-A.
Nilai opsional kosong/null/tidak valid menjadi `—`; angka nol tetap valid.
ICCID, SSID, dan field tak terpakai tidak masuk model, log, atau penyimpanan.
Body respons hanya dibaca sementara untuk decoding, maksimal 1 MiB per endpoint.

## Polling dan status

Poll pertama langsung saat startup, lalu setiap 2 detik. Satu worker menjalankan
dua GET paralel dengan deadline siklus 5 detik. Tick selama siklus masih berjalan
dilewati; tidak ada antrean polling. HTTP tidak berjalan pada event loop UI.

Kedua respons harus berhasil sebelum satu snapshot dipublikasikan. Data radio
berasal dari endpoint pertama; status PPP/bar/throughput dari endpoint kedua.
Kedua GET bukan transaksi atomik di modem, tetapi publikasi di aplikasi atomik.

| Status | Kondisi | Tampilan |
| --- | --- | --- |
| Loading | Belum ada hasil siklus pertama | Placeholder |
| Online | PPP terhubung dan jaringan tersedia | Pengukuran live dan sampel baru |
| No Signal | Jaringan kosong/no service/limited service | Pengukuran dan grafik disembunyikan |
| Disconnected | Transport gagal/timeout, atau PPP belum terhubung | Snapshot Online terakhir berlabel stale |
| API Error | HTTP gagal/redirect, HTML login, JSON rusak/terlalu besar, status wajib hilang/tidak dikenal | Snapshot Online terakhir berlabel stale |

Struktur wajib: network_type berupa string (boleh kosong untuk No Signal) dan
ppp_status berupa string yang dikenal. Status PPP yang didukung: ppp_connected,
ppp_connecting, ppp_disconnecting, dan ppp_disconnected.

Saat gagal sebelum Online pertama, nilai tetap `—`. Cache sukses tetap tersedia
setelah No Signal untuk kegagalan berikutnya. Usia data bertambah saat stale;
timestamp pengukuran tidak berubah. Polling terus mencoba otomatis.

History hanya menyimpan 30 sampel Online terbaru dalam memori (sekitar 1 menit
bila respons lancar). Nilai hilang memutus garis. Kembali ke Online memulai
segmen baru; tidak menghubungkan garis melewati outage.

## Struktur dan batasan

- `internal/modem`: client HTTP, pemetaan field, worker polling.
- `internal/monitor`: store tersinkronisasi dan history terbatas.
- `internal/model`: tipe data, nilai opsional, status, update/snapshot.
- `internal/windows`: lifecycle dua window dan pembatalan worker.
- `internal/components`, `internal/pages`: UI Gio dan grafik.
- `cmd/4g-monitor`: entry point.

Tidak ada database, persistence, system tray, autostart, speed test, rumus score,
atau kontrol modem. HTTP modem tidak terenkripsi; gunakan hanya pada LAN tepercaya.

## Verifikasi dan pekerjaan lanjutan

Pada sesi perencanaan 31 Agustus 2026, GET kedua endpoint memperoleh HTTP 200
tanpa cookie/login tambahan, dengan JSON ber-Content-Type text/html.
JavaScript modem memformat throughput bytes/detik menjadi bit/detik dengan ×8.
Temuan tersebut tidak membuktikan runtime client Go atau UI hasil integrasi.

Pemeriksaan implementasi terbatas pada parsing/format `gofmt`, penelusuran
import dan source, review alur data/lifecycle, serta `git diff --check`.
Tidak menjalankan `go build`, `go run`, `go test`, `go vet`, atau type-check penuh.

Lihat [TODO.md](TODO.md) untuk status pekerjaan dan
[MANUAL-CHECKLIST.md](MANUAL-CHECKLIST.md) untuk verifikasi pengguna.

Referensi: [Gio Windows](https://gioui.org/doc/install/windows),
[TopMost](https://pkg.go.dev/gioui.org@v0.10.2/app#TopMost),
[formatter modem](http://192.168.100.1/js/lib.js).

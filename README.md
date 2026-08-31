# Rabbit Modem Monitoring

Aplikasi desktop Windows dengan Go + Gio: widget always-on-top, satu dashboard
Overview, dan data modem live dari `192.168.100.1`. Semua UI ditulis dalam Go;
tidak menggunakan HTML, CSS, JavaScript, Electron, atau WebView.

Perangkat yang didukung: **modem Rabbit CP XR**. Kompatibilitas dengan tipe modem
lain belum diverifikasi.

**Status: form konfigurasi dan penyimpanan Windows DPAPI telah diimplementasikan.
Perubahan terbaru belum dijalankan di GUI atau di-build/test; pemeriksaan terbatas pada source, formatting, dan diff.**

## Menjalankan

Prasyarat: Windows 10/11, driver grafis yang mendukung Gio, Go 1.27.0 atau lebih
baru, dan akses LAN ke modem. Internet diperlukan untuk mengunduh dependency
pertama kali. Gio tetap dikunci ke `v0.10.2`.

Jalankan sendiri dari PowerShell:

```powershell
cd C:\Go\rabbit-monitoring
go run .
```

Untuk menjalankan GUI tanpa console melalui Go, gunakan opsi linker
`go run -ldflags="-H windowsgui" .`.
Cross-build varian biasa dan `windowsgui` sudah diverifikasi; eksekusi serta UAT
tetap harus dilakukan langsung pada Windows yang terhubung ke modem.

## EXE portable dan startup Windows

Aplikasi dapat dikemas sebagai satu EXE GUI tanpa installer atau Go di komputer tujuan.
Go hanya diperlukan di komputer yang membuat EXE. Dari PowerShell pada Windows:

```powershell
cd C:\Go\rabbit-monitoring
go build -trimpath -ldflags="-H windowsgui -s -w" -o "Rabbit Monitoring widget.exe" .
```

Hasilnya `C:\Go\rabbit-monitoring\Rabbit Monitoring widget.exe`. Klik dua kali file
tersebut untuk membuka aplikasi, atau jalankan dari PowerShell:

```powershell
& ".\Rabbit Monitoring widget.exe"
```

Opsi `-H windowsgui` membuat aplikasi tanpa console; `-s -w` menghapus informasi
debug untuk mengurangi ukuran EXE. Tutup aplikasi sebelum membuat ulang EXE pada
lokasi yang sama. Perintah ini merupakan panduan; build terbaru belum diverifikasi.

Salin EXE ke lokasi tetap pada komputer Windows tujuan; tidak perlu installer,
source code, atau Go. Driver grafis yang kompatibel dan akses jaringan ke modem
tetap diperlukan. Konfigurasi disimpan di LOCALAPPDATA, bukan di sebelah EXE.
Isi konfigurasi pada komputer tujuan; password DPAPI tidak portabel antar pengguna Windows.

Untuk menjalankan otomatis saat login Windows:

1. Buka EXE dari lokasi tetap.
2. Buka **Pengaturan** melalui dashboard atau tombol di kanan **Open Overview** pada widget.
3. Aktifkan toggle **Start on startup**. Matikan toggle untuk menonaktifkannya.

Toggle **Start on startup** tersedia di Pengaturan. Perubahan langsung diterapkan,
tidak memerlukan Simpan & Hubungkan, dan tidak dibatalkan oleh tombol Batal form koneksi.
Default nonaktif jika belum ada entri. Toggle mendaftarkan/menghapus hanya value
`Rabbit Modem Monitoring` pada `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`.
Tidak perlu administrator. Entri berisi path EXE ber-quote, tanpa password atau parameter rahasia.

Startup berarti saat pengguna login Windows, bukan service sebelum login atau auto-restart
setelah aplikasi ditutup. Pengaturan Windows Startup Apps dapat menonaktifkan peluncuran ini.
Jalankan EXE dari lokasi tetap sebelum mengaktifkan toggle; go run/debug executable sementara
ditolak. Jika EXE dipindah/diubah namanya, aktifkan ulang toggle dari lokasi baru.
Matikan toggle sebelum menghapus EXE. Tidak ada pemasangan otomatis saat aplikasi sekadar dibuka.

## Koneksi modem

Saat konfigurasi belum tersedia, dashboard membuka form Koneksi Modem otomatis.
Form awal terisi host `http://192.168.100.1` dan password `admin` (tetap tersamarkan).
Konfigurasi tersimpan tetap diutamakan. Sesuaikan bila perlu, lalu pilih
**Simpan & Hubungkan**. Host tanpa skema memakai HTTP. IP/hostname dan port opsional
didukung; path API, query, fragment, dan kredensial dalam URL ditolak.
Password tersamarkan, dengan tombol tampil/sembunyikan. Password tidak di-trim.

Host dan password DPAPI tersimpan di `%LOCALAPPDATA%\4G Monitor\connection.json`
(format versi 1). Hanya ciphertext password disimpan; cookie hanya di memori.
DPAPI memakai pengguna Windows saat ini. File rusak/tidak dapat didekripsi tidak
ditimpa otomatis: form meminta pengisian ulang. File sementara berisi ciphertext,
lalu diganti atomik. Gagal menyimpan mempertahankan koneksi sebelumnya.

Tombol **Pengaturan** di header dashboard membuka kembali form, tanpa sidebar.
**Batal** membuang perubahan tanpa menghentikan polling. Saat disimpan, worker lama
dibatalkan dan ditunggu, snapshot/history direset, lalu client/cookie jar baru dibuat.
Tidak perlu restart. Konfigurasi tetap disimpan meskipun modem offline/login ditolak.
Startup berikutnya langsung memakai konfigurasi tersimpan; environment password tidak dipakai lagi.

HTTP mengirim password tanpa enkripsi jaringan; Base64 bukan enkripsi. Gunakan LAN
tepercaya. HTTPS memakai verifikasi sertifikat standar, tanpa opsi insecure.

## Window dan tampilan

- Widget: konten tetap 300 × 380 dp, always-on-top, header terintegrasi dengan area drag dan tombol close.
- Dashboard: konten awal 900 × 650 dp, minimum 760 × 520 dp, dapat diperbesar
  dan digulir; grafik tersusun vertikal pada area konten yang lebih sempit.
- Klik konten widget membuka dashboard; klik berikutnya memfokuskan/memulihkan
  instance yang sama. Enter/Space pada widget yang fokus juga dapat membukanya.
- Menutup dashboard tidak menghentikan widget atau polling. Menutup widget
  membatalkan request aktif, menghentikan worker/ticker, dan menutup dashboard.
- Kedua window berlabel Live. Tidak ada mode atau pemilih skenario Demo.
- Dashboard tanpa sidebar dan tanpa title bar bawaan Windows. Header menyatu dengan halaman,
  tetap terlihat saat scroll/form terbuka, dengan Pengaturan, minimize, maximize/restore, dan close.
  Drag area judul untuk memindahkan window; widget juga memakai header terintegrasi dengan minimize, maximize/restore, dan close.
- Setiap window memiliki shaper, state interaksi, dan event loop sendiri.
  Window mengambil satu salinan snapshot per frame; repaint tidak harus serentak.

## Endpoint dan pemetaan

URL GET dibangun dari host konfigurasi di internal/modem/client.go menggunakan
path `/goform/goform_get_cmd_process`, `multi_data=1`, dan satu query `cmd` gabungan.
Sebanyak 20 field termasuk status sesi `loginfo` diminta: jaringan, sinyal, band/PCI, PPP, bar, throughput, serta detail operator dan counter modem.
Flag SMS/STS tetap 0. POST LOGIN otomatis memakai password dari form/profil DPAPI,
diubah menjadi Base64 lalu form-urlencoded sesuai protokol modem. Origin dan Referer mengikuti host.
Cookie disimpan hanya di memori jika modem mengirimkannya. Tidak ada proxy sistem,
redirect, logout, atau perubahan konfigurasi modem. HTTP/Base64 tidak mengenkripsi password;
gunakan hanya di LAN tepercaya. Jangan commit password atau menuliskannya ke log.

| Tampilan | Field | Perlakuan |
| --- | --- | --- |
| Network | network_type, sub_network_type | FDD_LTE/TDD_LTE menjadi LTE; LTE-A hanya bila dinyatakan eksplisit |
| Signal Strength | signalbar | Integer 0–5, ditampilkan sebagai bar dan n/5; tanpa score buatan |
| RSRP / RSSI | lte_rsrp / rssi | dBm; tidak ada konversi tanda tanpa bukti |
| RSRQ / SINR | nv_rsrq / nv_sinr | Raw, termasuk unit sumbu grafik; konversi dB belum terverifikasi |
| Band / PCI | lte_band / nv_pci | Band numerik menjadi Bn; PCI LTE 0–503 |
| Download / Upload | realtime_rx_thrpt / realtime_tx_thrpt | Bytes/detik × 8 ÷ 1.000.000; Mbps dengan tiga desimal |
| Operator / Roaming | network_provider / simcard_roam | Teks modem, kosong menjadi — |
| Counter perangkat | sta_count / m_sta_count | Ditampilkan terpisah; pembagian belum terverifikasi |
| Total download / upload | realtime_rx_bytes / realtime_tx_bytes | GB desimal dan bytes; periode reset mengikuti modem |
| Counter durasi | realtime_time | Raw; satuan belum terverifikasi, tidak dihitung maju saat stale |
| Updated | Waktu penerimaan siklus sukses | Dibekukan selama data stale |

Trafik adalah aktivitas modem saat itu, bukan speed test. Durasi GET bukan ping.
Metrik khusus LTE disembunyikan bila jenis jaringan bukan LTE/LTE-A.
Nilai opsional kosong/null/tidak valid menjadi `—`; angka nol tetap valid.
ICCID, SSID, dan field tak terpakai tidak masuk model, log, atau penyimpanan.
Body respons hanya dibaca sementara untuk decoding, maksimal 1 MiB per endpoint.

## Polling dan status

Poll pertama langsung setelah konfigurasi tersedia, lalu setiap 2 detik. Satu worker menjalankan
satu GET dengan timeout request 5 detik. Tick selama siklus masih berjalan
dilewati; tidak ada antrean polling. HTTP tidak berjalan pada event loop UI.

Data radio dan status PPP/bar/throughput didekode dari satu respons JSON.
Respons harus berhasil dan status wajib valid sebelum snapshot dipublikasikan.
Publikasi di aplikasi tetap atomik. Jika `loginfo` bukan `ok` atau GET mendapat
401/403/redirect, lakukan satu POST LOGIN lalu satu GET ulang. Jalur pemulihan
dibatasi total 15 detik; polling normal tetap satu GET. Nilai sinyal kosong saja
tidak memicu login selama sesi `ok`. Respons tanpa sesi tidak menggantikan cache;
data sebelumnya menjadi stale jika pemulihan gagal.
Login dibatasi satu percobaan per 60 detik. Hasil login nonzero atau 401/403
menghentikan percobaan login otomatis. Perbaiki konfigurasi melalui Pengaturan lalu
Simpan & Hubungkan untuk membuat client baru dan mencoba lagi tanpa restart.

| Status | Kondisi | Tampilan |
| --- | --- | --- |
| Koneksi belum diatur | Belum ada profil valid | Tidak melakukan polling |
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
- `main.go`: entry point di root project.

Hanya profil koneksi disimpan; history dan cookie tidak dipersist.
Tidak ada database, system tray, speed test, rumus score,
atau kontrol modem. HTTP modem tidak terenkripsi; gunakan hanya pada LAN tepercaya.

## Verifikasi dan pekerjaan lanjutan

Verifikasi session: POST LOGIN yang diberikan pengguna menghasilkan HTTP 200/result 0.
GET berikutnya menghasilkan loginfo=ok dan RSSI/RSRP -63; tidak ada cookie yang
dikirim pada percobaan ini. Alur Go otomatis dan expiry belum diuji runtime;
tidak ada build/test yang dijalankan untuk perubahan ini.

Perubahan satu request: GET URL gabungan dari Windows berhasil HTTP 200 dengan seluruh
12 field, termasuk RSSI/RSRP -63, PPP connected, dan throughput RX/TX.
Perubahan ini diperiksa dengan gofmt dan git diff --check, tanpa build/test atau
menjalankan GUI. Catatan build di bawah merupakan verifikasi versi sebelumnya.

Pada sesi perencanaan 31 Agustus 2026, GET kedua endpoint memperoleh HTTP 200
tanpa cookie/login tambahan, dengan JSON ber-Content-Type text/html.
JavaScript modem memformat throughput bytes/detik menjadi bit/detik dengan ×8.
Temuan tersebut tidak membuktikan runtime client Go atau UI hasil integrasi.

Pemeriksaan lanjutan 31 Agustus 2026 mencakup `gofmt`, `git diff --check`,
`go mod verify`, cross-build Windows amd64 dengan CGO nonaktif (biasa dan
`windowsgui`), serta `go vet ./...` untuk target Windows. Package non-window juga
berhasil di-type-check oleh `go test` dan tidak memiliki file unit test.

Build/test native Linux berhenti pada dependency sistem Gio yang tidak tersedia,
yaitu `xkbcommon` dan `wayland-client`. Ini tidak memengaruhi cross-build target
Windows. Container tidak dapat menghubungi `192.168.100.1`, sehingga aplikasi
GUI, data modem terbaru, polling live, DPI, dan lifecycle window masih memerlukan
verifikasi manual pada Windows.

Lihat [TODO.md](TODO.md) untuk status pekerjaan dan
[MANUAL-CHECKLIST.md](MANUAL-CHECKLIST.md) untuk verifikasi pengguna.

Referensi: [Gio Windows](https://gioui.org/doc/install/windows),
[TopMost](https://pkg.go.dev/gioui.org@v0.10.2/app#TopMost),
[formatter modem](http://192.168.100.1/js/lib.js).

# Checklist manual — integrasi modem live

Cross-build Windows amd64 biasa dan `windowsgui` sudah berhasil, tetapi aplikasi
hasil integrasi belum dijalankan pada Windows. Checklist ini untuk pengguna saat
menjalankan aplikasi; kasus respons khusus memerlukan lingkungan terkontrol,
bukan perubahan konfigurasi atau gangguan sengaja pada modem produksi. Container
verifikasi tidak dapat menjangkau `192.168.100.1`, jadi tidak ada item live atau
visual di bawah ini yang ditandai selesai hanya berdasarkan pemeriksaan source.

## Window dan DPI

- [ ] Startup menampilkan satu widget Live; Loading menunggu hasil GET pertama.
- [ ] Konten widget 300 × 380 dp; title bar native berada di luar area konten.
- [ ] Scaling 100%, 125%, 150%, dan 200% tetap terbaca tanpa tumpang tindih.
- [ ] Widget selalu di atas aplikasi biasa dan dapat dipindah melalui title bar.
- [ ] Klik konten atau Enter/Space membuka dashboard 900 × 650 dp.
- [ ] Klik cepat berulang tidak membuat dashboard duplikat.
- [ ] Dashboard yang diminimalkan dipulihkan saat widget diklik.
- [ ] Menutup dashboard menyisakan widget dan polling tetap berjalan.
- [ ] Membuka kembali dashboard menampilkan snapshot/history yang sama.
- [ ] Sidebar Live tetap terlihat saat Overview digulir.
- [ ] Pada lebar minimum 760 dp, grafik tersusun vertikal tanpa tumpang tindih.
- [ ] Sidebar tidak tampil; konten Overview memakai seluruh lebar dan tetap dapat digulir.
- [ ] Tidak ada pemilih skenario, label Demo, score 0–100, atau label kualitas fixture.

## Data live dan unit

- [ ] Band/PCI/network/bar cocok dengan respons modem pada waktu yang sama.
- [ ] FDD_LTE/TDD_LTE tampil LTE; tidak otomatis diubah menjadi LTE-A.
- [ ] Signal Strength menampilkan signalbar sebagai n/5 dengan lima segmen.
- [ ] Bar 0 tetap valid; bar kosong, negatif, pecahan, atau >5 menjadi —.
- [ ] RSRP dan RSSI memakai dBm; RSRQ dan SINR menampilkan nilai raw tanpa konversi.
- [ ] Label raw terlihat pada kartu metrik dan panel grafik RSRQ/SINR.
- [ ] Download/upload = bytes/detik ×8 ÷1.000.000, tampil tiga desimal Mbps.
- [ ] Throughput nol tampil 0.000; kosong/null/tidak valid tampil —.
- [ ] Trafik dilabeli aktivitas modem saat ini, bukan hasil speed test.
- [ ] Statistik trafik hanya Download dan Upload, tanpa Ping, pada widget dan dashboard.
- [ ] RSSI ditampilkan di Overview; seluruh metrik widget tetap tersedia.
- [ ] Metrik khusus LTE kosong saat jaringan bukan LTE/LTE-A.

## Polling dan error

- [ ] Poll pertama langsung; selanjutnya tiap 2 detik bila siklus sebelumnya selesai.
- [ ] Normalnya satu GET gabungan per siklus; pemulihan sesi maksimal GET + LOGIN + GET; siklus lambat tidak menumpuk request.
- [ ] Request yang macet dibatasi deadline siklus sekitar 5 detik.
- [ ] UI, pemindahan window, dan scroll tetap responsif selama request berlangsung.
- [ ] Snapshot dua window konsisten setelah masing-masing repaint.
- [ ] Request gabungan gagal: snapshot lama tetap stale tanpa campuran nilai baru.
- [ ] JSON valid ber-Content-Type text/html tetap diterima.
- [ ] HTML login, HTTP gagal/redirect, JSON rusak, {}, atau body >1 MiB menjadi API Error.
- [ ] Status network/PPP wajib yang hilang, null, salah tipe, atau PPP tak dikenal menjadi API Error.
- [ ] No service/limited service/network kosong menjadi No Signal, metrik dan chart kosong.
- [ ] PPP belum terhubung dengan jaringan tersedia menjadi Disconnected.
- [ ] Timeout/transport gagal menjadi Disconnected, lalu mencoba otomatis.
- [ ] Online → API Error/Disconnected mempertahankan data, history, dan timestamp dengan stale.
- [ ] Gagal sebelum Online pertama tetap — tanpa timestamp pengukuran palsu.
- [ ] No Signal → API Error masih dapat menunjukkan cache Online terakhir sebagai stale.
- [ ] Pemulihan menjadi Online menghilangkan stale dan memulai segmen history baru.

## Grafik, shutdown, dan privasi

- [ ] Grafik RSRP, RSRQ, SINR terpisah, memiliki unit serta timestamp awal/akhir.
- [ ] Sampel pertama berupa titik; nilai hilang memutus garis, bukan menjadi nol.
- [ ] History maksimal 30 sampel; state non-Online tidak menambahkan sampel.
- [ ] Menutup widget saat request aktif membatalkan request dan mengakhiri proses.
- [ ] Penutupan juga berhasil saat dashboard terbuka, diminimalkan, atau sedang dibuat.
- [ ] Tidak ada request lanjutan setelah aplikasi keluar.
- [ ] Tidak ada body mentah, ICCID, SSID, cookie, atau token dalam log/file.
- [ ] Tidak ada request perubahan konfigurasi modem, database, tray, atau speed test.
- [ ] Membuka ulang aplikasi memulai history baru tanpa data tersimpan.

## Detail modem tambahan

- [ ] Operator, roaming, dua counter perangkat, total RX/TX, dan realtime_time sesuai respons yang sama.
- [ ] Counter kosong/null/negatif/tidak valid tampil —; nol tetap valid.
- [ ] Seluruh detail mempertahankan nilai terakhir dengan label stale saat request gagal.
- [ ] Card detail dapat digulir pada ukuran dashboard minimum; widget tidak berubah.
- [ ] Satu GET per polling meminta 20 field, tanpa ICCID/SSID.

## Session modem

- [ ] Session aktif (loginfo=ok): hanya satu GET, termasuk jika RSSI/RSRP kosong.
- [ ] Session habis: LOGIN lalu GET ulang; snapshot hanya diganti setelah session valid.
- [ ] Konfigurasi belum diatur: tampilkan form, tanpa GET/POST.
- [ ] Login ditolak: hentikan POST sampai Simpan & Hubungkan atau restart; error transport dibatasi retry 60 detik.
- [ ] Timeout/cancel ketika login tidak menumpuk worker dan widget tetap dapat ditutup.
- [ ] Cookie/password/response login tidak masuk log, file, atau UI; cookie hanya di memori.

## Form koneksi dan DPAPI

- [ ] Startup pertama membuka form otomatis; widget Koneksi belum diatur dan tanpa polling.
- [ ] Host tanpa skema, HTTP/HTTPS, hostname/IP/port valid diterima; path/query/fragment/userinfo/port invalid ditolak.
- [ ] Password wajib, tidak di-trim, tersamarkan; tampil/sembunyikan bekerja. HTTP menampilkan peringatan.
- [ ] Batal membuang draft; monitoring sebelumnya terus berjalan saat form terbuka.
- [ ] Simpan & Hubungkan menonaktifkan submit selama proses; UI tetap responsif.
- [ ] File connection.json hanya berisi versi, host, dan ciphertext DPAPI; tidak ada plaintext password/cookie.
- [ ] Restart memuat profil dan mulai polling; profil rusak atau DPAPI gagal membuka form tanpa menimpa file otomatis.
- [ ] Gagal simpan/ganti file mempertahankan worker dan konfigurasi sebelumnya.
- [ ] Ganti host saat request aktif membatalkan worker lama; snapshot/history/cookie lama tidak terbawa.
- [ ] Nama host pada dashboard sesuai snapshot, termasuk selama pergantian koneksi.
- [ ] Password ditolak dapat diperbaiki dan dicoba ulang melalui Pengaturan tanpa restart.
- [ ] Menutup widget selama penerapan menghentikan coordinator dan worker; tidak ada request tersisa.
- [ ] Profil tetap tersimpan jika modem offline/login gagal; tidak ada password pada error/log/snapshot.

## Header dashboard terintegrasi

- [ ] Title bar native dashboard tidak muncul; header menyatu dengan warna halaman tanpa duplikasi judul.
- [ ] Drag area judul memindahkan window; resize tepi dan ukuran minimum tetap bekerja.
- [ ] Minimize, maximize/restore (termasuk perubahan dari OS), dan close bekerja.
- [ ] Header tetap terlihat ketika dashboard digulir atau form Pengaturan dibuka.
- [ ] Close dashboard tidak menghentikan widget/polling; klik widget dapat membuka dashboard lagi.
- [ ] Widget always-on-top memakai header terintegrasi: drag judul tidak membuka Overview; close menghentikan aplikasi.
- [ ] Widget tetap 300 × 380 dp; RSSI dan semua metrik/footer terlihat tanpa header ganda.
- [ ] Klik isi widget membuka Overview; drag/close tetap bekerja sebelum koneksi diatur.
- [ ] Pengaturan di kanan Open Overview pada widget membuka form langsung, baik dashboard tertutup, terbuka, maupun diminimalkan.
- [ ] Klik Pengaturan tidak ikut memicu klik body/Open Overview; draft form yang sudah terbuka tidak direset.
- [ ] Kedua tombol footer tetap terlihat saat koneksi belum diatur, tanpa menutupi metrik widget.

## EXE portable dan toggle startup

- [ ] EXE GUI dapat dibuka tanpa installer/Go; tidak ada console tambahan.
- [ ] Toggle langsung menulis/hapus hanya entri startup aplikasi pada HKCU, tanpa restart polling.
- [ ] Toggle default OFF; status bertahan setelah aplikasi dibuka ulang.
- [ ] Path EXE dengan spasi tersimpan ber-quote; tidak ada password pada registry command.
- [ ] Gagal akses registry menampilkan error dan mengembalikan nilai toggle sebelumnya.
- [ ] Startup ON membuka widget setelah login Windows; OFF tidak meluncurkan aplikasi.
- [ ] Menjalankan via go run tidak boleh mendaftarkan executable sementara.
- [ ] EXE dipindah: tampilkan peringatan dan aktifkan ulang untuk memperbarui lokasi.
- [ ] Tidak ada startup yang diaktifkan hanya karena aplikasi dibuka atau profil koneksi disimpan.

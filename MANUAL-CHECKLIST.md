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
- [ ] Signal, Cell, History, dan Settings tetap Coming soon dan nonaktif.
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
- [ ] Ping selalu — dengan keterangan tidak tersedia; durasi GET tidak dipakai.
- [ ] RSSI ditampilkan di Overview; seluruh metrik widget tetap tersedia.
- [ ] Metrik khusus LTE kosong saat jaringan bukan LTE/LTE-A.

## Polling dan error

- [ ] Poll pertama langsung; selanjutnya tiap 2 detik bila siklus sebelumnya selesai.
- [ ] Dua endpoint diminta paralel; siklus lambat tidak menumpuk request.
- [ ] Request yang macet dibatasi deadline siklus sekitar 5 detik.
- [ ] UI, pemindahan window, dan scroll tetap responsif selama request berlangsung.
- [ ] Snapshot dua window konsisten setelah masing-masing repaint.
- [ ] Satu endpoint gagal: tidak ada campuran nilai baru dengan sebagian nilai lama.
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
- [ ] Tidak ada request perubahan konfigurasi modem, database, tray, autostart, atau speed test.
- [ ] Membuka ulang aplikasi memulai history baru tanpa data tersimpan.

# Checklist manual — belum dijalankan

Checklist ini disiapkan untuk sesi menjalankan aplikasi oleh pengguna.
Tidak ada unit test, build, atau pengujian runtime selama implementasi.

## Window dan DPI

- [ ] Startup menampilkan satu widget dengan label Demo, awalnya Loading.
- [ ] Pada scaling Windows 100%, area konten sekitar 300 × 380; title bar di luar area tersebut.
- [ ] Pada scaling 125%, 150%, dan 200%, teks tetap terbaca dan tidak bertumpuk.
- [ ] Widget tetap di atas window aplikasi biasa saat aplikasi lain mendapat fokus.
- [ ] Widget dapat dipindah lewat title bar; klik area konten membuka Overview.
- [ ] Dashboard berukuran awal sekitar 900 × 650 dp dan dapat diperbesar.
- [ ] Klik widget berulang kali tidak membuat dashboard duplikat.
- [ ] Klik cepat saat dashboard sedang dibuat tidak membuat window tambahan.
- [ ] Dashboard yang diminimalkan dipulihkan lewat klik widget.
- [ ] Menutup dashboard menyisakan widget; klik berikutnya membuka dashboard baru.
- [ ] Menutup widget saat dashboard terbuka menutup keduanya dan mengakhiri proses.
- [ ] Menutup widget saat dashboard diminimalkan atau sedang dibuat tetap mengakhiri proses.
- [ ] Fokus keyboard/Enter/Space dapat membuka dashboard dan mengganti skenario.
- [ ] Sidebar menu selain Overview tidak dapat diklik dan bertanda Coming soon.

## Data dan chart

- [ ] Setelah startup 2 detik, fixture pertama: score 82/GOOD, RSRP -84 dBm,
      RSRQ -9 dB, SINR 21 dB, B3, PCI 153, 12.4/3.1 Mbps, ping 28 ms.
- [ ] Widget dan Overview menampilkan nilai yang sama setelah kedua window repaint.
- [ ] Urutan mock berulang; badge berubah antara LTE dan LTE-A.
- [ ] Hanya Overview menambahkan RSSI, tanpa menghilangkan metrik widget lainnya.
- [ ] Grafik RSRP, RSRQ, dan SINR terpisah, memiliki unit dan timestamp.
- [ ] Sampel pertama berupa titik; garis muncul setelah ada sampel berikutnya.
- [ ] Setelah lebih dari 1 menit, history tetap maksimal 30 sampel.
- [ ] Mengubah ukuran dashboard atau melakukan scroll tidak mengubah nilai mock.
- [ ] Label Demo tetap terlihat di sidebar saat Overview digulir ke bawah.
- [ ] Pada lebar dashboard minimum 760 dp, pilihan skenario membungkus ke dua baris
      dan grafik tersusun vertikal; konten bisa digulir tanpa menumpuk.

## State

- [ ] Pilih Loading: semua pengukuran dan grafik kosong, tidak berubah Online otomatis.
- [ ] Pilih No Signal: angka tampil `—`, bukan nol; status No Signal terlihat di kedua window.
- [ ] Online → API Error: nilai tetap sama, score/grafik ditandai stale, timestamp sukses tidak berubah.
- [ ] Online → Disconnected: kondisi stale sama; usia data terus bertambah.
- [ ] Error/Disconnected → Online: pengukuran diperbarui dan history dimulai dari segmen baru.
- [ ] Online → Loading/No Signal → Online juga memulai segmen grafik baru,
      langsung memakai fixture berikutnya tanpa mengulang urutan dari awal.
- [ ] Memilih ulang Online tidak mereset history atau menambah sampel di luar ticker.
- [ ] Pada startup, buka dashboard dan pilih Loading sebelum 2 detik; lalu pilih API Error
      dan Disconnected. Tanpa sampel sukses, tetap tampil `—` dan Waiting for data.
- [ ] Online → Loading/No Signal → API Error: cache sukses terakhir muncul sebagai stale.
- [ ] Demo API Error hanya tampilan simulasi; tidak menghubungi endpoint apa pun.
- [ ] Selama state non-Online, jumlah sampel tidak bertambah; stale mempertahankan
      grafik terakhir, sementara Loading/No Signal menyembunyikannya.

## Batasan MVP

- [ ] Tidak ada speed test, request jaringan aplikasi, database, system tray,
      autostart, atau pengaturan yang tersimpan.
- [ ] Saat aplikasi dibuka ulang, mock/history dimulai dari awal.

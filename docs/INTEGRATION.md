# Integrasi Dokumentasi ke Backend

Dokumentasi OpenAPI sudah terintegrasi ke dalam backend dan dapat diakses langsung melalui server.

## Endpoint Dokumentasi

Setelah server berjalan, dokumentasi dapat diakses di:

### 1. Swagger UI

```
http://localhost:8080/docs
```

Atau akses root URL yang akan redirect ke `/docs`:

```
http://localhost:8080/
```

### 2. OpenAPI Specification (YAML)

```
http://localhost:8080/openapi.yaml
```

## Cara Menggunakan

1. **Jalankan server:**

   ```bash
   go run main.go
   ```

2. **Buka browser dan akses:**

   ```
   http://localhost:8080/docs
   ```

3. **Test API langsung dari Swagger UI:**
   - Klik endpoint yang ingin di-test
   - Klik tombol "Try it out"
   - Isi parameter jika diperlukan
   - Klik "Execute"
   - Lihat response

## Fitur yang Tersedia

✅ **Swagger UI Interaktif**

- Preview semua endpoint
- Test API langsung dari browser
- Lihat request/response examples
- Validasi input

✅ **OpenAPI Specification**

- Download file YAML
- Import ke Postman/Insomnia
- Generate client code
- Validasi dengan tools lain

## Troubleshooting

### File tidak ditemukan

Pastikan file `swagger-ui.html` dan `openapi.yaml` ada di root directory project.

### Swagger UI tidak load

- Pastikan server sudah berjalan
- Cek console browser untuk error
- Pastikan URL `/openapi.yaml` dapat diakses

### CORS Error

CORS sudah dikonfigurasi di backend. Jika masih ada masalah, pastikan:

- Server berjalan di port yang benar
- URL di browser sesuai dengan server URL

## Update Dokumentasi

Saat menambah endpoint baru:

1. Update `openapi.yaml` dengan endpoint baru
2. Restart server
3. Refresh browser di `/docs`
4. Dokumentasi akan otomatis terupdate

## Integrasi dengan Tools Lain

### Postman

1. Import dari URL: `http://localhost:8080/openapi.yaml`
2. Atau download file dan import

### Insomnia

1. Import → From URL
2. Masukkan: `http://localhost:8080/openapi.yaml`

### Redoc

1. Buka https://redocly.github.io/redoc/
2. Masukkan URL: `http://localhost:8080/openapi.yaml`


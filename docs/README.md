# API Documentation

Dokumentasi API untuk Smart Aquarium Controller menggunakan OpenAPI 3.1.0.

## File Dokumentasi

- **`openapi.yaml`** - Spesifikasi OpenAPI 3.1.0 lengkap
- **`swagger-ui.html`** - Swagger UI untuk preview dokumentasi

## Cara Menggunakan

### 1. Menggunakan Swagger UI (Lokal)

1. Buka file `swagger-ui.html` di browser
2. File akan otomatis memuat `openapi.yaml`
3. Anda bisa test API langsung dari browser

### 2. Menggunakan Online Tools

#### Swagger Editor (Online)

1. Buka https://editor.swagger.io/
2. Copy-paste isi `openapi.yaml`
3. Preview dan edit dokumentasi

#### Redoc (Online)

1. Buka https://redocly.github.io/redoc/
2. Upload atau paste URL ke `openapi.yaml`
3. Dapatkan dokumentasi yang lebih rapi

### 3. Generate Client/Server Code

Menggunakan Swagger Codegen:

```bash
# Install swagger-codegen
npm install -g @openapitools/openapi-generator-cli

# Generate Go client
openapi-generator-cli generate \
  -i openapi.yaml \
  -g go \
  -o ./generated/go-client

# Generate JavaScript client
openapi-generator-cli generate \
  -i openapi.yaml \
  -g javascript \
  -o ./generated/js-client
```

### 4. Integrasi dengan Backend

Untuk menampilkan dokumentasi di endpoint backend, tambahkan handler:

```go
// Di routes/routes.go
api.StaticFile("/docs", "./swagger-ui.html")
api.StaticFile("/openapi.yaml", "./openapi.yaml")
```

Lalu akses di: `http://localhost:8080/docs`

## Tools Populer untuk OpenAPI

| Tool           | URL                                  | Keterangan                   |
| -------------- | ------------------------------------ | ---------------------------- |
| **Swagger UI** | https://swagger.io/tools/swagger-ui/ | UI interaktif untuk test API |
| **Redoc**      | https://redocly.com/redoc            | Dokumentasi yang lebih rapi  |
| **Stoplight**  | https://stoplight.io/                | Design API + dokumentasi     |
| **Postman**    | https://www.postman.com/             | Import OpenAPI untuk testing |
| **Insomnia**   | https://insomnia.rest/               | Import OpenAPI untuk testing |

## Validasi OpenAPI

Validasi file OpenAPI:

```bash
# Install validator
npm install -g @apidevtools/swagger-cli

# Validate
swagger-cli validate openapi.yaml

# Bundle (gabungkan semua $ref)
swagger-cli bundle openapi.yaml -o openapi-bundled.yaml
```

## Update Dokumentasi

Saat menambah endpoint baru:

1. Update `openapi.yaml` dengan endpoint baru
2. Tambahkan schema di `components/schemas` jika perlu
3. Validasi dengan `swagger-cli validate`
4. Test di Swagger UI

## Contoh Request/Response

Semua contoh request dan response sudah ada di `openapi.yaml`. Untuk melihat contoh lengkap, buka Swagger UI dan klik "Try it out" pada endpoint yang diinginkan.


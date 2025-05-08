# Go Base Project

Este es un proyecto base en Go para desarrollar APIs REST robustas y organizadas, utilizando:

* 🔧 Gin como router HTTP
* 📄 GORM como ORM para PostgreSQL
* 📁 Migraciones y seeders estructurados
* 🦢 Separación de lógica interna en carpetas `config/`
* 🛠️ CLI personalizada y extensible

---

## 📦 Requisitos

* Go 1.22 o superior
* PostgreSQL
* [Make](https://www.gnu.org/software/make/) (opcional pero recomendado)
* WSL2 (si usas Windows)

---

## ⚙️ Configuración inicial

1. **Clona el repositorio**:

```bash
git clone https://github.com/ronygcgarcia/go_base_project.git
cd go_base_project
```

2. **Crea tu archivo `.env`**:

```env
# Entorno
APP_ENV=development

# Base de datos
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=go_project
DB_SSLMODE=disable

# Servidor
SERVER_PORT=8080
SERVER_HOST=localhost
```

3. **Instala las dependencias**:

```bash
go mod tidy
```

---

## 📁 Estructura del proyecto

```
.
├── main.go
├── routes/
├── config/
├── controllers/
├── models/
├── migrations/
│   ├── config/              # Lógica y control de migraciones
│   └── YYYYMMDD_name.go     # Archivos de migración individuales
├── seeders/
│   ├── config/              # Lógica y control de seeders
│   └── YYYYMMDD_name.go     # Archivos de seeder individuales
├── .env
├── Makefile
```

---

## 📜 Comandos disponibles

Puedes usar `make` para ejecutar los comandos disponibles fácilmente:

| Comando                        | Descripción                                |
| ------------------------------ | ------------------------------------------ |
| `make run`                     | Inicia el servidor API                     |
| `make migrate`                 | Ejecuta todas las migraciones pendientes   |
| `make rollback`                | Revierte todas las migraciones             |
| `make rollback-step`           | Revierte solo la última migración aplicada |
| `make migration name=...` | Genera un nuevo archivo de migración       |
| `make seed`                    | Ejecuta todos los seeders pendientes       |
| `make seed-rollback`           | Revierte todos los seeders aplicados       |
| `make seed-rollback-step`      | Revierte el último seeder aplicado         |
| `make seeder name=...`    | Genera un nuevo archivo de seeder          |

### 🧪 Ejemplos:

```bash
make make-migration name=create_users_table
make migrate

make make-seeder name=init_roles
make seed
```

---

## 🔒 Producción

En modo producción (`APP_ENV=production`), el servidor se ejecuta en HTTPS sobre el puerto 443 usando certificados ubicados en:

```
./certs/server.crt
./certs/server.key
```

Asegúrate de tener los certificados correctamente configurados antes de hacer deploy.

---

## ✅ Buenas prácticas

* No modifiques los archivos dentro de `migrations/config/` o `seeders/config/`
* Usa `Register()` en cada archivo de migración o seeder para registrarlo automáticamente
* Nunca uses `AutoMigrate()` directamente en producción
* Cada migración o seeder debe tener su función `Up()` y `Down()`

---

## 📄 Licencia

MIT License.

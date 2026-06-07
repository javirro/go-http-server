# Notas sobre Go

Apuntes acumulados sobre cómo funciona Go, basados en preguntas sobre este proyecto.

---

## Cuándo usar punteros en Go

En Go, un valor `T` se pasa por copia y un puntero `*T` permite trabajar sobre la misma instancia en memoria.

Usa punteros cuando:

- **Necesitas modificar el original** dentro de una función o método.
- **El struct es grande o tiene estado interno** (por ejemplo, tipos de infraestructura como `http.Server`).
- **Quieres compartir una sola instancia** entre varios componentes (handlers, router, shutdown, etc.).
- **Necesitas representar optionalidad** (`nil` vs "valor cero") en DTOs de entrada, especialmente en `PATCH`.
- **Tus métodos mutan estado**, por lo que el receptor debe ser `*T`.

Ejemplo:

```go
func setDefaultPort(cfg *Config) {
    if cfg.Port == 0 {
        cfg.Port = 8080
    }
}
```

Si recibiera `Config` (sin puntero), modificaría solo una copia local.

### Caso típico: `http.Server`

Se crea como puntero:

```go
httpServer := &http.Server{
    Addr:    cfg.Addr(),
    Handler: handler,
}
```

Puntos clave:

- Se evita copiar un tipo con estado/concurrencia.
- `ListenAndServe` y `Shutdown` operan sobre la misma instancia viva.
- Es consistente devolver también `*Server` si el wrapper tiene métodos con receptor puntero (`func (s *Server) Start(...)`).

### Regla práctica rápida

- **Struct pequeño e inmutable**: normalmente valor (`T`).
- **Struct mutable, compartido o concurrente**: puntero (`*T`).

Go se encarga de la memoria automáticamente: si devuelves un puntero a una variable local, el compilador la mueve al heap cuando hace falta (escape analysis).

---

## Struct tags para JSON

En Go, los **struct tags** son metadatos que se escriben junto a cada campo para indicar cómo debe tratarse al serializar/deserializar.  
Con `encoding/json`, el tag `json:"..."` define el nombre del campo en JSON y opciones adicionales.

Ejemplo:

```go
type Product struct {
    ID          int      `json:"id"`
    Name        string   `json:"name"`
    PriceCents  int      `json:"price_cents"`
    Tags        []string `json:"tags,omitempty"`
    InternalSKU string   `json:"-"`
}
```

Qué hace cada tag:

- `json:"id"`: en JSON, el campo se llama `id` (no `ID`).
- `json:"price_cents"`: transforma el nombre a snake_case.
- `json:"tags,omitempty"`: omite `tags` si está vacío (`nil`, `[]`, `""`, `0`, `false`, etc.).
- `json:"-"`: nunca incluye ese campo en JSON (ni al codificar ni al decodificar).

### Ejemplo de uso con `encoding/json`

```go
p := Product{
    ID:          10,
    Name:        "Keyboard",
    PriceCents:  4999,
    Tags:        nil,          // se omitirá por omitempty
    InternalSKU: "INT-ABC-01", // se omitirá por "-"
}

b, _ := json.Marshal(p)
fmt.Println(string(b))
// {"id":10,"name":"Keyboard","price_cents":4999}
```

Notas importantes:

- Solo se serializan campos **exportados** (con mayúscula inicial).
- Si no pones tag, `encoding/json` usa el nombre del campo del struct.
- Los tags son clave para mantener una API estable aunque cambies nombres internos en Go.

## Paquetes y visibilidad entre archivos

El **paquete** es la unidad de visibilidad en Go, no el archivo. Todos los archivos `.go` que declaran el mismo `package` en su cabecera comparten el mismo espacio de nombres. El compilador los fusiona antes de compilar, como si fueran un único archivo.

```go
// store.go
package handler
var shop = newShopStore()   // declarado aquí
```

```go
// products.go
package handler
shop.mu.RLock()   // accesible directamente, mismo paquete
```

La división en múltiples archivos es puramente organizativa.

### Importar solo es necesario para otros paquetes

Dentro del mismo paquete no se importa nada. Los `import` solo aparecen para usar código de paquetes externos o de la librería estándar:

```go
import (
    "net/http"    // paquete estándar
    "sync"        // paquete estándar
)
```

Entre archivos del mismo paquete (`handler`, en este caso), todo está disponible de forma automática, sin ningún import.

### Regla de exportación: mayúsculas / minúsculas

| Nombre | Visibilidad |
|---|---|
| `shop`, `shopStore`, `mu` | Solo dentro del paquete `handler` |
| `Product`, `ListProducts`, `NewRouter` | Exportado — visible desde cualquier paquete |

---

## Concurrencia: `sync.RWMutex`

Un servidor HTTP atiende múltiples requests en paralelo (una goroutine por request). Cuando varias goroutines leen o escriben la misma variable al mismo tiempo sin coordinación se produce una **condición de carrera** y el comportamiento es indefinido.

`sync.RWMutex` es un mutex de lectura/escritura que resuelve esto con dos modos:

| Método | Modo | Comportamiento |
|---|---|---|
| `RLock()` / `RUnlock()` | Lectura compartida | Múltiples goroutines pueden leer a la vez |
| `Lock()` / `Unlock()` | Escritura exclusiva | Solo una goroutine puede escribir; bloquea toda lectura |

En este proyecto, `shop.mu` protege el acceso al store en memoria:

```go
// Lectura — varios requests pueden listar productos simultáneamente
shop.mu.RLock()
defer shop.mu.RUnlock()
// ... leer shop.products ...

// Escritura — solo un request puede crear/borrar a la vez
shop.mu.Lock()
defer shop.mu.Unlock()
// ... modificar shop.products ...
```

---

## `defer`

`defer` pospone la ejecución de una llamada a función hasta que la función que la contiene **retorne**, sin importar por qué retorna (return normal, return anticipado, panic).

Es el patrón idiomático para liberar recursos:

```go
shop.mu.RLock()
defer shop.mu.RUnlock()   // se ejecuta siempre al salir, aunque haya varios return
```

Sin `defer` habría que llamar a `RUnlock()` antes de cada `return` de la función, lo cual es propenso a errores (si se olvida uno, el mutex queda bloqueado para siempre → deadlock).

---


## Handler HTTP

```go
func ListProducts(w http.ResponseWriter, r *http.Request) {---}
```

- r *http.Request se pasa como puntero para evitar copias innecesarias y trabajar con la misma instancia del request durante el ciclo de vida de la petición.
- w http.ResponseWriter es una interfaz: el valor concreto que implementa esa interfaz ya se pasa dentro de ella (normalmente por referencia), por eso se usa la interfaz directamente y no *http.ResponseWriter.

---

## Un directorio = un paquete (causa de los errores al reorganizar carpetas)

Al mover los archivos a una estructura nueva (`internal/platform/...` e `internal/products/`) el proyecto dejó de compilar. La causa de fondo es una de las reglas más importantes de Go:

> **En Go, la unidad de compilación es el _directorio_, no el archivo. Cada directorio es exactamente un paquete, y la identidad de ese paquete es su _ruta de import_ (la carpeta), no el nombre que pongas en la línea `package`.**

De aquí salieron tres tipos de error distintos.

### 1. Mismo `package handler` en carpetas distintas ≠ mismo paquete

Varios archivos seguían declarando `package handler` aunque estaban repartidos en carpetas diferentes:

```
internal/handler/            → package handler
internal/products/           → package handler   (¡otra carpeta!)
internal/platform/server/respond/ → package handler   (¡otra más!)
internal/platform/server/routes/  → package handler
```

Aunque el nombre coincidía, para Go son **paquetes diferentes** porque están en directorios diferentes. Por eso el código de `internal/products` no veía `Product`, `shopStore`, `JSON`, etc. (vivían en `internal/handler`) y aparecían errores como:

```
undefined: ProductStatus
undefined: Product
undefined: JSON
undefined: shopStore
```

Dentro de un mismo paquete (misma carpeta) todo es visible sin importar nada. Pero **entre paquetes** hay que:

1. Importar la ruta del otro paquete.
2. Que el símbolo esté **exportado** (empiece por mayúscula).
3. Cualificarlo con el nombre del paquete: `respond.JSON(...)`, `products.Register(...)`.

### 2. El nombre del paquete debería coincidir con el de la carpeta

Tener `package handler` dentro de la carpeta `respond/` o `routes/` es legal, pero confunde: te obliga a importar `.../routes` y luego escribir `handler.NewRouter()`. La convención de Go es que **el nombre del paquete coincida con el nombre de la carpeta**. Por eso se renombró:

| Carpeta | Antes | Ahora |
|---|---|---|
| `internal/platform/server/respond` | `package handler` | `package respond` |
| `internal/platform/server/routes`  | `package handler` | `package routes` |
| `internal/products`                | `package handler` | `package products` |

### 3. Imports apuntando a rutas que ya no existían

Al mover carpetas, los `import` quedaron obsoletos. `main.go` y el test de middleware seguían importando las rutas viejas:

```go
// antes (rutas que ya no existen)
"github.com/javier/go-http-server/internal/config"
"github.com/javier/go-http-server/internal/handler"
"github.com/javier/go-http-server/internal/middleware"
"github.com/javier/go-http-server/internal/server"

// ahora (rutas reales tras la reorganización)
"github.com/javier/go-http-server/internal/platform/config"
"github.com/javier/go-http-server/internal/platform/server/middleware"
"github.com/javier/go-http-server/internal/platform/server/routes"
"github.com/javier/go-http-server/internal/platform/server/httpserver"
```

La ruta de import es **relativa al `module`** declarado en `go.mod` (`github.com/javier/go-http-server`) + la ruta de la carpeta. Si mueves la carpeta, cambias el import.

### Regla práctica al reorganizar carpetas en Go

1. Una carpeta = un paquete; muévelo entero, no a medias.
2. El `package` de todos los archivos de esa carpeta debe ser igual y, por convención, igual al nombre de la carpeta.
3. Tras mover, actualiza **todos** los `import` que apuntaban ahí.
4. Lo que use otro paquete debe estar **exportado** (mayúscula) y cualificado (`paquete.Símbolo`).
5. Verifica con `go build ./...` y `go vet ./...` (compilan todos los paquetes del módulo).

---

## Conexión a base de datos: por qué `pgx` (`jackc/pgx`)

Para PostgreSQL en Go, la librería más usada en producción hoy es **`pgx`** (`github.com/jackc/pgx/v5`), normalmente a través de su pool de conexiones **`pgxpool`**. Es la opción recomendada por la comunidad y la base sobre la que se apoyan muchos ORMs/generadores (por ejemplo `sqlc`).

### El panorama de opciones

| Opción | Qué es | Estado / uso |
|---|---|---|
| `database/sql` (stdlib) | API genérica de SQL; **necesita un driver** aparte | Estándar, pero es solo la abstracción, no habla con Postgres por sí sola |
| `lib/pq` | Driver clásico de Postgres para `database/sql` | En **modo mantenimiento**; el propio repo recomienda migrar a `pgx` |
| **`pgx`** | Driver + toolkit nativo de Postgres | **Recomendado en producción**; activo, rápido, soporta tipos de Postgres |
| ORMs (GORM, etc.) | Capa de abstracción sobre el driver | Útiles, pero por debajo suelen usar un driver como `pgx` |

### Por qué `pgx` y no `database/sql` + `lib/pq`

- **Rendimiento**: usa el protocolo binario de Postgres y hace menos copias/parseos que `lib/pq`.
- **Tipos nativos**: soporta de forma natural tipos propios de Postgres (`jsonb`, arrays, `uuid`, `numeric`, etc.).
- **Mantenimiento activo**: `lib/pq` está congelado; `pgx` recibe mejoras y fixes.
- **Pool de calidad**: `pgxpool` es un pool pensado para concurrencia (cada request HTTP es una goroutine), con control de conexiones máximas, healthchecks y reciclado.

> Nota: `pgx` puede usarse de dos formas: con su **API nativa** (`pgx`/`pgxpool`, lo que usamos aquí) o como **driver de `database/sql`** (`stdlib`). La API nativa da acceso a todas las features de Postgres; la vía `database/sql` aporta portabilidad si algún día cambias de motor.

### Cómo está montado en este proyecto

La conexión vive en `internal/platform/database/postgres.go` y expone un único constructor:

```go
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error)
```

Pasos clave:

1. `pgxpool.ParseConfig(cfg.DatabaseURL)` parsea el **DSN** (`postgres://user:pass@host:port/db?sslmode=...`).
2. Se ajusta `MaxConns` (tamaño máximo del pool) desde configuración.
3. `pgxpool.NewWithConfig(...)` crea el **pool** (no una sola conexión).
4. `pool.Ping(...)` con timeout verifica que la DB responde **antes** de aceptar tráfico (fail-fast en el arranque).

### El pool y la concurrencia

`*pgxpool.Pool` es **seguro para uso concurrente**: se crea **una vez** en `main` y se comparte entre todas las goroutines (una por request). No se abre una conexión por request; el pool las reparte y reutiliza. Por eso:

- Se construye en el arranque y se inyecta hacia abajo (repositorios).
- Se cierra una sola vez al apagar: `defer pool.Close()`.

### Configuración (variables de entorno)

| Variable | Default | Significado |
|---|---|---|
| `DATABASE_URL` | `postgres://app:secret@localhost:5432/football_store?sslmode=disable` | DSN de conexión |
| `DB_MAX_CONNS` | `10` | Conexiones máximas en el pool |
| `DB_CONN_TIMEOUT` | `5s` | Timeout para conectar/hacer ping |

El default apunta al Postgres del `docker-compose.yml`. Para levantarlo en local:

```bash
make db-up      # docker compose up -d
make db-logs    # ver logs
make db-down    # parar (conserva el volumen de datos)
```

### Persistencia de productos en PostgreSQL

El repositorio de productos se persiste en **PostgreSQL**. La interfaz `ProductRepository` define el contrato y `PostgresProductRepository` es la única implementación; el resto de la app (servicio, controlador, rutas) depende de la interfaz, no de la implementación:

```go
type ProductRepository interface {
    List(ctx context.Context, filters ProductFilters) ([]Product, error)
    Count(ctx context.Context) (int, error)
    GetByID(ctx context.Context, id int64) (Product, error)
    Create(ctx context.Context, product Product) (Product, error)
    Update(ctx context.Context, product Product) (Product, error)
    Delete(ctx context.Context, id int64) error
}
```

- `PostgresProductRepository` → `NewPostgresRepository(pool)`, en `internal/products/product_repository_postgres.go`.

> Se mantiene la **interfaz** aunque haya una sola implementación: es la costura que permite inyectar dependencias (`main` construye el repo y lo pasa a `routes.NewRouter(repo)`) y sustituirlo en los tests sin tocar el resto del código.

Los tests de `products` son **de integración**: usan el repo Postgres real y se **saltan** (`t.Skip`) si no hay base de datos accesible, de modo que `go test ./...` sigue pasando sin DB. Con `make db-up` levantada, se ejecutan de verdad.

#### ctx + errores en la interfaz

Al pasar a una DB real, los métodos del repositorio ahora reciben `context.Context` (para timeouts/cancelación de la request, viaja desde `r.Context()`) y devuelven `error`. Un producto inexistente se reporta con el centinela `ErrProductNotFound`, que el controlador traduce a `404`.

#### Modelo relacional con JSONB

El producto tiene colecciones anidadas (`options`, `variants`, `images`). En lugar de normalizarlas en varias tablas con joins, se guarda un **híbrido**:

- Campos escalares y filtrables como columnas (`title`, `handle`, `vendor`, `product_type`, `status`, ...).
- Colecciones anidadas como columnas **`JSONB`**.

Esto encaja muy bien con `pgx`, que sabe serializar/deserializar `JSONB` directamente a/desde structs y slices de Go (vía `encoding/json`). Así el `scan` puede leer `&p.Options`, `&p.Variants`, `&p.Images` sin trabajo manual, y los filtros (`vendor`, `status`, etc.) siguen siendo columnas indexadas.

#### Migraciones

Las migraciones viven en `internal/platform/database/migrations/` como ficheros versionados (`0001_create_products.up.sql` / `.down.sql`) y se **embeben en el binario** con `//go:embed`. Un migrador minimalista (`database.Migrate`) crea una tabla `schema_migrations`, y aplica en orden solo las versiones que falten, cada una dentro de una transacción (idempotente). Se ejecuta al arrancar, antes de servir tráfico.

> Nota: aquí el migrador es casero y minimalista con fines didácticos. En proyectos grandes se suele usar una herramienta dedicada como **`golang-migrate`** o **`goose`**.

#### Seed

`PostgresProductRepository.Seed` inserta el catálogo inicial **solo si la tabla está vacía** (reutiliza `seedProducts()`, el mismo que usa el repo en memoria). Como inserta IDs explícitos, después avanza la secuencia `IDENTITY` con `setval(pg_get_serial_sequence(...))` para que los `INSERT` posteriores no colisionen.

#### Flujo de arranque en `main`

```
config.Load → logger → database.NewPool (Ping) → database.Migrate → productRepo.Seed → routes.NewRouter(repo) → server.Start
```

Para probarlo en local:

```bash
make db-up    # levanta Postgres
make run      # migra, siembra y arranca la API
```
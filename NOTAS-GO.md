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
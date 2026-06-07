# Concurrencia en Go vs tareas en background (Cloud Tasks)

Comparativa para developers que vienen de Node.js y están acostumbrados a
delegar tareas pesadas a Cloud Tasks / Pub/Sub para no bloquear el event loop.

---

## El modelo de Node.js: qué bloquea y qué no

Node.js tiene **un solo hilo** (event loop), pero eso no significa que cualquier
operación lenta lo bloquee. La distinción clave es si el código **cede el
control** o no.

### `await` no bloquea el event loop

```javascript
// Node.js
async function handler(req, res) {
    await sleep(5000)  // ← cede el control al event loop durante 5s
    res.json({ ok: true })
}
```

Durante esos 5 segundos el event loop está **libre** para atender otros
requests. `await` le dice a Node "mientras esperas esto, ve a hacer otra cosa".
Es concurrencia **cooperativa**: el código cede el control explícitamente.

### Lo que sí bloquea: cómputo síncrono pesado

```javascript
// Esto SÍ bloquea — no hay await, no hay I/O, solo CPU
function handler(req, res) {
    const result = parsearCSVDe100000Filas(data)  // ← bloquea 3 segundos
    res.json(result)
}
```

Mientras ese bucle corre, **ningún otro request puede ser atendido**. Aquí es
donde Node.js tiene el problema real: si el código no cede el control, el event
loop se congela.

```
Node.js (un hilo):
Request 1 → cómputo síncrono pesado (3s) → bloquea TODO → Request 2 espera...
Request 1 → await sleep(5s) → cede control → Request 2 se atiende → Request 1 continúa
```

Por eso la solución habitual para cómputo pesado era sacar la tarea fuera del
proceso: Cloud Tasks, Pub/Sub, una queue...

---

## Cómo funciona Go

Go tiene un **scheduler de goroutines** sobre múltiples hilos del OS
(`GOMAXPROCS` = nº de CPUs por defecto). Cada request HTTP llega en **su propia
goroutine**, y si una goroutine se bloquea (esperando DB, red, etc.), Go pausa
esa goroutine y pone otra a correr. Nunca bloqueas el servidor entero.

```
Go (múltiples hilos + scheduler):

Request 1 → goroutine 1 → espera DB → Go pausa goroutine 1
Request 2 → goroutine 2 → se atiende inmediatamente
goroutine 1 → DB responde → continúa
```

El servidor HTTP de Go atiende cada request en su propia goroutine
automáticamente. No tienes que hacer nada especial para que esto ocurra.

---

## La diferencia real entre Node.js y Go

**Node.js:** concurrencia **cooperativa**. El código tiene que ceder el control
explícitamente con `await`/Promises. Si no lo hace (cómputo síncrono), bloquea.

**Go:** concurrencia **apropiativa** (gestionada por el scheduler). Las
goroutines se pausan y reanudan automáticamente, incluso en mitad de código
síncrono, sin que el programador tenga que recordar nada:

```go
// Go — esto no bloquea el servidor aunque tarde 5s.
// El scheduler pausa esta goroutine y atiende otras.
func handler(w http.ResponseWriter, r *http.Request) {
    time.Sleep(5 * time.Second)  // goroutine en pausa, el resto sigue
    respond.JSON(w, r, 200, "ok")
}
```

| Operación | Node.js | Go |
|---|---|---|
| `await sleep(5s)` / esperar DB | No bloquea — event loop libre | No bloquea — goroutine en pausa |
| Bucle CPU de 3s sin yield | **Bloquea todo** | No bloquea — otras goroutines siguen |
| 1000 requests esperando DB | OK con async/await | OK con goroutines |
| 1000 requests con cómputo pesado | El primero bloquea a los demás | El scheduler los reparte entre goroutines |

---

## La distinción clave: I/O-bound vs CPU-bound

### Tareas I/O-bound (red, DB, archivos)

Llamadas a APIs externas, queries SQL, escribir ficheros... el programa
**espera** más que computa. Go las maneja perfecto: mientras una goroutine
espera, el scheduler corre otras. **No necesitas Cloud Tasks para esto.**

```go
// Esto en Go es perfectamente válido dentro de un handler.
// La goroutine de este request espera la DB; el resto del servidor sigue libre.
func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) {
    product, err := c.service.Create(r.Context(), input)
    // ...
}
```

### Tareas CPU-bound (cómputo intensivo)

Redimensionar imágenes, generar PDFs, procesar vídeo, parsear CSVs grandes...
estas tareas **consumen CPU de forma sostenida**. En Go también se pueden hacer
en goroutines, pero si son muy largas o costosas, tiene sentido sacarlas del
proceso — no por el modelo de concurrencia, sino por **fiabilidad** y
**timeouts** (Cloud Run mata requests que superen 60 min).

---

## ¿Cuándo sigue teniendo sentido Cloud Tasks con Go?

| Situación | ¿Cloud Tasks? | Por qué |
|---|---|---|
| Query a la DB lenta (1–2 s) | No | I/O-bound, goroutine espera, el resto sigue |
| Llamada a API externa lenta | No | Igual, es I/O-bound |
| Enviar un email tras crear un recurso | Depende | Si puede fallar y reintentar, Cloud Tasks da reintentos automáticos |
| Procesar un CSV de 50 000 filas | Sí | Larga, usa mucha CPU/memoria, el cliente no debería esperar |
| Generar un PDF complejo | Sí | Larga, el usuario no espera la respuesta HTTP |
| Tarea que puede tardar más de 60 min | Sí | Cloud Run tiene timeout máximo por request |
| Necesitas garantía de ejecución (at-least-once) | Sí | Cloud Tasks reintenta si el proceso muere a mitad |

---

## El patrón Go equivalente a "mandar a background"

Para cosas intermedias (no tan críticas como para Cloud Tasks), puedes lanzar
una goroutine y responder al cliente inmediatamente:

```go
func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) {
    product, err := c.service.Create(r.Context(), input)
    if err != nil { /* ... */ }

    // Responde al cliente ya — no espera al trabajo de background.
    respond.JSON(w, r, http.StatusCreated, product)

    // Lanza trabajo en background sin bloquear el response.
    go func() {
        // IMPORTANTE: usa context.Background(), no r.Context().
        // r.Context() se cancela cuando el handler termina.
        _ = sendConfirmationEmail(context.Background(), product)
    }()
}
```

### Problema de este patrón en Cloud Run

Si Cloud Run escala a cero o la instancia se reinicia **mientras la goroutine
está corriendo**, esa goroutine muere sin terminar. Para tareas que deben
ejecutarse sí o sí (emails, facturación, notificaciones críticas), Cloud Tasks
sigue siendo la opción correcta porque **persiste la tarea y reintenta**.

---

## Resumen de decisión

```
¿La tarea es I/O-bound y dura < timeout?  → goroutine Go, sin Cloud Tasks
¿La tarea es CPU-intensiva o muy larga?   → Cloud Tasks (o Pub/Sub)
¿Necesitas garantía de ejecución?         → Cloud Tasks siempre
¿El cliente puede esperar la respuesta?   → hazlo inline en el handler
¿El cliente no debe esperar?              → responde ya + goroutine o Cloud Tasks
```

---

## La diferencia filosófica respecto a Node.js

En Node.js, el motivo para usar Cloud Tasks solía ser **no bloquear el event
loop** con cómputo pesado síncrono. Pero hay un matiz importante: un `await`
en Node **no bloquea** — cede el control cooperativamente. Lo que sí bloqueaba
era el cómputo síncrono sin yields.

En Go ese problema directamente no existe: el scheduler gestiona las goroutines
de forma apropiativa, así que **el motivo para usar Cloud Tasks nunca es el
modelo de concurrencia**. El motivo es siempre la **fiabilidad** (reintentos,
persistencia de la tarea si la instancia muere) y los **timeouts** (tareas que
superan el límite del proceso o de Cloud Run).

> En Node.js: Cloud Tasks para no bloquear.
> En Go: Cloud Tasks para garantizar ejecución, no para desbloquear.

# JSON en Go: Marshal/Unmarshal vs Streams

Explica cuándo usar cada enfoque para serializar y deserializar JSON,
tanto en el servidor como en el cliente HTTP.

---

## Los dos enfoques

### Marshal / Unmarshal — trabaja con `[]byte` en memoria

```go
// Serializar struct → []byte
data, err := json.Marshal(input)

// Deserializar []byte → struct
err := json.Unmarshal(data, &result)
```

Materializa todo en un `[]byte` antes de operar. Simple y directo.

### Encoder / Decoder — trabaja con streams (io.Reader / io.Writer)

```go
// Serializar struct → stream de salida
json.NewEncoder(w).Encode(input)

// Deserializar stream de entrada → struct
json.NewDecoder(r).Decode(&result)
```

Lee o escribe en un flujo sin pasar por un `[]byte` intermedio.

---

## En el servidor

El servidor trabaja con streams HTTP por naturaleza:
- `r.Body` es un `io.Reader` — el body llega como flujo, no como bytes en memoria.
- `w` (`http.ResponseWriter`) es un `io.Writer` — la respuesta se escribe al vuelo.

Por eso en `respond.go` se usan encoder/decoder:

```go
// Recibir request body
func DecodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
    dec := json.NewDecoder(r.Body)   // lee desde el stream del request
    dec.DisallowUnknownFields()
    dec.Decode(v)
}

// Enviar respuesta
func JSON(w http.ResponseWriter, r *http.Request, status int, v any) {
    json.NewEncoder(w).Encode(v)     // escribe directamente al stream de respuesta
}
```

Usar `io.ReadAll` + `Unmarshal` en el servidor funcionaría, pero cargaría todo
el body en memoria antes de procesarlo — innecesario para la mayoría de casos.

---

## En el cliente

El cliente hace peticiones HTTP con `http.Client`. Aquí el enfoque cambia:

### Enviar datos (request body): casi siempre `Marshal`

```go
body, _ := json.Marshal(input)
req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
```

¿Por qué `Marshal` y no un encoder con pipe? Porque `http.NewRequest` necesita
un `io.Reader` con los bytes ya disponibles para calcular el `Content-Length`.
Los payloads de request son normalmente pequeños (pocos KB), así que el coste
de materializarlos en memoria es despreciable.

### Recibir respuesta: depende del tamaño

**Respuestas pequeñas** — `io.ReadAll` + `Unmarshal` (más común):

```go
defer resp.Body.Close()
data, _ := io.ReadAll(resp.Body)
json.Unmarshal(data, &result)
```

**Respuestas grandes** — `NewDecoder` sobre el stream (más eficiente):

```go
defer resp.Body.Close()
json.NewDecoder(resp.Body).Decode(&result)
```

Evita cargar todo en memoria. Útil con listas largas, exports o respuestas de
varios MB.

---

## Comparativa

| | Marshal / Unmarshal | Encoder / Decoder |
|---|---|---|
| Trabaja con | `[]byte` en memoria | Stream (`io.Reader` / `io.Writer`) |
| Uso de memoria | Todo el JSON en RAM a la vez | Procesa en fragmentos |
| Cuándo usarlo | Payloads pequeños, request bodies | Respuestas grandes, servidor HTTP |
| Complejidad | Menor — más explícito | Menor — directo sobre el stream |
| Rendimiento | Suficiente para < 1 MB | Mejor para > 1 MB o streaming |

---

## Patrón helper en proyectos reales

En producción, los proyectos suelen encapsular el cliente HTTP en un helper
genérico que combina ambos enfoques:

```go
// Marshal para enviar (necesita bytes), Decoder para recibir (stream)
func do[T any](ctx context.Context, method, url string, in any) (T, error) {
    var zero T

    body, err := json.Marshal(in)
    if err != nil {
        return zero, err
    }

    req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
    if err != nil {
        return zero, err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return zero, err
    }
    defer resp.Body.Close()

    var result T
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return zero, err
    }
    return result, nil
}
```

---

## Cuándo JSON deja de ser la respuesta correcta

En servicios internos de alta frecuencia (microservicios que se llaman miles de
veces por segundo), el coste de serialización JSON se nota. La alternativa
habitual es **gRPC + Protobuf**:

- Protobuf es binario: más compacto y más rápido de serializar que JSON.
- El codec lo genera automáticamente el compilador de Protobuf.
- No hay `Marshal`/`Unmarshal` manual: el cliente y servidor generados lo hacen todo.

| | JSON | gRPC / Protobuf |
|---|---|---|
| Legibilidad | Alta (texto plano) | Baja (binario) |
| Tamaño del mensaje | Mayor | 3–10x menor |
| Velocidad de serialización | Moderada | Alta |
| Tooling / debug | Fácil (curl, Postman) | Requiere herramientas específicas |
| Uso habitual | APIs públicas, REST | Comunicación interna entre servicios |

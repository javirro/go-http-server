# Deploy y Serverless (Cloud Run)

Este documento resume cómo aplicar despliegue serverless a esta API en Go y qué implicaciones tiene sobre memoria/estado.

---

## Qué significa "memoria local"

Cuando digo **memoria local**, me refiero a la memoria RAM del proceso que corre tu API en **una instancia concreta** (su heap de Go, maps, slices, caches en memoria, etc.).

Ejemplos en este proyecto:

- `shop` en memoria (`map`, contadores, etc.) es estado local de la instancia.
- Cualquier dato guardado en variables globales o en structs del proceso vive solo en esa instancia.

Importante:

- Si Cloud Run levanta varias instancias, cada una tiene su propia memoria local.
- Esa memoria **no se comparte** entre instancias.
- Si una instancia se reinicia o escala a cero, se pierde su estado local.

---

## Cómo funciona el heap en Cloud Run

Cloud Run es serverless, pero tu app Go corre dentro de un contenedor Linux normal:

- Sí existe heap de Go.
- El GC funciona igual que en cualquier binario Go.
- El heap tiene un límite práctico marcado por la memoria asignada a la instancia (512Mi, 1Gi, etc.).

Si excedes la memoria asignada, la instancia puede terminar por OOM.

---

## ¿Tiene sentido serverless para Go?

Sí, en muchos casos:

- APIs stateless.
- Tráfico variable.
- Menos operación de infraestructura.
- Escalado automático y pago por uso.

Go encaja bien porque suele tener buen rendimiento, footprint moderado y arranque rápido.

---

## Cuándo encaja muy bien

- API HTTP sin estado en memoria como fuente de verdad.
- Estado persistente en DB/cache externo (Cloud SQL, Firestore, Redis, etc.).
- Jobs cortos, webhooks o microservicios.

---

## Cuándo no encaja tan bien

- Dependes de estado compartido en RAM entre requests/instancias.
- Procesos largos y muy stateful.
- Carga totalmente estable 24/7 donde puede convenir costo fijo.

---

## Recomendación para este repo

La API se puede desplegar en Cloud Run sin problema técnico (escucha en host/puerto configurable y tiene cierre graceful).  
Para producción, lo recomendable es:

1. Mantener la API stateless.
2. Mover el store en memoria a persistencia externa.
3. Mantener `PORT` desde entorno y timeouts bien definidos.

---

## Checklist rápido de despliegue a Cloud Run

1. Construir imagen Docker.
2. Subir imagen a Artifact Registry.
3. Desplegar con `gcloud run deploy`.
4. Configurar variables de entorno (`ENV`, CORS, rate limits, etc.).
5. Definir memoria, CPU, concurrencia y min/max instances según carga.


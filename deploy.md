# Deploy y Serverless (Cloud Run)

Este documento resume cómo desplegar esta API en Go (Cloud Run) y qué
implicaciones tiene sobre estado, base de datos y configuración.

---

## Estado actual del proyecto

La API ya **no usa estado en memoria** como fuente de verdad. Los productos se
persisten en **PostgreSQL** (vía `pgxpool`). En el arranque, el proceso:

1. Conecta al pool y hace ping (`database.NewPool`).
2. Aplica migraciones pendientes (`database.Migrate` — embebidas en el binario).
3. Siembra datos iniciales si la tabla está vacía (`productRepo.Seed`).

Esto significa que la API es **stateless** a nivel de instancia: cualquier
instancia de Cloud Run puede atender cualquier request sin necesitar estado
local compartido.

---

## Cómo funciona el heap en Cloud Run

Cloud Run es serverless, pero tu app Go corre dentro de un contenedor Linux normal:

- Sí existe heap de Go.
- El GC funciona igual que en cualquier binario Go.
- El heap tiene un límite práctico marcado por la memoria asignada a la instancia (512Mi, 1Gi, etc.).

Con la API actual (sin estado en memoria), el consumo de heap es muy bajo y
estable. El único estado en RAM es el pool de conexiones (`pgxpool`), que se
gestiona automáticamente y tiene un máximo configurable (`DB_MAX_CONNS`).

Si excedes la memoria asignada, la instancia puede terminar por OOM.

---

## ¿Tiene sentido serverless para Go?

Sí, y ahora más que antes:

- La API es **stateless** — las instancias no comparten nada en RAM.
- Estado persistente en PostgreSQL (Cloud SQL o cualquier Postgres gestionado).
- Go tiene arranque rápido y footprint moderado: el binario compilado con
  `scratch` pesa poco y arranca en milisegundos.
- Escalado automático y pago por uso.

---

## Cuándo encaja muy bien

- API HTTP stateless con persistencia externa (como ahora).
- Tráfico variable o impredecible.
- Jobs cortos, webhooks o microservicios.

---

## Cuándo no encaja tan bien

- Dependes de estado compartido en RAM entre requests/instancias.
- Procesos largos y muy stateful.
- Carga totalmente estable 24/7 donde puede convenir costo fijo.
- Latencia de cold-start crítica (aunque con Go suele ser de decenas de ms).

---

## Variables de entorno necesarias en producción

Además de las variables de servidor ya existentes, hay que configurar la base
de datos:

| Variable | Ejemplo | Descripción |
|---|---|---|
| `DATABASE_URL` | `postgres://user:pass@host:5432/db?sslmode=require` | DSN de conexión a PostgreSQL |
| `DB_MAX_CONNS` | `10` | Tamaño máximo del pool de conexiones |
| `DB_CONN_TIMEOUT` | `5s` | Timeout para conectar/ping al arrancar |
| `PORT` | `8080` | Puerto de escucha (Cloud Run lo inyecta automáticamente) |
| `ENV` | `production` | Entorno (`development` / `staging` / `production`) |
| `LOG_FORMAT` | `json` | Formato de logs (`json` recomendado en producción) |

En Cloud Run, `DATABASE_URL` debería apuntar a **Cloud SQL** (via socket Unix
con el Cloud SQL Auth Proxy, o TCP con IP privada). Nunca pongas credenciales
en el código; usa **Secret Manager** para inyectarlas como variables de entorno.

> `sslmode=require` es obligatorio para conexiones a Cloud SQL desde Cloud Run.

---

## Consideraciones con el pool de conexiones

Cloud Run puede escalar a **varias instancias** simultáneas. Cada instancia
abre su propio pool. Si tienes `DB_MAX_CONNS=10` y Cloud Run sube a 5
instancias, son hasta **50 conexiones** simultáneas contra la base de datos.

Ajusta `DB_MAX_CONNS` y la configuración de conexiones máximas en Cloud SQL
según la carga esperada. Con Cloud Run, valores bajos (5–10 por instancia)
suelen ser suficientes.

---

## Checklist de despliegue a Cloud Run

1. Aprovisionar **Cloud SQL** (PostgreSQL 16) y crear la base de datos y el usuario.
2. Configurar **Cloud SQL Auth Proxy** o acceso por IP privada desde Cloud Run.
3. Guardar `DATABASE_URL` y otras credenciales en **Secret Manager**.
4. Construir la imagen Docker:
   ```bash
   docker build -t REGION-docker.pkg.dev/PROJECT/REPO/server:TAG .
   ```
5. Subir la imagen a **Artifact Registry**:
   ```bash
   docker push REGION-docker.pkg.dev/PROJECT/REPO/server:TAG
   ```
6. Desplegar con `gcloud run deploy`:
   ```bash
   gcloud run deploy football-store \
     --image REGION-docker.pkg.dev/PROJECT/REPO/server:TAG \
     --region europe-west1 \
     --set-env-vars ENV=production,LOG_FORMAT=json,DB_MAX_CONNS=5 \
     --set-secrets DATABASE_URL=database-url:latest \
     --add-cloudsql-instances PROJECT:REGION:INSTANCE \
     --allow-unauthenticated
   ```
7. Verificar que las **migraciones** se aplicaron en el primer arranque (revisar logs).
8. Definir memoria, CPU, concurrencia y min/max instances según carga.

> Las migraciones se aplican automáticamente al arrancar la primera instancia
> gracias al migrador embebido en el binario. Si hay varias instancias
> arrancando a la vez, la tabla `schema_migrations` con su `PRIMARY KEY` evita
> que se aplique una misma migración dos veces (el segundo `INSERT` fallará y
> hará rollback, dejando la migración ya aplicada intacta).

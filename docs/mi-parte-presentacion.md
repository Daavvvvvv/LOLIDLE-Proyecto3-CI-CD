# Mi parte — Presentación Lolidle

Tres secciones de la rúbrica:

1. **Diagrama de arquitectura de software**
2. **Evidencia de las 3 modificaciones / introducciones**
3. **Demostración de App funcional en cada ambiente**

---

## URLs en vivo

| Ambiente | Frontend | Backend (ALB) |
|---|---|---|
| **dev** | http://lolidle-dev-frontend-560989793068.s3-website-us-east-1.amazonaws.com | http://lolidle-dev-alb-925096534.us-east-1.elb.amazonaws.com |
| **staging** | http://lolidle-staging-frontend-560989793068.s3-website-us-east-1.amazonaws.com | http://lolidle-staging-alb-1043385350.us-east-1.elb.amazonaws.com |
| **prod** | (no aplicado — ver opción A) | — |

**Dashboards CloudWatch:**
- Dev: https://us-east-1.console.aws.amazon.com/cloudwatch/home?region=us-east-1#dashboards/dashboard/lolidle-dev
- Staging: https://us-east-1.console.aws.amazon.com/cloudwatch/home?region=us-east-1#dashboards/dashboard/lolidle-staging

---

# SECCIÓN 1 — Diagrama de arquitectura de software (3-4 min)

## Diagrama (proyectarlo)

```
   Browser
      |  HTTP
      v
   S3 (frontend estático)  ----  DDragon (Riot, imágenes)
      |  XHR /api/*
      v
   ALB  --- Listener default -> Active TG (blue o green)
      |
      v
   ECS Fargate Cluster
      |-- Service blue  (2 tasks)
      `-- Service green (0 o 2 tasks)
      |
      v
   DynamoDB (sessions + lore-cache)
   Secrets Manager (Gemini key)
   Gemini API (externa)
   CloudWatch (logs + alarmas + dashboard)
   ECR (imágenes Docker, alimentado por CI)
```

## Guión literal

> "Nuestra arquitectura tiene 3 capas: presentación, lógica y datos.
>
> **Capa de presentación:** el frontend en React compilado a archivos estáticos vive en un bucket S3 con static website hosting. Cuando el usuario entra a la URL, S3 le sirve el HTML, CSS y JS. El frontend luego hace XHR al backend.
>
> **Capa de lógica:** el backend en Go corre en ECS Fargate, que es el servicio de containers serverless de AWS. Tenemos un cluster con dos services — blue y green — para implementar blue/green deployment. Las requests entran por un Application Load Balancer que distribuye tráfico al target group activo. Cada service tiene 2 tasks corriendo el container del backend, descargado de ECR (Elastic Container Registry, donde nuestro CI publica las imágenes).
>
> **Capa de datos:** DynamoDB con dos tablas — una para sesiones de juego activas con TTL nativo de 30 minutos, y otra para cachear las respuestas de Gemini. Los secretos como la API key de Gemini están en Secrets Manager, inyectados como env vars en la task definition.
>
> **Externos:** consumimos Data Dragon de Riot para los portraits de campeones — es CDN público sin auth — y Gemini API de Google para el lore.
>
> **Observabilidad:** todo se loguea en CloudWatch, donde tenemos un dashboard con 4 paneles, un log group por ambiente, y dos alarmas que el deploy script consulta para auto-rollback."

## Inventario de servicios AWS (por si te preguntan)

| Servicio | Para qué | Por qué |
|---|---|---|
| **S3** | Hostear frontend | Más barato y simple para archivos estáticos |
| **ECR** | Registry de imágenes Docker | Necesario para que ECS las jale |
| **ECS Fargate** | Correr containers backend | Serverless — no administramos servidores |
| **ALB** | Distribuir tráfico HTTP + blue/green | Permite swap del listener entre target groups |
| **DynamoDB** | Sesiones + caché Gemini | Serverless, TTL nativo, sin VPC privada |
| **Secrets Manager** | API key de Gemini | Cifrada en reposo, nunca en código |
| **CloudWatch** | Logs + métricas + alarmas + dashboard | Nativo AWS, integrado con ECS y ALB |
| **IAM (LabRole)** | Permisos para ECS | En Academy no podemos crear roles, usamos LabRole |

## Servicios que NO usamos y por qué

| Servicio | Por qué no |
|---|---|
| **EC2** | Fargate elimina el overhead de administrar VMs |
| **CloudFront** | Academy `voclabs` bloquea `cloudfront:Create*` |
| **RDS** | Requiere VPC privada + cuesta $15+/mes en reposo |
| **Lambda** | Backend HTTP de larga vida, no event-driven |
| **API Gateway** | El ALB ya da routing HTTP suficiente |
| **CodePipeline / CodeDeploy** | GitHub Actions es vendor-agnostic y mejor UI |

## Preguntas frecuentes y respuestas

### "¿Por qué Fargate y no EC2?"
> "Fargate es serverless — solo definimos CPU y memoria, AWS administra el host. EC2 nos hubiera obligado a mantener AMIs, parches del OS y monitoring de la VM. En Academy con LabRole esto sería más fricción que valor."

### "¿Por qué DynamoDB y no Postgres?"
> "Patrón de acceso 100% por clave única (gameId), sin queries complejas — es el caso ideal de DynamoDB. RDS requiere VPC privada con subnet groups, y cuesta $15+/mes en reposo aunque nadie lo use. DynamoDB es serverless — pagas por request, $0 en reposo."

### "¿Por qué ALB y no CloudFront/API Gateway?"
> "ALB nos da exactamente lo que necesitamos: terminación HTTP, health checks, dos target groups para blue/green. Probamos CloudFront para el frontend pero el LabRole de Academy lo bloquea. API Gateway sería redundante con el ALB."

### "¿Dónde está la API key de Gemini?"
> "No está en el código. Está en mi `terraform.tfvars` local que está gitignored, en AWS Secrets Manager cifrado, y en RAM del container al runtime. Pueden grepear el repo entero por `AIzaSy` (el prefijo de keys de Google) y no aparece."

---

# SECCIÓN 2 — Las 3 modificaciones (5-7 min)

## Mod 1 — Blue/Green con auto-rollback por alarmas CloudWatch

**Qué es:**
- Dos ECS services (blue y green), cada uno con su target group
- `scripts/deploy-app.sh` orquesta el swap del listener
- Después del swap: 5 min de observación contra alarmas CloudWatch (5xx rate, p95 latency)
- Si alguna alarma dispara → rollback automático del listener

**Por qué importa:** zero-downtime real + safety net. Si el deploy nuevo está roto, el sistema lo revierte solo.

**Demo:**
1. Abre pestaña **Actions** de GitHub
2. Muestra el último CD ejecutado verde
3. Click en `deploy-dev` → muestra los logs:
   - "Active: green, deploying to blue"
   - "Registered task def..."
   - "Waiting for blue to be stable..."
   - "Swapping listener default..."
   - "(1/30) alarms quiet" hasta "(30/30) alarms quiet"
   - "Draining old service..."
   - "Deploy successful: green -> blue"

**Frase para clase:**
> "El blue/green tiene 7 pasos: detecta el color activo, registra una nueva task definition con la imagen nueva, escala el color inactivo a 2 tasks, smoke test via X-Preview header, swap del listener default, ventana de observación 5 min contra alarmas CloudWatch — si dispara, rollback automático — si no, drena el color viejo."

## Mod 2 — DevSecOps multi-capa en CI

**Qué es:** 5 herramientas de seguridad corriendo automáticas en cada push:
- **gosec** — SAST en código Go (severity HIGH)
- **npm audit** — vulnerabilidades en deps de Node (HIGH+ bloquea)
- **hadolint** — best practices en Dockerfile
- **Trivy** — escaneo de CVEs en la imagen Docker (CRITICAL bloquea)
- **Secrets Manager** — credenciales fuera del código

**Por qué importa:** atrapa problemas de seguridad antes de que lleguen a prod, sin esfuerzo manual.

**Demo:**
1. Pestaña Actions → último CI exitoso
2. Click en job `backend` → muestra step "Gosec (SAST — HIGH severity)"
3. Click en job `docker` → muestra step "Trivy scan (CRITICAL fails)"
4. **Power move:** muestra el commit `39a2ad4` ("fix(ci): suppress gosec G404") — evidencia de que **gosec efectivamente detectó algo** que tuvimos que justificar/suprimir, no son scans de adorno.

**Frase para clase:**
> "Tenemos 5 capas de seguridad en CI: gosec hace SAST sobre el código Go, npm audit revisa dependencias del frontend, hadolint lintea el Dockerfile, Trivy escanea la imagen Docker buscando CVEs CRITICAL, y Secrets Manager mantiene la API key de Gemini fuera del código fuente. Cualquier hallazgo HIGH o CRITICAL bloquea el merge."

## Mod 3 — Observabilidad + IA con secretos seguros

**Qué es:** combo de 3 cosas:
- **CloudWatch dashboard** con 4 paneles (request count + 5xx, latency p50/p95/p99, ECS CPU/memory, target health)
- **Logging estructurado JSON** desde Go con `slog`
- **Integración con Gemini** para lore, con cache-aside en DynamoDB y graceful degradation
- **API key de Gemini en Secrets Manager**, inyectada en la task

**Por qué importa:** demuestra observabilidad real (no solo logs sueltos) + integración de IA con buenas prácticas de manejo de secretos.

**Demo:**
1. Abre el dashboard CloudWatch (URL arriba)
2. Muestra los 4 paneles con datos reales
3. Vuelve al juego, gana una partida → muestra el lore de Gemini renderizado en el WinBanner
4. Abre AWS Secrets Manager en otra pestaña → muestra el secret `lolidle/dev/gemini-api-key` (con valor "Hide" para no exponerlo)

**Frase para clase:**
> "Nuestro backend escribe logs JSON estructurados con slog. Esos logs llegan a CloudWatch Logs en el grupo `/ecs/lolidle-<env>`. CloudWatch automáticamente extrae métricas estándar del ALB y ECS — request count, errores, latencias, CPU. Sobre esas métricas definimos 2 alarmas críticas: 5xx rate y p95 latency. Esas alarmas alimentan el rollback automático del deploy script. Y todo lo importante está visualizado en un dashboard con 4 paneles para dev y otro para staging."

---

# SECCIÓN 3 — Demostración App funcional en cada ambiente (2-3 min)

## Opción A — Aplicar prod ahora (recomendado, ~15 min antes de presentar)

```bash
cd infra/envs/prod
cp terraform.tfvars.example terraform.tfvars
# editar y poner la gemini key (la que tienes en infra/envs/dev/terraform.tfvars)

terraform init -plugin-dir="$HOME/.terraform.d/plugins"
terraform apply -target=module.dynamodb -target=module.secrets -target=module.frontend -target=module.alb -target=module.observability -auto-approve
terraform apply -auto-approve
# después: build + push imagen, sync frontend
```

## Opción B — Defender que prod no está aplicado

> "Prod está completamente codificado e idéntico a staging, pero no lo aplicamos por restricciones de presupuesto Academy — un ALB cuesta ~$16/mes y tres en paralelo agotarían el saldo en pocos días. La estrategia es aplicarlo solo cuando se vaya a hacer un release real con un git tag v*.*.*."

## Guión para los ambientes vivos

> "Tenemos los ambientes corriendo con la misma versión del código.
>
> **Dev:** [abro URL dev]. El frontend carga, los 172 campeones aparecen en el autocomplete. Hago un guess... el backend compara y devuelve los 7 atributos con sus estados. Acertamos... aparece el WinBanner con el lore de Gemini.
>
> **Staging:** [abro URL staging]. La misma app, idéntica funcionalidad. Mismo SHA del commit, distinto ambiente.
>
> **Prod:** [si Opción A: abro URL prod, mismo flow]. La diferencia entre los ambientes no es código, es separación de blast radius — si rompemos dev, staging y prod siguen funcionando."

---

# BONUS — Quality Gates (por si te preguntan)

Tenemos **22 gates** distribuidos en 4 etapas. Cualquiera que falle bloquea el avance.

## CI (en cada push/PR) — bloquea el merge

### Backend
| # | Gate | Tool | Falla si... |
|---|---|---|---|
| 1 | Format | `gofmt -l .` | hay un .go sin formatear |
| 2 | Vet | `go vet ./...` | hay errores estáticos |
| 3 | Staticcheck | `staticcheck ./...` | hay code smells |
| 4 | SAST Go | `gosec -severity high` | vulnerabilidades HIGH |
| 5 | Tests | `go test -race` | algún test falla o detecta race |
| 6 | **Coverage** | `go tool cover` | coverage < **80%** |
| 7 | Build | `go build` | no compila |

### Frontend
| # | Gate | Tool | Falla si... |
|---|---|---|---|
| 8 | Lint | `eslint .` | errores de ESLint |
| 9 | Type check | `tsc --noEmit` | errores de tipos |
| 10 | Tests | `vitest --coverage` | algún test falla |
| 11 | **Deps audit** | `npm audit --audit-level=high` | vulnerabilidades HIGH+ |
| 12 | Build | `npm run build` | no buildea |

### Docker
| # | Gate | Tool | Falla si... |
|---|---|---|---|
| 13 | Dockerfile lint | `hadolint` | viola best practices |
| 14 | Build | `docker build` | no construye |
| 15 | **CVE scan** | `Trivy --severity CRITICAL` | hay alguna CVE CRITICAL |

## CD dev+staging — bloquea el deploy

| # | Gate | Validación | Falla si... |
|---|---|---|---|
| 16 | Service stable | `aws ecs wait services-stable` | tasks no pasan health check en ~5 min |
| 17 | Smoke test | `curl -H "X-Preview: green" /api/*` | no devuelven 200 |
| 18 | **Observation window** | poll 30×10s a alarmas | 5xx-rate o p95-latency dispara → **rollback automático** |
| 19 | E2E Playwright | tests contra dev URL | falla → staging NO se despliega |

## CD prod — bloquea el release

| # | Gate | Validación | Falla si... |
|---|---|---|---|
| 20 | **Approval manual** | GitHub Environment "prod" reviewer | reviewer no aprueba |
| 21 | ECR image existe | `aws ecr describe-images imageTag=v*.*.*` | el tag no fue construido por CI |
| 22 | Smoke prod | `curl /api/health` y `/api/champions` | no devuelven 200 |

## Thresholds concretos

| Threshold | Valor |
|---|---|
| Coverage backend | ≥ 80% |
| Severity gosec | HIGH+ bloquea |
| Severity npm audit | HIGH+ bloquea |
| Severity Trivy | CRITICAL bloquea |
| 5xx errors | > 5 en 60s × 2 períodos |
| Latency p95 | > 2 segundos × 2 períodos |
| E2E timeout por test | 60 segundos |
| Observation window | 5 min (30 × 10s polls) |

## Frase para clase

> "Tenemos 22 quality gates distribuidos en 4 etapas. **CI** valida 15 cosas en cada push: formato, lint, tests, coverage del 80%, SAST con gosec, npm audit en HIGH+, hadolint en el Dockerfile, y Trivy en la imagen para CVEs CRITICAL. Si todo pasa, **CD dev** valida que las nuevas tasks ECS estén estables, que los smoke tests respondan 200, y vigila las alarmas CloudWatch durante 5 minutos para auto-rollback. Antes de pasar a staging, **Playwright** corre tests E2E contra dev. Para **prod** además se requiere approval manual via GitHub Environment, validación de que el tag exista en ECR, y un smoke test final post-swap."

---

# Checklist final pre-presentación

- [ ] Renovar creds Vocareum (expiran cada ~4h)
- [ ] Actualizar 3 secrets de GitHub si vas a hacer demo de pipeline en vivo
- [ ] Verificar las URLs cargan abriéndolas 30 min antes
- [ ] Pre-calentar el dashboard generando tráfico:
  ```bash
  ALB="http://lolidle-dev-alb-925096534.us-east-1.elb.amazonaws.com"
  for i in {1..50}; do
    curl -s "$ALB/api/champions" > /dev/null
    curl -s "$ALB/api/health" > /dev/null
  done
  ```
- [ ] Tabs abiertas:
  - GitHub → Actions
  - CloudWatch → Dashboard `lolidle-dev`
  - AWS Secrets Manager → `lolidle/dev/gemini-api-key`
  - Frontend dev (con juego empezado)
  - Frontend staging
- [ ] (Opcional) Aplicar prod si quieres demo de los 3 ambientes
- [ ] Practicar el guión de cada sección 1-2 veces

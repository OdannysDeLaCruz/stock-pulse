<p align="center" width="100%">
  <img src="https://superbeauty-public.s3.us-east-1.amazonaws.com/images/stocks-pulse.jpg" />
</p>

# 📈 Stocks Pulse - Real-Time Stock Monitoring

Un sistema en tiempo real para la actualización y visualización de precios de acciones, utilizando **WebSockets, Redis y Go**. 🚀

---

## 🌟 Features

- 📡 **WebSockets** para comunicación en tiempo real con los clientes.
- 🔥 **Simulación de datos** con actualización en tiempo real de precios de acciones.
- 💾 **Almacenamiento en CockroachDB** y persistencia en base de datos.
- ⚡ **Pub/Sub con Redis** para comunicación/integración desacoplada
- 🔄 **Actualización cada 5 segundos** con información en vivo.
- 🎯 **Soporte para múltiples acciones**, permitiendo suscripción a una o varias.
- 🛠️ **Frontend en Vue.js** para visualización de datos en tiempo real.

---

## 🏗️ Arquitectura

1. **Simulador de Precios:** Genera datos aleatorios de acciones y los publica en Redis.
2. **Redis Pub/Sub:** Recibe y distribuye las actualizaciones a los servicios suscritos.
3. **Backend en Go:**
   - Consume los datos desde mensajes de eventos con Redis.
   - Envía actualizaciones a los clientes vía WebSockets.
   - Persiste los datos en la base de datos.
4. **Frontend en Vue.js:**
   - Se conecta a WebSockets para recibir actualizaciones.
   - Muestra gráficos y datos en tiempo real.
   - Ver repositorio aquí -> [https://github.com/OdannysDeLaCruz/stock-pulse-frontend](https://github.com/OdannysDeLaCruz/stock-pulse-frontend)

---

## 🚀 Cómo Ejecutar el Proyecto

### 📦 Requisitos

- Docker & Docker Compose
- **Go 1.23.2+** (viene listo para usar en la configuración del proyecto)
- **Redis** (viene listo para usar en la configuración del proyecto)
- **CockroachDB** (viene listo para usar en la configuración del proyecto)

### 🏃‍♂️ Pasos para levantar el proyecto

#### 1️⃣ Clonar el repositorio

```bash
$ git clone https://github.com/OdannysDeLaCruz/stock-pulse.git
$ cd stock-pulse
```

#### 2️⃣ Configurar variables de entorno

```bash
$ cp .env.example .env
```

El nombre de la base de datos es el nombre del servicio `cockroachdb`, o el nombre que le quieras configurar al servicio. Pasa lo mismo con redis, se debe usar el nombre del servicio configurado en el docker-compose.yml, en este caso es `redis`. El `user`, `dbname` y el `port` serán los que configuras al crear la base de datos desde `docker-compose.yml` en el servicio `db-init`. En nuestro caso, usaremos `user=root`, `dbname=stocks_pulse_db` y el puerto por defecto de la base de datos `cockroachdb` que es `port=26257`.

#### 3️⃣ Levantar entorno con Docker

Cuando tengas todas las variables solicitadas en `.env.example`, entonces puedes levantar los servicios.

```bash
$ docker-compose up --build -d
```

No te olvides de tener tu docker daemon corriendo.

La configuración del archivo `docker-compose.yml` encontrará 5 servicios:

- `redis`: inicia base de datos redis
- `redis-labs`: inicia un gestor gráfico para redis
- `cockroachdb`: inicia base de datos Cockroach y un dashboard de administración
- `db-init`: se asegura que exista la base de datos llamada `stocks_pulse_db`
- `stock-pulse`: inicia backend stock pulse

#### 5️⃣ Ejecutar migraciones

Hasta este punto solo hay una base de datos llamada `stocks_pulse_db`, sin ninguna tabla en su interior. Para crear las tablas necesarias se deben ejecutar las `"migraciones"`.

Para ejecutar las migraciones solo debe ejecutar el siguiente comando:

```Bash
$ docker-compose -f docker-compose.migrate.yml up --build -d
```

Esto ejecutará un servicio que creará 2 tablas: `stocks` y `stock_price_histories`.

### 6️⃣ Popular la base de datos (seed)

Con los servicios arribas y la base de datos lista, se procede a obtener los datos devueltos por el api. Estos datos servirán para poblar la base de datos. Solo tiene que ejecutar el endpoint `PUT /stocks` (Endpoint descritos mas abajo).

Eso debe bastar para tener un total de `50` stocks en la tabla `stocks`. Esto no genera datos en la tabla `stock_price_history`, para generar historial de precios a un stock, continua con el siguiente paso.

#### 7️⃣ Ejecutar el simulador de cambio de precios

El simulador de cambio de precio hace simulaciones de cambios en los precios de los stocks, tal como haría las proyección un broker o una agencia financiera acciones. En la vida real este sería un sistema aparte que envía datos en tiempo real o cada cierto tiempo de las proyecciones de las acciones.

Este paquete no tiene acceso a la base de datos, solo se encarga de enviar eventos al sistema de cola de nuestro sistema (`Redis Pub/Sub`). Y nuestro sistema se encarga de actualizar la base de datos y enviar los mensajes `Websocket` a todos los clientes suscritos a los stocks que fueron actualizados en ese momento.

Por temas de facilidad, este paquete espera que se configure una lista finita de stocks a los cuales les hará el seguimiento. También puede configurar una lista de rating (rating to).

```go
// Lista de acciones simuladas
stocksTicker := []string{'NVDA', 'NFLX'}

// Lista de estados de rating simulados
ratingList := []string{'Buy', 'Hold', 'Sell', 'Strong-Buy', 'Strong-Sell', 'Neutral', 'Overweight', 'Underweight', 'Outperform', 'Underperform'}
```

Esta lista de símbolos de stocks deben existir en la base de datos previamente. Para ejecutar el simulador ejecuta el siguiente comando:

```bash
$ docker-compose -f docker-compose.simulator.yml up --build -d
```

Se enviarán datos a `Redis Pub/Sub` cada 5 sg. Asegúrate de utilizar la misma conexión que usa el proyecto.

#### 8️⃣ Ejecutar el frontend

Para ejecutar la aplicación frontend por favor siga los paso descritos en el siguiente repositorio:

[https://github.com/OdannysDeLaCruz/stock-pulse-frontend](https://github.com/OdannysDeLaCruz/stock-pulse-frontend)

---

## 🎯 Endpoints

- **WebSockets:**
  - `ws://localhost:8080/ws?ticker=NVDA` para suscribirse a una acción específica.
- **REST API:**
  - `GET /stocks` → Últimos precios almacenados.
  - `GET /stocks/:symbol` → Obtener todos los datos de una acción específica.
  - `GET /stocks/:symbol/history` → Obtener histórico de precio de una acción específica.
  - `GET /stocks/recommendations` → Acciones recomendadas del día.
  - `GET /stocks/not-recommended` → Acciones no recomendadas del día.
  - `GET /stocks/search?q=NVDA` → Buscar acción por símbolo o nombre de la empresa.
  - `PUT /stocks` → Popular (seeding) base de datos desde api externa (solo para entorno de pruebas).

---

## 🛠 Tecnologías Usadas

### Backend

- Go
- Gin Web Framework - [https://gin-gonic.com/](https://gin-gonic.com/)
- GORM, ORM Library for Golan - [https://gorm.io/](https://gorm.io/)
- Gorilla WebSocket - [https://github.com/gorilla/websocket](https://github.com/gorilla/websocket)
- Redis Pub/Sub
- Base de Datos: CockroachDB
- Mensajería en tiempo real: Redis

### Frontend 

(ver en [https://github.com/OdannysDeLaCruz/stock-pulse-frontend](https://github.com/OdannysDeLaCruz/stock-pulse-frontend))

- Vue 3
- Pinia
- TailwindCSS
- Trading View Lightweight Charts - [https://tradingview.github.io/lightweight-charts/docs](https://tradingview.github.io/lightweight-charts/docs)

## 📌 Contribuciones

¡Las contribuciones son bienvenidas! Abre un issue o envía un PR. 💡🔥

---

## Generador de README ?

No, no usé ningún generador o plantilla, todo fue escrito meticulosamente desde cero por mí 😉.

---

## 📜 Licencia

MIT License. Hecho con ❤️ por [Odannys De La Cruz](https://github.com/OdannysDeLaCruz)

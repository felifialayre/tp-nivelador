# Informe — TP Nivelador

En este TP se desarrolló un sistema distribuido para simular una lotería mediante
varias agencias (clientes) que envían apuestas y un servidor capaz de recibir dichas
apuestas concurrentemente (varios clientes a la vez) para finalmente responder con los
ganadores correspondientes a cada agencia.

## Arquitectura

- **Cliente (Go):** lee las apuestas del `INPUT_FILE`, las envía en lotes,
  recibe sus ganadores y los persiste en `OUTPUT_FILE`.
- **Servidor (Python):** acepta conexiones concurrentes, almacena las apuestas,
  espera un quórum mínimo de agencias y luego responde a cada una con sus
  ganadores.

## Protocolo de comunicación

El protocolo implementado es de tipo request-reply sobre TCP stateful por lo que el funcionamiento 
del mismo depende del correcto seguimiento entre los siguientes estados:

- Etapa de saludo: Al iniciar la comunicación el cliente envía un paquete de saludo que contiene
  solo al `agency_id` de la agencia en cuestión para que el servidor lo registre
- Etapa de envío de apuestas: Aquí el cliente enviará la información de todas las apuestas
  serializándolas de a batches y esperando el ACK del servidor para enviar la siguiente. Nótese
  que la agencia informará que no enviará más apuestas enviando un paquete con payload vacío
  con el opcode correspondiente.
- Etapa de envío de ganadores: En esta etapa, cada instancia de `ClientHandler`enviará a la
  agencia la lista de los ganadores correspondientes a la misma. 
- Etapa final: Una vez recibidos los ganadores se corta la comunicación y da por terminada la
  transferencia.


### Formato de frame

```
[ opcode : 1 byte ][ length : 4 bytes ][ payload : length bytes ]
```

Tipos de mensaje por sus `opcode`

| Opcode  | Valor | Dirección        | Significado                          |
|---------|-------|------------------|--------------------------------------|
| BATCH   | 0     | cliente→servidor | lote de apuestas                     |
| END     | 1     | cliente→servidor | fin del envío de apuestas            |
| ACK     | 2     | servidor→cliente | lote procesado correctamente         |
| WINNERS | 3     | servidor→cliente | listado de ganadores de la agencia   |
| HELLO   | 4     | cliente→servidor | saludo inicial del cliente           |

### Serialización de una apuesta

```
[ len_nombre : 1 ][ nombre ][ len_apellido : 1 ][ apellido ]
[ nacimiento : 10 ][ documento : 4 ][ numero : 4 ]
```

El payload de un `BATCH` es `[ cantidad : 4 ]` seguido de N apuestas serializadas.

### Flujo de una ronda

```
cliente                          servidor
  | ---- HELLO(agency_id) ------> |
  | ---- BATCH -----------------> | store_bets()
  | <--- ACK -------------------- |
  |            (se repite)        |
  | ---- END -------------------> |
  |                               | (espera quórum)
  | <--- WINNERS ---------------- | load_bets() + has_won()
```

## Concurrencia y sincronización (servidor)

Se usa multiprocessing: el servidor acepta conexiones en un loop y por cada
cliente lanza un proceso. Se eligió procesos para no depender del GIL.

Primitivas de sincronización:

- **`multiprocessing.Barrier(AGENCY_QUORUM_MIN)`**: cada handler, tras recibir
  todas las apuestas de su agencia, hace `barrier.wait()`. El sorteo no se
  resuelve hasta que se juntan las `AGENCY_QUORUM_MIN` agencias. La barrera se
  resetea sola, lo que permite múltiples rondas sin reiniciar el servidor.
- **`multiprocessing.Lock`** (en `SafeLottery`): serializa el acceso al archivo
  de apuestas, evitando la race entre `store_bets` (escritura) de un handler y
  `load_bets` (lectura) de otro.

Cada handler responde **solo con los ganadores de su propia agencia** (no hay
broadcast).

## Cierre graceful (SIGTERM)

El cierre está acotado en el tiempo (`SHUTDOWN_TIMEOUT_SECONDS`), por debajo del
plazo de `docker compose stop -t`.

### Servidor

- Un handler de `SIGTERM` en el proceso principal cierra el socket de escucha
  (lo que destraba el `accept()`) y aborta la barrera (`barrier.abort()`), que
  despierta a los handlers que estuvieran esperando quórum.
- Al salir del loop de accept, el proceso principal propaga la señal a los hijos
  (`terminate()`), los espera con timeout (`join`).
- Cada `ClientHandler` instala su propio handler de `SIGTERM` que cierra su
  socket, y un `finally` garantiza el cierre del socket en todos los caminos
  (fin normal, error o `BrokenBarrierError` por el abort).

### Cliente

- `main` traduce la señal a una cancelación con `signal.NotifyContext`.
- Una goroutine espera `ctx.Done()` y **cierra la conexión**, lo que destraba
  cualquier `Send`/`Recv` bloqueado.
- Los `defer` (conexión y archivos) garantizan la liberación de recursos.
- El error que produce el cierre de la conexión se distingue de un error real
  chequeando `ctx.Err()`: si el contexto fue cancelado, la aplicación termina de
  forma limpia.

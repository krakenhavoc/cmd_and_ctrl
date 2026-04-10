# Protocol v0

Wire protocol between the `cmd_and_ctrl` Go game server and its TypeScript
clients. Version 0 is the absolute minimum needed to prove the seam in S01:
one request type, one response type, one error type. S02 and S03 expand it
into the real action / state-delta protocol.

- **Transport:** WebSocket (`ws://host:8080/ws`, `wss://` in production)
- **Framing:** one JSON object per WebSocket text frame
- **Authoritative direction:** the server. Clients propose actions; the
  server accepts, validates, and emits state deltas.
- **Version negotiation:** every frame carries a `v` field (integer). Server
  refuses any frame whose `v` does not match its supported version.

---

## Frame envelope

Every frame is a JSON object with at least:

```json
{
  "v": 0,
  "kind": "<kind>",
  "id": "<uuid-v4>"
}
```

- `v` — protocol version. Currently `0`.
- `kind` — discriminator. Each kind has its own additional fields below.
- `id` — client-generated UUIDv4 (client → server frames) or
  server-generated (server → client frames). Used to correlate
  responses/errors to the originating action.

Unknown fields MUST be ignored by readers to allow forward-compatible
additions within a version.

---

## v0 frame kinds

### `ping` (client → server)

A no-op action used to prove the WebSocket round-trip. The server replies
with a `pong` delta carrying the same `id`.

```json
{
  "v": 0,
  "kind": "ping",
  "id": "3c2f8d7e-6b5a-4f1e-9a2c-0d8e7f6a5b4c",
  "payload": {
    "msg": "hello"
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `payload.msg` | string | no | Echoed back in the `pong`. For S01 debugging only. |

### `pong` (server → client)

Server response to a `ping`.

```json
{
  "v": 0,
  "kind": "pong",
  "id": "3c2f8d7e-6b5a-4f1e-9a2c-0d8e7f6a5b4c",
  "payload": {
    "msg": "hello",
    "server_time": "2026-04-10T17:42:31Z"
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `payload.msg` | string | no | Echoed from the originating `ping`. Omitted if the ping carried no msg. |
| `payload.server_time` | string (RFC3339) | yes | Server wall clock at the moment of reply. |

### `error` (server → client)

Any server-side failure in processing a client frame.

```json
{
  "v": 0,
  "kind": "error",
  "id": "3c2f8d7e-6b5a-4f1e-9a2c-0d8e7f6a5b4c",
  "payload": {
    "code": "bad_request",
    "message": "unknown kind 'explode'"
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `payload.code` | string | yes | Short machine-readable code. v0 codes: `bad_version`, `bad_json`, `bad_request`, `internal`. |
| `payload.message` | string | yes | Human-readable message safe to display to the user. |

`id` matches the originating client frame's `id` when possible; otherwise
empty string.

---

## Connection lifecycle (v0)

1. Client opens WebSocket to `/ws`.
2. Server accepts, optionally logs client IP.
3. Client and server exchange frames. No authentication at v0 (S04 adds it).
4. Either side may close at any time. Server closes with a standard close
   code; client handles reconnection itself.

---

## Out of scope for v0

Explicitly deferred to v1 (S02) and v2 (S03):

- Game creation / join / leave
- Player identity and seat assignment
- Any game state (zones, cards, turn structure)
- Any game actions (draw, play, tap, pass priority)
- State deltas larger than `pong`
- Reconnection and session resume
- Server push of lobby events
- Spectator mode
- Chat

The rule of thumb for v0: **if it's not `ping` or `pong`, it waits for v1.**

---

## Schema evolution rules

- **Breaking changes** bump `v` and require updating both server and client
  in lockstep. Keep breaking changes rare.
- **Additive changes within a version** (new optional fields, new error
  codes) are allowed and do not bump `v`. Readers ignore unknown fields.
- **Removing a field** is always breaking. Don't do it without a `v` bump.
- Each version of this document lives at `docs/protocol.md`; previous
  versions are preserved in git history and can be retrieved by commit hash
  when diagnosing issues with old deployments.

---

## References

- RFC 6455 — The WebSocket Protocol
- [ADR 0001 — WebSocket library](decisions/0001-ws-library.md)
- Sprint S02 — [action protocol + state deltas](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3)

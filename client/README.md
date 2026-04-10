# client

The `cmd_and_ctrl` web client. TypeScript + Vite + Svelte 5 + (eventually)
PixiJS. See the [top-level PLAN.md](../PLAN.md) and [AGENTS.md](../AGENTS.md).

## Running

```bash
# install deps (first time)
npm install

# dev server on :5173, proxies /ws to ws://localhost:8080
npm run dev

# build and type-check
npm run build

# lint + format check
npm run lint

# auto-format
npm run format

# unit tests
npm test
```

The Vite dev server proxies `/ws` to the Go game server, so run the server
in another terminal (`cd ../server && make dev`) and open
[http://localhost:5173](http://localhost:5173) to drive the ping round-trip.

## Layout

```
client/
├── src/
│   ├── main.ts             # app entry point, mount(App, ...)
│   ├── App.svelte          # S01 demo UI: send ping, show log
│   ├── app.css             # global styles
│   └── lib/
│       ├── protocol.ts     # hand-maintained mirror of server protocol types
│       └── ws.ts           # WebSocket client + Svelte stores
├── index.html              # Vite entry HTML
├── vite.config.ts          # Vite + /ws proxy to Go server
├── svelte.config.js        # Svelte preprocessor
├── tsconfig.json
├── eslint.config.js        # ESLint 9 flat config
├── .prettierrc
├── .prettierignore
└── package.json
```

## Architecture notes

- State from the server lives in Svelte stores exposed by `GameClient`
  (`lib/ws.ts`). Components subscribe to stores; they never call the
  WebSocket directly.
- The protocol types in `lib/protocol.ts` are a hand-maintained mirror of
  `server/internal/protocol/protocol.go`. When the spec in
  `docs/protocol.md` changes, update both sides in lockstep.
- Svelte 5 runes (`$state`, etc.) are used throughout.

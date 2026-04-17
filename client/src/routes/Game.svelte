<script lang="ts">
  import { GameClient } from "../lib/ws";
  import { navigate } from "../lib/router";
  import { session } from "../lib/session";
  import { TableRenderer } from "../lib/table";
  import type { GameView } from "../lib/protocol";

  interface Props {
    gameID: string;
  }
  const { gameID }: Props = $props();

  // Build the WS URL from the stored session + route. Query string
  // carries the session token (browsers can't send Authorization on
  // WS upgrade) and identifies which seat to render. If the session
  // has a player_id bound to this game, we pin it; admin sessions
  // may omit ?player= and fall through to the spectator view.
  const baseURL = (location.protocol === "https:" ? "wss://" : "ws://") + location.host + "/ws";
  const sess = $derived($session);
  const wsURL = $derived.by(() => {
    const params = new URLSearchParams();
    params.set("game", gameID);
    if (sess?.token) params.set("token", sess.token);
    if (sess?.playerID && sess.gameID === gameID) params.set("player", sess.playerID);
    return `${baseURL}?${params.toString()}`;
  });

  // One GameClient per component instance. Start with an empty URL
  // — the $effect below installs the real URL on first run and
  // reconnects whenever wsURL changes. This preserves the store
  // subscribers across reactive reruns (vs. replacing the client
  // instance, which would strand subscriptions on the old object).
  const client = new GameClient("");
  const { status, snapshot, lastSeq } = client;

  $effect(() => {
    client.disconnect();
    client.setURL(wsURL);
    client.connect();
    return () => client.disconnect();
  });

  // PixiJS table renderer. init() is async; we track the in-flight
  // init so snapshots that arrive before the canvas is ready don't
  // try to render into a half-built application. destroyed flags
  // guard against late init callbacks firing after the component
  // has already torn down (HMR, route change).
  let canvasEl: HTMLDivElement | undefined = $state();
  let renderer: TableRenderer | null = null;
  let rendererReady = $state(false);

  $effect(() => {
    if (!canvasEl) return;
    const r = new TableRenderer();
    let destroyed = false;
    // Pixi's Application.destroy is not idempotent — guard so both
    // the cleanup callback and a late-resolving init promise can
    // dispatch dispose() without double-firing the underlying call.
    let disposed = false;
    const dispose = (): void => {
      if (disposed) return;
      disposed = true;
      r.destroy();
    };
    void r.init(canvasEl).then(() => {
      if (destroyed) {
        dispose();
        return;
      }
      renderer = r;
      rendererReady = true;
    });
    return () => {
      destroyed = true;
      if (renderer === r) {
        renderer = null;
        rendererReady = false;
      }
      dispose();
    };
  });

  // sendAction is a thin shim over GameClient.sendAction that the
  // renderer invokes from interactive events. Bound via $derived so
  // session changes (logout + re-login) pick up a fresh viewer ID.
  const sendAction = (type: string, params?: unknown, player?: string): void => {
    client.sendAction(type, player, params);
  };

  // Re-render whenever a new snapshot arrives OR the renderer just
  // finished initialising. Reading $snapshot and rendererReady in the
  // same $effect ties both reactive inputs to the redraw.
  $effect(() => {
    const view: GameView | null = $snapshot;
    if (!renderer || !rendererReady || !view) return;
    renderer.render(view, { viewerID: sess?.playerID ?? null, sendAction });
  });

  // Watch the container size. PIXI's resizeTo handles the canvas
  // sizing; we just need to trigger a redraw so the layout recomputes
  // against the new dimensions.
  $effect(() => {
    if (!canvasEl) return;
    const obs = new ResizeObserver(() => {
      if (!renderer || !rendererReady) return;
      const view = $snapshot;
      if (view) renderer.render(view, { viewerID: sess?.playerID ?? null, sendAction });
    });
    obs.observe(canvasEl);
    return () => obs.disconnect();
  });

  function back(): void {
    navigate("#/lobby");
  }

  // ---- Testing affordances (superseded by S07 turn/phase UI) ----
  // S06 exit criteria reference "draws 7 cards" and "passes turn";
  // the proper turn/phase bar arrives in S07. Until then these three
  // buttons expose the actions needed to drive a game manually.
  function openingHand(): void {
    if (!sess?.playerID) return;
    client.sendAction("mulligan", sess.playerID, { hand_size: 7 });
  }
  function passTurn(): void {
    client.sendAction("pass_turn");
  }
  function shuffle(): void {
    if (!sess?.playerID) return;
    client.sendAction("shuffle_library", sess.playerID);
  }
</script>

<section>
  <header>
    <button onclick={back}>← lobby</button>
    <h1>game {gameID.slice(0, 8)}</h1>
    <span class={`tag tag-${$status}`}>{$status}</span>
    <span class="muted">seq {$lastSeq}</span>
  </header>

  {#if $snapshot && sess?.playerID}
    <div class="dev-controls" aria-label="testing controls (temporary until S07)">
      <button onclick={openingHand}>draw 7 (mulligan)</button>
      <button onclick={shuffle}>shuffle library</button>
      <button onclick={passTurn}>pass turn</button>
      <span class="muted"
        >· click library to draw · click hand card to play · click battlefield card to tap · drag to
        reposition</span
      >
    </div>
  {/if}

  <div class="table" bind:this={canvasEl}></div>

  {#if !$snapshot}
    <p class="muted centered">waiting for snapshot…</p>
  {/if}
</section>

<style>
  section {
    max-width: 1280px;
    margin: 1rem auto;
    padding: 1rem;
  }
  header {
    display: flex;
    gap: 0.75rem;
    align-items: baseline;
    margin-bottom: 0.75rem;
  }
  .table {
    width: 100%;
    height: min(720px, 70vh);
    border-radius: 6px;
    overflow: hidden;
  }
  .muted {
    color: #888;
  }
  .centered {
    text-align: center;
    margin-top: 1rem;
  }
  .tag {
    padding: 0.1rem 0.4rem;
    border-radius: 3px;
    font-size: 0.8em;
  }
  .tag-connected {
    background: #cfc;
  }
  .tag-connecting {
    background: #ffc;
  }
  .tag-disconnected {
    background: #fcc;
  }
  .dev-controls {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    flex-wrap: wrap;
    margin-bottom: 0.5rem;
    padding: 0.4rem 0.5rem;
    background: #1a2540;
    border-radius: 4px;
    color: #bbc4dd;
    font-size: 0.85em;
  }
  .dev-controls button {
    padding: 0.25rem 0.6rem;
    font-size: 0.9em;
  }
</style>

<script lang="ts">
  // Dev-only replay scrubber (ADR 0023): step a recorded game frame
  // by frame on the real board.
  //
  // Client-only. GET /games/{id}/replay has existed since S11 and
  // already streams JSONL of complete unfiltered snapshots, so
  // "render frame N" is picking a line and handing it to the same
  // Board the live game uses — no re-simulation, no new route.
  //
  // Selecting a frame puts Game.svelte into replay mode: the board
  // renders the past state and the quick-action toolbar is hidden,
  // because the actions would otherwise mutate the LIVE game while
  // you are looking at a snapshot of the past.
  import { fetchReplay } from "../../api";
  import { parseReplay, frameLabel, describeDelta, type ReplayFrame } from "../../replay";

  interface Props {
    gameID: string;
    // Index of the frame being rendered, or null for live.
    index: number | null;
    onselect: (frame: ReplayFrame | null, index: number | null) => void;
  }
  const { gameID, index, onselect }: Props = $props();

  let frames: ReplayFrame[] = $state([]);
  let loading = $state(false);
  let error: string | null = $state(null);
  let loaded = $state(false);
  let playing = $state(false);
  let speed = $state(500);

  const current = $derived(index === null ? null : (frames[index] ?? null));
  const prev = $derived(index === null || index === 0 ? null : (frames[index - 1] ?? null));

  async function load() {
    loading = true;
    error = null;
    try {
      const parsed = parseReplay(await fetchReplay(gameID));
      frames = parsed;
      loaded = true;
      if (parsed.length) select(parsed.length - 1);
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    } finally {
      loading = false;
    }
  }

  function select(i: number | null) {
    if (i === null) {
      onselect(null, null);
      return;
    }
    const clamped = Math.max(0, Math.min(frames.length - 1, i));
    onselect(frames[clamped] ?? null, clamped);
  }

  function step(delta: number) {
    if (index === null) return;
    playing = false;
    select(index + delta);
  }

  // Playback. Deliberately a timeout chain rather than an interval:
  // an interval keeps firing after the component unmounts or the last
  // frame is reached, and stopping it correctly is more code than
  // this.
  $effect(() => {
    if (!playing || index === null) return;
    if (index >= frames.length - 1) {
      playing = false;
      return;
    }
    const t = setTimeout(() => select(index + 1), speed);
    return () => clearTimeout(t);
  });
</script>

<div class="scrubber">
  {#if !loaded}
    <div class="load">
      <button onclick={load} disabled={loading}>
        {loading ? "loading…" : "load replay"}
      </button>
      <p class="hint">
        Downloads this game's recorded snapshots. Every frame is the unfiltered view — you will see
        every hand.
      </p>
      {#if error}<p class="error" role="alert">{error}</p>{/if}
    </div>
  {:else if frames.length === 0}
    <p class="hint">
      No frames recorded yet. The log gets a line per applied action, so a game that has not been
      acted on has an empty replay.
    </p>
  {:else}
    <div class="bar">
      <button onclick={() => select(0)} disabled={index === 0} title="first">|&lt;</button>
      <button onclick={() => step(-1)} disabled={index === 0} title="previous">&lt;</button>
      <button onclick={() => (playing = !playing)} title={playing ? "pause" : "play"}>
        {playing ? "||" : "&gt;"}
      </button>
      <button onclick={() => step(1)} disabled={index === frames.length - 1} title="next"
        >&gt;</button
      >
      <button
        onclick={() => select(frames.length - 1)}
        disabled={index === frames.length - 1}
        title="last">&gt;|</button
      >

      <input
        type="range"
        min="0"
        max={frames.length - 1}
        value={index ?? frames.length - 1}
        oninput={(e) => {
          playing = false;
          select(Number((e.currentTarget as HTMLInputElement).value));
        }}
        aria-label="Replay position"
      />
      <span class="pos">{(index ?? 0) + 1}/{frames.length}</span>

      <select bind:value={speed} aria-label="Playback speed">
        <option value={1000}>1×</option>
        <option value={500}>2×</option>
        <option value={200}>5×</option>
        <option value={60}>fast</option>
      </select>

      <button class="live" onclick={() => select(null)} disabled={index === null}>
        back to live
      </button>
    </div>

    <dl class="readout">
      <dt>frame</dt>
      <dd>{frameLabel(current)}</dd>
      <dt>changed</dt>
      <dd>{describeDelta(prev, current)}</dd>
    </dl>

    <button class="reload" onclick={load} disabled={loading}>
      {loading ? "reloading…" : "reload (pick up newer frames)"}
    </button>
  {/if}
</div>

<style>
  .scrubber {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.6rem;
    overflow-y: auto;
  }
  .load {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    align-items: flex-start;
  }
  .bar {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    flex-wrap: wrap;
  }
  button,
  select {
    padding: 0.2rem 0.5rem;
    border: 1px solid var(--border, #273049);
    border-radius: var(--radius-sm, 4px);
    background: var(--surface, #111a2e);
    color: var(--fg-muted, #9aa5cd);
    font: inherit;
    cursor: pointer;
  }
  button:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  button:not(:disabled):hover {
    color: var(--fg, #e6ecff);
  }
  .live {
    margin-left: auto;
    border-color: var(--gold, #ffd07a);
    color: var(--gold, #ffd07a);
  }
  input[type="range"] {
    flex: 1;
    min-width: 8rem;
    accent-color: var(--gold, #ffd07a);
  }
  .pos {
    min-width: 4.5em;
    text-align: right;
    color: var(--fg-dim, #6c7a99);
  }
  .readout {
    display: grid;
    grid-template-columns: 5em 1fr;
    gap: 0.15rem 0.6rem;
    margin: 0;
  }
  .readout dt {
    color: var(--fg-dim, #6c7a99);
  }
  .readout dd {
    margin: 0;
    color: var(--fg, #e6ecff);
  }
  .reload {
    align-self: flex-start;
  }
  .hint {
    margin: 0;
    color: var(--fg-dim, #6c7a99);
  }
  .error {
    margin: 0;
    color: var(--danger, #ff7a7a);
  }
</style>

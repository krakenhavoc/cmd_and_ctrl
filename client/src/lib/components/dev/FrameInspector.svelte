<script lang="ts">
  // Dev-only raw WebSocket frame log.
  //
  // Mounted by Game.svelte only when the server reports env=dev AND
  // the frame_inspector feature. Purely client-side: it reads frames
  // the GameClient already handled, so there is no route to gate and
  // nothing here is privileged. The value is turning "the board
  // desynced" into a frame id, a seq number, and a payload you can
  // paste into a bug report.
  import type { GameClient, FrameRecord } from "../../ws";

  interface Props {
    client: GameClient;
  }

  const { client }: Props = $props();

  const frames = $derived(client.frames);

  let open = $state(false);
  let selectedID: string | null = $state(null);
  let filter: "all" | "in" | "out" = $state("all");
  // Kinds are low-cardinality; a free-text box would be overkill.
  let kindFilter = $state("");
  let paused = $state(false);
  // Snapshot taken when paused, so the list stops moving under the
  // cursor while you read a payload.
  let frozen: FrameRecord[] = $state([]);

  const source = $derived(paused ? frozen : $frames);

  const shown = $derived(
    source.filter(
      (f) => (filter === "all" || f.dir === filter) && (kindFilter === "" || f.kind === kindFilter),
    ),
  );

  const kinds = $derived([...new Set(source.map((f) => f.kind))].sort());

  const selected = $derived(shown.find((f) => f.id + f.at === selectedID) ?? null);

  function togglePause() {
    paused = !paused;
    frozen = paused ? [...$frames] : [];
  }

  function fmtTime(ms: number): string {
    const d = new Date(ms);
    return (
      d.toLocaleTimeString("en-US", { hour12: false }) +
      "." +
      String(d.getMilliseconds()).padStart(3, "0")
    );
  }

  function fmtBytes(n: number): string {
    return n < 1024 ? `${n}B` : `${(n / 1024).toFixed(1)}K`;
  }

  async function copySelected() {
    if (!selected) return;
    try {
      await navigator.clipboard.writeText(JSON.stringify(selected.frame, null, 2));
    } catch {
      // Clipboard can be denied; the payload is on screen and
      // selectable either way, so this is not worth a toast.
    }
  }
</script>

<div class="inspector" class:open>
  <button class="handle" onclick={() => (open = !open)} aria-expanded={open}>
    <span class="label">FRAMES</span>
    <span class="count">{$frames.length}</span>
  </button>

  {#if open}
    <div class="body">
      <div class="toolbar">
        <div class="seg" role="group" aria-label="Direction filter">
          {#each ["all", "in", "out"] as const as d (d)}
            <button class:active={filter === d} onclick={() => (filter = d)}>{d}</button>
          {/each}
        </div>
        <select bind:value={kindFilter} aria-label="Frame kind filter">
          <option value="">all kinds</option>
          {#each kinds as k (k)}<option value={k}>{k}</option>{/each}
        </select>
        <button onclick={togglePause} class:active={paused}>
          {paused ? "resume" : "pause"}
        </button>
        <button onclick={() => client.clearFrames()}>clear</button>
      </div>

      <div class="split">
        <ol class="list">
          {#each shown as f (f.id + f.at)}
            <li>
              <button
                class="row"
                class:sel={selectedID === f.id + f.at}
                onclick={() => (selectedID = f.id + f.at)}
              >
                <span class="dir {f.dir}">{f.dir === "in" ? "←" : "→"}</span>
                <span class="kind">{f.kind}</span>
                <span class="seq">{f.seq ?? ""}</span>
                <span class="bytes">{fmtBytes(f.bytes)}</span>
                <span class="time">{fmtTime(f.at)}</span>
              </button>
            </li>
          {:else}
            <li class="empty">
              {source.length === 0 ? "No frames captured yet." : "No frames match this filter."}
            </li>
          {/each}
        </ol>

        <div class="detail">
          {#if selected}
            <div class="detail-head">
              <code>{selected.id || "(no id)"}</code>
              <button onclick={copySelected}>copy</button>
            </div>
            <pre>{JSON.stringify(selected.frame, null, 2)}</pre>
          {:else}
            <p class="hint">Select a frame to see its payload.</p>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .inspector {
    position: fixed;
    left: 0;
    bottom: 0;
    z-index: 9998;
    font-family: var(--font-mono, ui-monospace, "SF Mono", Menlo, monospace);
    font-size: 0.7rem;
    color: var(--fg, #e6ecff);
  }

  .handle {
    display: flex;
    align-items: center;
    gap: 0.5em;
    padding: 0.3rem 0.7rem;
    border: 1px solid var(--border-strong, #3a4570);
    border-left: none;
    border-bottom: none;
    border-top-right-radius: var(--radius, 8px);
    background: var(--bg-1, #0b1220);
    color: var(--fg-muted, #9aa5cd);
    font: inherit;
    letter-spacing: 0.1em;
    cursor: pointer;
  }
  .handle:hover {
    color: var(--fg, #e6ecff);
  }
  .count {
    padding: 0 0.4em;
    border-radius: var(--radius-sm, 4px);
    background: var(--surface-raised, #17213a);
    color: var(--accent, #7aa7ff);
  }

  .body {
    width: min(46rem, 100vw);
    height: min(24rem, 50vh);
    display: flex;
    flex-direction: column;
    border: 1px solid var(--border-strong, #3a4570);
    border-left: none;
    border-bottom: none;
    background: var(--bg, #070b14);
    box-shadow: var(--shadow-lg, 0 12px 32px rgba(0, 0, 0, 0.55));
  }

  .toolbar {
    display: flex;
    gap: 0.4rem;
    align-items: center;
    padding: 0.35rem 0.5rem;
    border-bottom: 1px solid var(--border, #273049);
  }
  .toolbar button,
  .toolbar select,
  .detail-head button {
    padding: 0.2rem 0.5rem;
    border: 1px solid var(--border, #273049);
    border-radius: var(--radius-sm, 4px);
    background: var(--surface, #111a2e);
    color: var(--fg-muted, #9aa5cd);
    font: inherit;
    cursor: pointer;
  }
  .toolbar button.active {
    border-color: var(--accent, #7aa7ff);
    color: var(--accent, #7aa7ff);
  }
  .seg {
    display: flex;
    gap: 0.2rem;
  }

  .split {
    flex: 1;
    display: grid;
    grid-template-columns: 1fr 1fr;
    min-height: 0;
  }

  .list {
    margin: 0;
    padding: 0;
    list-style: none;
    overflow-y: auto;
    border-right: 1px solid var(--border, #273049);
  }
  .row {
    display: grid;
    grid-template-columns: 1.2em 6em 3.5em 3em 1fr;
    gap: 0.4em;
    width: 100%;
    padding: 0.15rem 0.5rem;
    border: none;
    background: none;
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }
  .row:hover {
    background: var(--surface, #111a2e);
  }
  .row.sel {
    background: var(--accent-soft, rgba(122, 167, 255, 0.18));
  }
  .dir.in {
    color: var(--mint, #7aff9a);
  }
  .dir.out {
    color: var(--gold, #ffd07a);
  }
  .seq,
  .bytes,
  .time {
    color: var(--fg-dim, #6c7a99);
  }
  .empty,
  .hint {
    padding: 0.6rem;
    color: var(--fg-dim, #6c7a99);
  }

  .detail {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
  .detail-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.5rem;
    padding: 0.3rem 0.5rem;
    border-bottom: 1px solid var(--border, #273049);
  }
  .detail-head code {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--fg-muted, #9aa5cd);
  }
  .detail pre {
    flex: 1;
    margin: 0;
    padding: 0.5rem;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-word;
  }
</style>

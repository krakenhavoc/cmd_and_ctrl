<script lang="ts">
  // CommanderDamageGrid is the always-visible NxN matrix of commander
  // damage, anchored bottom-right of the board. Mirrors the placement
  // convention of HoverZoomOverlay (top-right) and StackOverlay (top-
  // left) so the four corners host the four floating affordances.
  //
  // Layout: rows = recipient (defender), columns = attacker. Reading
  // row-by-row "how much commander damage has player X taken from each
  // opponent" matches how the rule actually applies (lethal at 21 from
  // any single opponent).
  //
  // Click semantics:
  //   - Left-click a cell → +1 commander damage from col → row.
  //   - Shift-click → -1 (clamped at 0).
  //   - Diagonal (self → self) is greyed and disabled.
  //   - Cells reach lethal (≥ 21) flash red.
  //
  // The action is set_commander_damage with set semantics (server
  // replaces, not increments), so we read the current value and send
  // current ± delta. Two stale tabs both clicking +1 land at the same
  // total after one snapshot round-trip.

  import type { ActionPayload, GameView } from "../../protocol";
  import { seatColor } from "../../colors";

  type ActionSender = (type: string, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    view: GameView;
    sendAction: ActionSender;
  }

  const { view, sendAction }: Props = $props();

  // Sort seats by their seat index so the matrix axes match the table
  // order (the layout reorders for the viewer's perspective, but the
  // damage matrix is canonical — every opponent is a column for every
  // recipient regardless of where they sit on screen).
  const seats = $derived([...view.seats].sort((a, b) => a.seat - b.seat));

  function damageFromTo(fromID: string, toID: string): number {
    const recipient = view.seats.find((s) => s.id === toID);
    return recipient?.commander_damage[fromID] ?? 0;
  }

  function bump(fromID: string, toID: string, delta: number): void {
    const current = damageFromTo(fromID, toID);
    const next = Math.max(0, current + delta);
    sendAction("set_commander_damage", { from: fromID, to: toID, amount: next });
  }

  function handleCellClick(fromID: string, toID: string, ev: MouseEvent): void {
    if (fromID === toID) return;
    bump(fromID, toID, ev.shiftKey ? -1 : 1);
  }

  // Collapse / expand toggle so the grid doesn't dominate small screens.
  // Persisted to localStorage so the viewer's choice survives reloads.
  let collapsed = $state(loadCollapsed());
  function loadCollapsed(): boolean {
    if (typeof localStorage === "undefined") return false;
    return localStorage.getItem("cmdctrl.cmdrDamageCollapsed") === "true";
  }
  function toggleCollapsed(): void {
    collapsed = !collapsed;
    if (typeof localStorage !== "undefined") {
      localStorage.setItem("cmdctrl.cmdrDamageCollapsed", String(collapsed));
    }
  }
</script>

<div class="overlay" class:collapsed aria-label="commander damage grid">
  <button
    type="button"
    class="toggle"
    onclick={toggleCollapsed}
    title="toggle commander damage grid"
  >
    <span class="title">cmdr dmg</span>
    <span class="chevron" aria-hidden="true">{collapsed ? "▴" : "▾"}</span>
  </button>
  {#if !collapsed}
    <table class="grid">
      <thead>
        <tr>
          <th class="corner" aria-hidden="true"></th>
          {#each seats as col (col.id)}
            <th
              class="head col"
              style:--seat-color={seatColor(col.seat)}
              title={`from ${col.name}`}
            >
              <span class="dot" aria-hidden="true"></span>
              <span class="name">{col.name.slice(0, 6)}</span>
            </th>
          {/each}
        </tr>
      </thead>
      <tbody>
        {#each seats as row (row.id)}
          <tr>
            <th class="head row" style:--seat-color={seatColor(row.seat)} title={`to ${row.name}`}>
              <span class="dot" aria-hidden="true"></span>
              <span class="name">{row.name.slice(0, 6)}</span>
            </th>
            {#each seats as col (col.id)}
              {@const value = damageFromTo(col.id, row.id)}
              {@const isSelfCell = row.id === col.id}
              {@const lethal = value >= 21}
              <td class="cell" class:self={isSelfCell} class:has={value > 0} class:lethal>
                {#if isSelfCell}
                  <span class="cell-x" aria-hidden="true">·</span>
                {:else}
                  <button
                    type="button"
                    class="cell-btn"
                    onclick={(e) => handleCellClick(col.id, row.id, e)}
                    title={`${col.name} → ${row.name}: ${value} (click +1, shift-click -1)`}
                    aria-label={`${col.name} dealt ${value} commander damage to ${row.name}`}
                  >
                    {value}
                  </button>
                {/if}
              </td>
            {/each}
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  .overlay {
    position: absolute;
    right: 12px;
    bottom: 12px;
    background: linear-gradient(180deg, rgba(19, 26, 44, 0.92) 0%, rgba(8, 12, 24, 0.92) 100%);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    border: 1px solid rgba(122, 167, 255, 0.2);
    border-radius: var(--radius-lg);
    box-shadow:
      0 12px 32px rgba(0, 0, 0, 0.55),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
    padding: 8px 10px 10px;
    z-index: 35;
    pointer-events: auto;
    max-width: 260px;
    font-size: 11px;
    color: var(--fg);
  }
  .toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    width: 100%;
    background: transparent;
    border: 0;
    padding: 0 2px 6px;
    color: inherit;
    font: inherit;
    cursor: pointer;
    box-shadow: none;
  }
  .toggle:hover .chevron {
    color: var(--fg);
  }
  .title {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--fg-dim);
    font-weight: 700;
  }
  .chevron {
    font-size: 10px;
    color: var(--fg-dim);
    transition: color 120ms var(--ease);
  }
  .grid {
    border-collapse: collapse;
    margin: 0;
  }
  th,
  td {
    padding: 0;
    margin: 0;
    border: 1px solid rgba(255, 255, 255, 0.06);
    text-align: center;
    vertical-align: middle;
  }
  .head {
    background: rgba(0, 0, 0, 0.3);
    padding: 3px 5px;
    color: var(--seat-color, #888);
    font-weight: 700;
    font-size: 10px;
    white-space: nowrap;
  }
  .head .dot {
    display: inline-block;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--seat-color, #888);
    margin-right: 3px;
    vertical-align: middle;
  }
  .head .name {
    text-transform: lowercase;
  }
  .corner {
    background: transparent;
    border: 0;
  }
  .cell {
    width: 28px;
    height: 26px;
    background: rgba(0, 0, 0, 0.2);
  }
  .cell-btn {
    width: 100%;
    height: 100%;
    background: transparent;
    border: 0;
    color: inherit;
    font: inherit;
    cursor: pointer;
    font-weight: 800;
    font-variant-numeric: tabular-nums;
    box-shadow: none;
    transition:
      background 120ms var(--ease),
      color 120ms var(--ease);
  }
  .cell-btn:hover {
    background: rgba(255, 208, 122, 0.15);
    color: var(--gold);
  }
  .cell.has .cell-btn {
    color: var(--gold);
  }
  .cell.lethal {
    background: rgba(255, 60, 60, 0.22);
  }
  .cell.lethal .cell-btn {
    color: var(--danger);
    animation: pulse-lethal 1.2s ease-in-out infinite;
  }
  .cell.self {
    background: rgba(0, 0, 0, 0.4);
  }
  .cell-x {
    color: rgba(255, 255, 255, 0.12);
    font-size: 14px;
  }
  @keyframes pulse-lethal {
    0%,
    100% {
      text-shadow: 0 0 0 transparent;
    }
    50% {
      text-shadow: 0 0 8px rgba(255, 122, 122, 0.85);
    }
  }
</style>

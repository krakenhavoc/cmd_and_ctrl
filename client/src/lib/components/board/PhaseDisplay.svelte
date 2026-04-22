<script lang="ts">
  // PhaseDisplay is the "Phases Visualized" widget that lives in the
  // bottom-right of the viewer's own PlayerPanel. It replaces the
  // global turn-bar strip that used to sit above the board.
  //
  // Three stacked rows:
  //   1. Turn number + active player (dot + name)
  //   2. Phase track — 12 step dots in canonical turn order, with the
  //      current step highlighted. Phase boundaries (beginning /
  //      precombat main / combat / postcombat / ending) are separated
  //      by a faint gap so the structure of a turn is visible at a
  //      glance.
  //   3. Priority pills — one chip per seat, glowing for the seat
  //      currently holding priority. Dash when priority is unheld
  //      (Untap / Cleanup per CR 502.4 / 514.3).

  import type { PlayerView, TurnView } from "../../protocol";
  import { seatColor } from "../../colors";
  import { STEP_IDS, STEP_LABELS } from "../../turn";

  interface Props {
    turn: TurnView;
    seats: PlayerView[];
    mulligansOpen: boolean;
  }

  const { turn, seats, mulligansOpen }: Props = $props();

  const activeSeat = $derived(turn.active_seat ?? 0);
  const prioritySeat = $derived(turn.priority_holder ?? 0);
  const priorityHeld = $derived((turn.priority_holder ?? -1) >= 0);
  const activePlayer = $derived(seats[activeSeat]);
  const stepLabel = $derived(STEP_LABELS[turn.step as keyof typeof STEP_LABELS] ?? turn.step);

  // Phase-group boundaries: a faint separator between groups of steps
  // makes the five MTG phases (beginning / precombat main / combat /
  // postcombat / ending) visually distinct without labels.
  const BOUNDARIES: ReadonlySet<number> = new Set([3, 4, 8, 9]);
</script>

<div class="phase-display" aria-label="turn and phase indicator">
  <div class="row summary">
    <span class="turn-no">T{turn.number}</span>
    <span class="active">
      <span class="seat-dot" style="background:{seatColor(activeSeat)}"></span>
      <span class="active-name">{activePlayer?.name ?? `seat ${activeSeat}`}</span>
    </span>
  </div>

  <div class="row track" aria-label="phase track">
    {#each STEP_IDS as id, i (id)}
      {#if BOUNDARIES.has(i)}
        <span class="track-gap" aria-hidden="true"></span>
      {/if}
      <span
        class="step-dot"
        class:current={turn.step === id}
        title={STEP_LABELS[id]}
        aria-current={turn.step === id ? "step" : undefined}
      ></span>
    {/each}
  </div>

  <div class="row step-label">{stepLabel}</div>

  <div class="row pills" role="group" aria-label="priority indicator">
    {#each seats as seat (seat.id)}
      <span
        class="pill"
        class:has-priority={priorityHeld && seat.seat === prioritySeat && !seat.eliminated}
        class:is-active={seat.seat === activeSeat && !seat.eliminated}
        class:eliminated={seat.eliminated}
        style="--seat-color: {seatColor(seat.seat)}"
        title={seat.eliminated
          ? `${seat.name} — eliminated`
          : `${seat.name}${priorityHeld && seat.seat === prioritySeat ? " (priority)" : ""}${seat.seat === activeSeat ? " (active)" : ""}`}
      >
        {seat.name}{seat.eliminated ? " ✕" : ""}
      </span>
    {/each}
    {#if !priorityHeld && !mulligansOpen}
      <span
        class="no-priority"
        title={`no player holds priority during ${stepLabel} (turn-based actions auto-fire)`}
      >
        —
      </span>
    {/if}
  </div>
</div>

<style>
  .phase-display {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 8px 10px;
    min-width: 160px;
    max-width: 220px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.04) 0%, rgba(0, 0, 0, 0.2) 100%), var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--fg-muted);
    font-size: 0.8em;
    box-shadow: var(--shadow-sm);
  }
  .row {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }
  .summary {
    justify-content: space-between;
  }
  .turn-no {
    font-weight: 700;
    color: var(--fg);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
  }
  .active {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    min-width: 0;
  }
  .seat-dot {
    display: inline-block;
    width: 0.65em;
    height: 0.65em;
    border-radius: 50%;
    flex: 0 0 auto;
  }
  .active-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-weight: 600;
    color: var(--fg);
    max-width: 10ch;
  }

  /* Phase track — one dot per step, with a faint gap at each phase
     boundary. Current step pops with a brighter ring and accent glow. */
  .track {
    gap: 3px;
    padding: 2px 0;
  }
  .step-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.1);
    flex: 0 0 auto;
    transition:
      background 160ms var(--ease),
      box-shadow 160ms var(--ease);
  }
  .step-dot.current {
    background: var(--accent);
    border-color: var(--accent-strong);
    box-shadow: 0 0 8px color-mix(in srgb, var(--accent) 65%, transparent);
  }
  .track-gap {
    width: 6px;
    height: 1px;
    flex: 0 0 auto;
  }
  .step-label {
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    font-size: 0.82em;
    color: var(--fg);
  }

  .pills {
    gap: 4px;
  }
  .pill {
    padding: 1px 7px;
    border-radius: 999px;
    font-size: 0.75em;
    border: 1px solid var(--seat-color);
    color: var(--fg);
    background: transparent;
    opacity: 0.5;
    font-weight: 600;
    max-width: 10ch;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    transition:
      opacity 140ms var(--ease),
      box-shadow 140ms var(--ease);
  }
  .pill.is-active {
    opacity: 1;
  }
  .pill.has-priority {
    background: var(--seat-color);
    color: #0c1426;
    font-weight: 700;
    box-shadow:
      0 0 12px var(--seat-color),
      inset 0 1px 0 rgba(255, 255, 255, 0.25);
  }
  .pill.eliminated {
    opacity: 0.3;
    text-decoration: line-through;
    border-style: dashed;
  }
  .no-priority {
    color: var(--fg-dim);
    font-weight: 600;
    cursor: help;
  }
</style>

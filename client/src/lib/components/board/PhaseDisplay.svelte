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
  import { avatarURL } from "../../api";
  import { getAvatarColor } from "../../avatarColor";
  import { STEP_IDS, STEP_LABELS, type StepID } from "../../turn";
  import { canManuallyStop, manualStops, toggleManualStop } from "../../priorityStops";
  import PhaseIcon from "./PhaseIcon.svelte";

  interface Props {
    turn: TurnView;
    seats: PlayerView[];
    mulligansOpen: boolean;
    // Priority controls — only rendered on the viewer's own panel
    // (PlayerPanel gates the mount with `isSelf`), so these are
    // always wired to *viewer* state.
    viewerHasPriority: boolean;
    autopassEnabled: boolean;
    onPassPriority: () => void;
    onToggleAutopass: () => void;
  }

  const {
    turn,
    seats,
    mulligansOpen,
    viewerHasPriority,
    autopassEnabled,
    onPassPriority,
    onToggleAutopass,
  }: Props = $props();

  const activeSeat = $derived(turn.active_seat ?? 0);
  const prioritySeat = $derived(turn.priority_holder ?? 0);
  const priorityHeld = $derived((turn.priority_holder ?? -1) >= 0);
  const activePlayer = $derived(seats[activeSeat]);
  const stepLabel = $derived(STEP_LABELS[turn.step as keyof typeof STEP_LABELS] ?? turn.step);

  // Phase-group boundaries: a faint separator between groups of steps
  // makes the five MTG phases (beginning / precombat main / combat /
  // postcombat / ending) visually distinct without labels.
  const BOUNDARIES: ReadonlySet<number> = new Set([3, 4, 8, 9]);

  // Per-seat accent colour, keyed by player id. Seeded synchronously
  // from seatColor() so first paint has the right shape; avatarColor
  // resolves async and flips this map to the avatar-derived hex when
  // the Discord image has been sampled. Seat-palette players (no
  // Discord identity) never update — getAvatarColor short-circuits
  // with the fallback when url is null.
  const playerColors = $state<Record<string, string>>({});
  $effect(() => {
    for (const s of seats) {
      const url = avatarURL(s.discord_id, s.discord_avatar_hash);
      const fallback = seatColor(s.seat);
      const id = s.id;
      playerColors[id] = getAvatarColor(url, fallback, (c) => {
        playerColors[id] = c;
      });
    }
  });
  const colorFor = (seat: PlayerView): string => playerColors[seat.id] ?? seatColor(seat.seat);
  const activeColor = $derived(activePlayer ? colorFor(activePlayer) : seatColor(activeSeat));

  // S13.6: manual one-time stops. Click a priority-granting icon to
  // pin the cursor there the next time the viewer holds priority.
  // Overrides autoPassPriority + smartAutoPass so the viewer can
  // "fake a game action" — stop to think / bluff / respond even
  // when the engine sees nothing to do. Consumed on step transition
  // by the consumer in Game.svelte.
  const pinned = $derived($manualStops);
  function onIconClick(id: StepID): void {
    if (!canManuallyStop(id)) return;
    toggleManualStop(id);
  }
</script>

<div
  class="phase-display"
  aria-label="turn and phase indicator"
  style:--active-player-color={activeColor}
>
  <div class="row summary">
    <span class="turn-no">T{turn.number}</span>
    <span class="active">
      <span class="seat-dot" style="background:{activeColor}"></span>
      <span class="active-name">{activePlayer?.name ?? `seat ${activeSeat}`}</span>
    </span>
  </div>

  <div class="row track" aria-label="phase track">
    {#each STEP_IDS as id, i (id)}
      {#if BOUNDARIES.has(i)}
        <span class="track-gap" aria-hidden="true"></span>
      {/if}
      {@const clickable = canManuallyStop(id)}
      {@const isPinned = pinned.has(id)}
      <button
        type="button"
        class="step-icon"
        class:current={turn.step === id}
        class:pinned={isPinned}
        class:clickable
        disabled={!clickable}
        title={clickable
          ? isPinned
            ? `${STEP_LABELS[id]} — click to unpin`
            : `${STEP_LABELS[id]} — click to pin a one-time stop`
          : `${STEP_LABELS[id]} — no priority`}
        aria-current={turn.step === id ? "step" : undefined}
        aria-pressed={clickable ? isPinned : undefined}
        onclick={() => onIconClick(id)}
      >
        <PhaseIcon step={id} />
      </button>
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
        style="--seat-color: {colorFor(seat)}"
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

  <div class="row actions" role="group" aria-label="priority controls">
    <button
      type="button"
      class="action next"
      class:viewer-priority={viewerHasPriority}
      disabled={!viewerHasPriority}
      onclick={onPassPriority}
      title={viewerHasPriority ? "pass priority — rotates to next seat" : "you don't hold priority"}
    >
      next
    </button>
    <button
      type="button"
      class="action autopass"
      class:on={autopassEnabled}
      aria-pressed={autopassEnabled}
      onclick={onToggleAutopass}
      title={autopassEnabled
        ? "autopass ON — every time priority lands on you, it passes; click to turn off"
        : "autopass OFF — click to pass every priority window (bypasses stops, smart-skip, and manual pins)"}
    >
      {autopassEnabled ? "autopass ✓" : "autopass"}
    </button>
  </div>
</div>

<style>
  .phase-display {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px 14px;
    min-width: 220px;
    max-width: 300px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.04) 0%, rgba(0, 0, 0, 0.2) 100%), var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--fg-muted);
    font-size: 0.95em;
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

  /* Phase track — one pictogram per step, with a faint gap at each
     phase boundary. Current step pops by switching to the active
     player's avatar-derived colour with a soft glow; inactive steps
     render at low opacity in the chrome colour so the row reads as a
     subtle timeline rather than a noisy icon strip. */
  .track {
    gap: 3px;
    padding: 3px 0;
    flex-wrap: nowrap;
  }
  .step-icon {
    width: 17px;
    height: 17px;
    flex: 0 0 auto;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    position: relative;
    /* Reset button chrome — we're reusing <button> for the
       keyboard / click affordance, not the default look. */
    padding: 0;
    margin: 0;
    border: none;
    background: transparent;
    color: rgba(255, 255, 255, 0.3);
    opacity: 0.85;
    cursor: default;
    transition:
      color 160ms var(--ease),
      opacity 160ms var(--ease),
      filter 160ms var(--ease),
      transform 160ms var(--ease);
  }
  .step-icon.clickable {
    cursor: pointer;
  }
  .step-icon.clickable:hover {
    color: rgba(255, 255, 255, 0.55);
  }
  .step-icon.clickable:focus-visible {
    outline: 1px solid var(--accent);
    outline-offset: 2px;
    border-radius: 3px;
  }
  .step-icon.current {
    color: var(--active-player-color);
    opacity: 1;
    transform: scale(1.18);
    filter: drop-shadow(0 0 6px color-mix(in srgb, var(--active-player-color) 70%, transparent));
  }
  /* Pinned: a small filled dot in the top-right corner. Uses
     --accent so the pin reads as UI state, not game state (the
     active-player colour is already load-bearing on the current
     icon). Pairs with a subtle lift on opacity so pinned steps
     feel distinct from the ambient row even when not current. */
  .step-icon.pinned {
    opacity: 1;
    color: var(--accent);
  }
  .step-icon.pinned.current {
    color: var(--active-player-color);
  }
  .step-icon.pinned::after {
    content: "";
    position: absolute;
    top: -1px;
    right: -1px;
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--accent);
    box-shadow: 0 0 4px color-mix(in srgb, var(--accent) 75%, transparent);
  }
  .track-gap {
    width: 8px;
    height: 1px;
    flex: 0 0 auto;
  }
  .step-label {
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    font-size: 0.92em;
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

  /* Priority controls inside the box — two buttons split the row
     evenly so the panel reads as a self-contained widget. Colours
     echo the Game.svelte toolbar (green = you-have-priority,
     amber = autopass-engaged) so players carry the same visual
     grammar across surfaces. */
  .actions {
    gap: 6px;
    margin-top: 2px;
  }
  .action {
    flex: 1 1 0;
    min-width: 0;
    padding: 4px 8px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: rgba(255, 255, 255, 0.04);
    color: var(--fg);
    font-size: 0.85em;
    font-weight: 600;
    letter-spacing: 0.02em;
    cursor: pointer;
    transition:
      background 140ms var(--ease),
      border-color 140ms var(--ease),
      opacity 140ms var(--ease);
  }
  .action:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.08);
    border-color: color-mix(in srgb, var(--border) 60%, white 40%);
  }
  .action:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .action.next.viewer-priority {
    background: linear-gradient(180deg, #b3e5b3 0%, #7fc87f 100%);
    color: #0a1a0a;
    font-weight: 700;
    border-color: rgba(127, 200, 127, 0.6);
  }
  .action.next.viewer-priority:hover:not(:disabled) {
    background: linear-gradient(180deg, #c4f0c4 0%, #8fd88f 100%);
    border-color: rgba(127, 200, 127, 0.85);
  }
  .action.autopass.on {
    background: linear-gradient(180deg, #f5c76b 0%, #d99a2e 100%);
    color: #1a0e00;
    font-weight: 700;
    border-color: rgba(217, 154, 46, 0.75);
  }
  .action.autopass.on:hover:not(:disabled) {
    background: linear-gradient(180deg, #ffda82 0%, #edaf47 100%);
    border-color: rgba(217, 154, 46, 0.9);
  }
</style>

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
  import { holdPriority, toggleHoldPriority } from "../../holdPriority";
  import { settings } from "../../settings";
  import { effectiveBindings, formatChord, isMacLike } from "../../shortcuts";
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
  const priorityHeld = $derived((turn.priority_holder ?? -1) >= 0);
  const activePlayer = $derived(seats[activeSeat]);
  const stepLabel = $derived(STEP_LABELS[turn.step as keyof typeof STEP_LABELS] ?? turn.step);

  // Key hints for the three priority buttons (ADR 0046). Read from
  // the same binding map the dispatcher uses — imported directly
  // rather than prop-drilled through Board → PlayerPanel, the way
  // this component already imports holdPriority — so a rebound key
  // updates the tooltip and a hint can never advertise a dead key.
  const mac = isMacLike();
  const keys = $derived(effectiveBindings($settings.shortcuts.bindings));
  function keyHint(chord: string): string {
    if (!$settings.shortcuts.enabled || !chord) return "";
    return ` (${formatChord(chord, mac)})`;
  }

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

  <div class="row step-label">
    {stepLabel}
    {#if !priorityHeld && !mulligansOpen}
      <span
        class="no-priority"
        title={`no player holds priority during ${stepLabel} (turn-based actions auto-fire)`}
      >
        · no priority
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
      title={viewerHasPriority
        ? `pass priority — rotates to next seat${keyHint(keys.passPriority)}`
        : "you don't hold priority"}
    >
      next
    </button>
    <!-- #323: the escape hatch for "I DO want to respond to my own
         spell". Lives next to `next` because it has to be clickable
         BEFORE the cast — once the spell is announced the client
         auto-passes on the following snapshot and there is no moment
         left to interrupt. Sticky until clicked off, so `hold ✓` is
         exactly the pre-#323 behaviour: every stack stops. -->
    <button
      type="button"
      class="action hold"
      class:on={$holdPriority}
      aria-pressed={$holdPriority}
      onclick={toggleHoldPriority}
      title={($holdPriority
        ? "hold ON — your own spells and triggers keep the cursor so you can respond to them; click to release"
        : "hold OFF — your own spells and triggers resolve without asking. Click before you cast to keep priority and respond to them") +
        keyHint(keys.holdPriority)}
    >
      {$holdPriority ? "hold ✓" : "hold"}
    </button>
    <button
      type="button"
      class="action autopass"
      class:on={autopassEnabled}
      aria-pressed={autopassEnabled}
      onclick={onToggleAutopass}
      title={(autopassEnabled
        ? "autopass ON — every time priority lands on you, it passes; click to turn off"
        : "autopass OFF — click to pass every priority window (bypasses stops, smart-skip, and manual pins)") +
        keyHint(keys.toggleAutopass)}
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

  .no-priority {
    color: var(--fg-dim);
    font-weight: 500;
    letter-spacing: 0;
    text-transform: none;
    cursor: help;
  }

  /* Priority controls inside the box — two buttons split the row
     evenly so the panel reads as a self-contained widget. The pass
     button is the one gold primary on the table (gold = priority
     everywhere: avatar ring, this button); autopass-engaged is the
     soft gold fill. */
  .actions {
    gap: 6px;
    margin-top: 2px;
  }
  .action {
    /* basis auto (not 0) since #323 added a third button: the row
       sizes each label to its own text and shares the slack, so
       "autopass ✓" can't get squeezed narrower than it reads. The
       row wraps rather than clipping if the panel is at its 220px
       floor. */
    flex: 1 1 auto;
    min-width: 0;
    white-space: nowrap;
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
    background: var(--accent);
    color: var(--accent-fg);
    font-weight: 700;
    border-color: var(--accent-strong);
  }
  .action.next.viewer-priority:hover:not(:disabled) {
    background: var(--accent-strong);
    border-color: var(--accent-strong);
  }
  /* Hold engaged reads in the "user override" colour rather than
     gold — gold is priority everywhere on the table and hold isn't
     priority, it's a standing instruction about it. */
  .action.hold.on {
    background: color-mix(in srgb, var(--magenta) 18%, transparent);
    color: var(--magenta);
    font-weight: 700;
    border-color: color-mix(in srgb, var(--magenta) 55%, transparent);
  }
  .action.hold.on:hover:not(:disabled) {
    background: color-mix(in srgb, var(--magenta) 28%, transparent);
    border-color: var(--magenta);
  }
  .action.autopass.on {
    background: var(--accent-soft);
    color: var(--accent-strong);
    font-weight: 700;
    border-color: rgba(217, 180, 92, 0.55);
  }
  .action.autopass.on:hover:not(:disabled) {
    background: rgba(217, 180, 92, 0.24);
    border-color: var(--accent);
  }
</style>

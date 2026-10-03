<script lang="ts">
  // PhaseDisplay is the header of the action dock (ADR 0111 §1): the
  // turn line, the phase track with its one-time pins, the step label,
  // the turn-rule lines and the ready-actions live region. ActionDock
  // mounts it, and nothing else does. The buttons that used to sit
  // under it (next, hold, bluff, autopass) are the dock's own rows now.
  //
  // Two stacked rows:
  //   1. Turn number + active player (dot + name) … step label
  //   2. Phase track — one icon per step in canonical turn order, with
  //      the current step highlighted. Phase boundaries (beginning /
  //      precombat main / combat / postcombat / ending) are separated
  //      by a faint gap so the structure of a turn is visible at a
  //      glance. A priority-granting step can be clicked to pin a stop.
  //   The step label ends the turn line, with "no priority" during
  //   Untap / Cleanup (CR 502.4 / 514.3).
  //
  // `.turn-no` and `.step-label` are read by the e2e suite with
  // `.first()`, so this must stay the only element on the page with
  // either class.

  import type { PlayerView, TurnView } from "../../protocol";
  import { seatColor } from "../../colors";
  import { avatarURL } from "../../api";
  import { getAvatarColor } from "../../avatarColor";
  import { STEP_IDS, STEP_LABELS, type StepID } from "../../turn";
  import { canManuallyStop, manualStops, toggleManualStop } from "../../priorityStops";
  import PhaseIcon from "./PhaseIcon.svelte";
  import {
    damageCantBePreventedLine,
    damageShieldsLine,
    exileIfCreaturesDieLine,
  } from "../../turnRules";
  import { QUIET_ANNOUNCER, announceArrival, type ReadyAnnouncer } from "../../legalActions";

  interface Props {
    turn: TurnView;
    seats: PlayerView[];
    mulligansOpen: boolean;
    // ADR 0107 §5 (CR 615.12): the sources of the live "damage can't
    // be prevented this turn" grants (GameView.damage_cant_be_prevented).
    damageCantBePrevented?: string[];
    exileIfCreaturesDie?: string[];
    // ADR 0108 §7: the live source shields (GameView.damage_shields).
    damageShields?: string[];
    // ADR 0105 §7 (sub-PR 6): how many of the viewer's cards have a
    // highlighted action on this frame (legalActions.ts
    // actionableCount over the HIGHLIGHT lookup, so 0 while highlights
    // are off or autopass is about to pass). The live region below
    // says it once, when a decision arrives.
    readyActions?: number;
    // ADR 0111 §8: at phone width the track is folded away until the
    // dock's ▴ opens it. Ignored above 600px, where it always shows.
    trackOpen?: boolean;
  }

  const {
    turn,
    seats,
    mulligansOpen,
    damageCantBePrevented = [],
    exileIfCreaturesDie = [],
    damageShields = [],
    readyActions = 0,
    trackOpen = false,
  }: Props = $props();

  // ADR 0105 §7: "N actions available", once per arrival. The
  // announcer is plain state on purpose: it is read and written only
  // by the effect, which depends on readyActions alone, so a frame
  // that repeats the same decision changes nothing the region shows.
  let announcer: ReadyAnnouncer = QUIET_ANNOUNCER;
  let readyLine = $state("");
  $effect(() => {
    announcer = announceArrival(announcer, readyActions);
    readyLine = announcer.text;
  });

  const unpreventableLine = $derived(damageCantBePreventedLine(damageCantBePrevented));
  const exileOnDeathLine = $derived(exileIfCreaturesDieLine(exileIfCreaturesDie));
  const shieldsLine = $derived(damageShieldsLine(damageShields));

  const activeSeat = $derived(turn.active_seat ?? 0);
  const priorityHeld = $derived((turn.priority_holder ?? -1) >= 0);
  const activePlayer = $derived(seats[activeSeat]);
  const stepLabel = $derived(STEP_LABELS[turn.step as keyof typeof STEP_LABELS] ?? turn.step);
  // ADR 0059 Decision 11 (#753): the next queued extra turn, if any.
  // "T{n}" stays the round (owner decision 1); an extra turn is marked
  // instead, and the queue shows who takes the next one.
  const nextExtra = $derived.by(() => {
    const seat = turn.extra_turns?.[0];
    if (seat === undefined) return null;
    return seats[seat]?.name ?? `seat ${seat}`;
  });

  // Phase-group boundaries: a faint separator between groups of steps
  // makes the five MTG phases (beginning / precombat main / combat /
  // postcombat / ending) visually distinct without labels.
  // Indices into STEP_IDS: before precombat_main, before begin_combat,
  // before end_combat and before postcombat_main. They move whenever a
  // step is added to the list — first_strike_damage (#717) pushed the
  // last two along by one.
  const BOUNDARIES: ReadonlySet<number> = new Set([3, 4, 9, 10]);

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
  // Overrides autoPassPriority + smartAutoPass + the autopass
  // toggle (#526) so the viewer can "fake a game action" — stop to
  // think / bluff / respond even when the engine sees nothing to do.
  // Consumed on step transition by the consumer in Game.svelte.
  const pinned = $derived($manualStops);

  function onIconClick(id: StepID): void {
    if (!canManuallyStop(id)) return;
    toggleManualStop(id);
  }
</script>

<div
  class="phase-display"
  class:track-open={trackOpen}
  aria-label="turn and phase indicator"
  style:--active-player-color={activeColor}
>
  <div class="row summary">
    <span class="turn-no">T{turn.number}</span>
    <span class="active">
      <span class="seat-dot" style="background:{activeColor}"></span>
      <span class="active-name">{activePlayer?.name ?? `seat ${activeSeat}`}</span>
    </span>
    {#if turn.extra}
      <span class="extra-turn" title="This turn was created by an effect">Extra turn</span>
    {/if}
    {#if nextExtra}
      <span
        class="next-extra"
        title="Queued extra turns are taken before normal turn order resumes"
      >
        Next: {nextExtra} (extra)
      </span>
    {/if}
    <!-- The step label shares the turn line, which keeps the dock one
         row shorter (and the self panel's rows one row taller). -->
    <span class="step-label">
      {stepLabel}
      {#if !priorityHeld && !mulligansOpen}
        <span
          class="no-priority"
          title={`no player holds priority during ${stepLabel} (turn-based actions auto-fire)`}
        >
          · no priority
        </span>
      {/if}
    </span>
  </div>

  <div class="row track" id="dock-phase-track" aria-label="phase track">
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

  <!-- ADR 0105 §7 (#1789): the spoken half of the ready highlights.
       Always mounted, so a screen reader is already listening when the
       line arrives; empty between decisions, so the same line is news
       again the next time priority comes round. -->
  <span class="sr-only" role="status" aria-live="polite" aria-atomic="true" data-ready-announcer
    >{readyLine}</span
  >

  <!-- ADR 0107 §5: "damage can't be prevented this turn" changes what
       every Fog and shield at the table does, so the turn says so. -->
  {#if unpreventableLine}
    <div class="row turn-rule" role="status">{unpreventableLine}</div>
  {/if}
  {#if exileOnDeathLine}
    <div class="row turn-rule" role="status">{exileOnDeathLine}</div>
  {/if}
  {#if shieldsLine}
    <div class="row turn-rule" role="status">{shieldsLine}</div>
  {/if}
</div>

<style>
  /* The dock draws the box; the header is its top section. */
  .phase-display {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
    color: var(--fg-muted);
    font-size: 0.95em;
  }
  .row {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }
  .summary {
    justify-content: flex-start;
  }
  .extra-turn,
  .next-extra {
    font-size: 0.75em;
    padding: 0 0.4em;
    border-radius: 999px;
    border: 1px solid var(--active-player-color, currentColor);
    white-space: nowrap;
  }
  .next-extra {
    opacity: 0.75;
    border-style: dashed;
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
    max-width: 14ch;
  }

  /* Phase track — one pictogram per step, with a faint gap at each
     phase boundary. Current step pops by switching to the active
     player's avatar-derived colour with a soft glow; inactive steps
     render at low opacity in the chrome colour so the row reads as a
     subtle timeline rather than a noisy icon strip. */
  .track {
    gap: 3px;
    padding: 2px 0;
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
    margin-left: auto;
    white-space: nowrap;
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
  .turn-rule {
    font-size: 0.72rem;
    color: var(--danger);
    opacity: 0.9;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  /* ADR 0111 §8: at phone width the header folds to one line
     (T7 · Alice · DECLARE BLOCKERS) and the dock's ▴ opens the track. */
  @media (max-width: 599px) {
    .phase-display:not(.track-open) .track {
      display: none;
    }
  }
</style>

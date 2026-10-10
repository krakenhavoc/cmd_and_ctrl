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
  //      the current step highlighted, in five groups: CR 500.1's
  //      phases (beginning / precombat main / combat / postcombat main
  //      / ending), each `role="group"` with a tiny caption under it
  //      (BEGIN, MAIN 1, COMBAT, MAIN 2, END) and a hairline between
  //      groups (#2214). A priority-granting step can be clicked to pin
  //      a stop; its button name (its title) is unchanged by #2214.
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
  import { STEP_LABELS, type StepID } from "../../turn";
  import { canManuallyStop, manualStops, toggleManualStop } from "../../priorityStops";
  import PhaseIcon from "./PhaseIcon.svelte";
  import {
    damageCantBePreventedLine,
    damageMultiplierLines,
    damageRedirectionLines,
    damageShieldsLine,
    exileIfCreaturesDieLine,
  } from "../../turnRules";
  import { QUIET_ANNOUNCER, announceArrival, type ReadyAnnouncer } from "../../legalActions";
  import { L } from "../../labels";
  import { dayNightChip } from "../../dayNight";

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
    // ADR 0108 §3: the live damage multipliers' banner lines
    // (GameView.damage_multipliers).
    damageMultipliers?: string[];
    // ADR 0108 §9: the live damage redirections' banner lines
    // (GameView.damage_redirections).
    damageRedirections?: string[];
    // ADR 0132 (CR 731): GameView.day_night, absent while the game has
    // neither designation.
    dayNight?: string;
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
    damageMultipliers = [],
    damageRedirections = [],
    dayNight = undefined,
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
  const multiplierLines = $derived(damageMultiplierLines(damageMultipliers));
  const redirectionLines = $derived(damageRedirectionLines(damageRedirections));
  const dayNightBadge = $derived(dayNightChip(dayNight));

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

  // #2214: the track is CR 500.1's five phases, each a labelled group
  // with a tiny caption under its steps. `name` is the group's
  // accessible name (new in #2214; the step buttons' names did not
  // change), `caption` the visible one. Every step of STEP_IDS sits in
  // exactly one group, in order: phaseDisplay.track.render.test.ts
  // fails when a step is added to turn.ts and not here.
  const PHASES: ReadonlyArray<{
    key: string;
    name: string;
    caption: string;
    steps: readonly StepID[];
  }> = [
    {
      key: "beginning",
      name: "beginning phase",
      caption: "Begin",
      steps: ["untap", "upkeep", "draw"],
    },
    {
      key: "precombat_main",
      name: "precombat main phase",
      caption: "Main 1",
      steps: ["precombat_main"],
    },
    {
      key: "combat",
      name: "combat phase",
      caption: "Combat",
      steps: [
        "begin_combat",
        "declare_attackers",
        "declare_blockers",
        "first_strike_damage",
        "combat_damage",
        "end_combat",
      ],
    },
    {
      key: "postcombat_main",
      name: "postcombat main phase",
      caption: "Main 2",
      steps: ["postcombat_main"],
    },
    { key: "ending", name: "ending phase", caption: "End", steps: ["end", "cleanup"] },
  ];

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
  // Overrides the pass mode, the stops grid and the autopass
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
  aria-label={L.turnPhase}
  style:--active-player-color={activeColor}
>
  <div class="row summary">
    <span class="turn-no">T{turn.number}</span>
    <span class="active">
      <span class="seat-dot" style="background:{activeColor}"></span>
      <span class="active-name">{activePlayer?.name ?? `seat ${activeSeat}`}</span>
    </span>
    {#if dayNightBadge}
      <span class="day-night" data-day-night={dayNight} title={dayNightBadge.title}>
        <span aria-hidden="true">{dayNightBadge.glyph}</span>
        {dayNightBadge.label}
      </span>
    {/if}
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
    {#each PHASES as phase, gi (phase.key)}
      {#if gi > 0}
        <span class="track-gap" aria-hidden="true"></span>
      {/if}
      <div
        class="phase-group"
        class:current-phase={phase.steps.includes(turn.step as StepID)}
        role="group"
        aria-label={phase.name}
        data-phase={phase.key}
      >
        <div class="phase-steps">
          {#each phase.steps as id (id)}
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
              data-step={id}
              onclick={() => onIconClick(id)}
            >
              <PhaseIcon step={id} />
            </button>
          {/each}
        </div>
        <!-- The group's name says it to a screen reader; this is the
             sighted half. -->
        <span class="phase-caption" aria-hidden="true">{phase.caption}</span>
      </div>
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
  {#each multiplierLines as line, i (i)}
    <div class="row turn-rule" role="status">{line}</div>
  {/each}
  {#each redirectionLines as line, i (i)}
    <div class="row turn-rule" role="status">{line}</div>
  {/each}
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
  .day-night,
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
    gap: 0;
    padding: 2px 0 0;
    flex-wrap: nowrap;
    align-items: stretch;
    min-width: 0;
  }
  /* #2214: one group per phase, its steps over a tiny caption. A main
     phase is one step under a wider caption, so the group is as wide
     as the wider of the two. */
  .phase-group {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    flex: 0 0 auto;
  }
  .phase-steps {
    display: flex;
    align-items: center;
    gap: 2px;
  }
  .phase-caption {
    font-size: 0.5rem;
    line-height: 1;
    font-weight: 700;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    white-space: nowrap;
    color: var(--fg-dim);
    transition: color 160ms var(--ease);
  }
  .phase-group.current-phase .phase-caption {
    color: var(--active-player-color, var(--fg));
  }
  .step-icon {
    width: 16px;
    height: 16px;
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
    /* From --fg, not a fixed white, so the light theme draws the
       idle steps too (#2214). */
    color: color-mix(in srgb, var(--fg) 40%, transparent);
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
    color: color-mix(in srgb, var(--fg) 62%, transparent);
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
  /* Between two phases: a hairline beside the icons, with the spare
     width of a wide dock shared out between the four of them. */
  .track-gap {
    flex: 1 1 0;
    min-width: 3px;
    max-width: 26px;
    align-self: stretch;
    --gap-line: color-mix(in srgb, var(--fg) 16%, transparent);
    background: linear-gradient(var(--gap-line), var(--gap-line)) center 1px / 1px 14px no-repeat;
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
  /* The narrowest phones (a 320px screen leaves the track ~250px):
     the icons drop to 14px so the five groups still fit on one line. */
  @media (max-width: 359px) {
    .phase-steps {
      gap: 1px;
    }
    .step-icon {
      width: 14px;
      height: 14px;
    }
  }
</style>

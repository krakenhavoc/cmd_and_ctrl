<script lang="ts">
  // ActionDock — ADR 0111 (S56 PR 2). Every button a player presses to
  // move the game on, in one place: the screen's bottom-right corner
  // (owner decision 1). Game.svelte mounts it beside the Board, inside
  // `.play-area`, for a seated viewer on the live game only: never for a
  // spectator, never while the dev replay scrubber shows a past frame,
  // never before the first snapshot.
  //
  // Top to bottom:
  //   1. Header — PhaseDisplay: turn, phase track and pins, step label.
  //   2. Status line — the running bluff, the CR 732 loop notice, or
  //      what `next` will do. Always one row tall, so the dock does not
  //      change height (and move the self panel) on every priority pass.
  //   3. Toggles row — hold, autopass, bluff (`group "priority controls"`).
  //   4. Action bar — the secondary on the left, ONE primary on the
  //      right. Until PR 3 brings requests in, that is always Pass turn
  //      and `next`.
  //
  // `next` and Pass turn are rendered disabled, never hidden, when the
  // viewer cannot press them: the e2e suite reads `isEnabled()` on both
  // and the tutorial anchors to them.
  //
  // Keys (ADR 0111 §1, ADR 0047): Space is still "pass priority" through
  // the shortcut layer and still defers to a focused button. Enter on a
  // focused action-bar button presses that button and nothing else (it
  // is stopped here, so the targeting walk's window-level Enter does not
  // also fire). Enter with focus on the board never presses `next`, and
  // Escape never touches it. Focus does not move on an ordinary priority
  // frame: that would steal it on every pass.

  import type { GameView } from "../../protocol";
  import PhaseDisplay from "./PhaseDisplay.svelte";
  import BluffChip from "./BluffChip.svelte";
  import { holdPriority, toggleHoldPriority } from "../../holdPriority";
  import { bluffStatus, bluffStatusText } from "../../bluff";
  import { settings } from "../../settings";
  import { ariaKeyshortcuts, effectiveBindings, formatChord, isMacLike } from "../../shortcuts";
  import { passHint } from "../../dockHint";

  interface Props {
    view: GameView;
    viewerHasPriority: boolean;
    viewerIsActive: boolean;
    // The active player's name, for Pass turn's disabled tooltip.
    activePlayerName?: string;
    autopassEnabled: boolean;
    // #628 (CR 732): the loop-breaker line, or "" when the table is
    // quiet. While it is non-empty NOTHING passes automatically — the
    // autopass toggle is suspended rather than switched off.
    loopNotice?: string;
    // ADR 0105 §7: the header's "N actions available" count.
    readyActions?: number;
    onPassPriority: () => void;
    onPassTurn: () => void;
    onToggleAutopass: () => void;
    // The dock's live size, for --dock-w / --dock-h (ADR 0111 §4).
    onSize?: (width: number, height: number) => void;
  }

  const {
    view,
    viewerHasPriority,
    viewerIsActive,
    activePlayerName = "another seat",
    autopassEnabled,
    loopNotice = "",
    readyActions = 0,
    onPassPriority,
    onPassTurn,
    onToggleAutopass,
    onSize,
  }: Props = $props();

  // ---- keys --------------------------------------------------------
  // Read from the same binding map the dispatcher uses, so a rebound
  // key updates the tooltip, the cap and aria-keyshortcuts together,
  // and none of them can advertise a dead key.
  const mac = isMacLike();
  const keys = $derived(effectiveBindings($settings.shortcuts.bindings));
  const keysOn = $derived($settings.shortcuts.enabled);
  function keyHint(chord: string): string {
    if (!keysOn || !chord) return "";
    return ` (${formatChord(chord, mac)})`;
  }
  function ariaKeys(chord: string): string | undefined {
    if (!keysOn || !chord) return undefined;
    return ariaKeyshortcuts(chord) || undefined;
  }
  const nextCap = $derived(keysOn && keys.passPriority ? formatChord(keys.passPriority, mac) : "");

  // Enter on a focused action-bar button presses it, once, and the
  // keypress ends there. A disabled button never gets focus, but the
  // guard costs nothing.
  function enterPresses(e: KeyboardEvent, enabled: boolean, press: () => void): void {
    if (e.key !== "Enter" || e.shiftKey || e.ctrlKey || e.altKey || e.metaKey) return;
    e.preventDefault();
    e.stopPropagation();
    if (enabled) press();
  }

  // ---- status line ---------------------------------------------------
  const autopassPaused = $derived(loopNotice !== "");

  // #1307: the running bluff's line. The countdown ticks twice a
  // second while a timed bluff runs.
  let now = $state(Date.now());
  $effect(() => {
    const s = $bluffStatus;
    if (!s || s.manual) return;
    now = Date.now();
    const id = setInterval(() => (now = Date.now()), 500);
    return () => clearInterval(id);
  });
  const bluffLine = $derived(bluffStatusText($bluffStatus, now));
  const hint = $derived(passHint(view, viewerHasPriority));

  // ---- phone header ----------------------------------------------------
  // ADR 0111 §8: below 600px the header is one line and ▴ opens the
  // phase track (and its pins). Above it the toggle is not drawn.
  let trackOpen = $state(false);

  // ---- size ------------------------------------------------------------
  let root: HTMLElement | null = $state(null);
  $effect(() => {
    const el = root;
    if (!el || !onSize) return;
    const report = (): void => onSize(el.offsetWidth, el.offsetHeight);
    report();
    if (typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(report);
    ro.observe(el);
    return () => ro.disconnect();
  });
</script>

<section class="action-dock" aria-label="actions" bind:this={root}>
  <div class="dock-head">
    <PhaseDisplay
      turn={view.turn}
      seats={view.seats}
      mulligansOpen={view.mulligans_open === true}
      damageCantBePrevented={view.damage_cant_be_prevented ?? []}
      exileIfCreaturesDie={view.exile_if_creatures_die ?? []}
      {readyActions}
      {trackOpen}
    />
    <button
      type="button"
      class="track-toggle"
      aria-expanded={trackOpen}
      aria-controls="dock-phase-track"
      aria-label={trackOpen ? "hide the phase track" : "show the phase track"}
      title={trackOpen ? "hide the phase track" : "show the phase track and its stops"}
      onclick={() => (trackOpen = !trackOpen)}
    >
      {trackOpen ? "▾" : "▴"}
    </button>
  </div>

  <!-- One row, always present, so a hint coming and going does not
       resize the dock and reflow the self panel under the pointer. -->
  <div class="dock-status">
    {#if autopassPaused}
      <!-- #628 (CR 732): the loop breaker. "Why has autopass stopped
           working" is the only question it answers, so it sits right
           above the toggle it is talking about. -->
      <div class="loop-notice" role="status" aria-live="polite">
        <span class="loop-label">loop detected</span>
        <span class="loop-text">{loopNotice} Autopass paused.</span>
      </div>
    {:else if bluffLine}
      <!-- Only the viewer sees this line. -->
      <span class="bluff-status" role="status">{bluffLine}</span>
    {:else if hint}
      <span class="pass-hint">{hint}</span>
    {/if}
  </div>

  <div class="dock-toggles" role="group" aria-label="priority controls">
    <!-- #323: the escape hatch for "I DO want to respond to my own
         spell". It has to be clickable BEFORE the cast — once the spell
         is announced the client auto-passes on the following snapshot.
         Sticky until clicked off. -->
    <button
      type="button"
      class="action hold"
      class:on={$holdPriority}
      aria-pressed={$holdPriority}
      aria-keyshortcuts={ariaKeys(keys.holdPriority)}
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
      class:on={autopassEnabled && !autopassPaused}
      class:paused={autopassPaused}
      aria-pressed={autopassEnabled}
      aria-keyshortcuts={ariaKeys(keys.toggleAutopass)}
      onclick={onToggleAutopass}
      title={(autopassPaused
        ? "autopass PAUSED — a loop is resolving (CR 732). Use next to step through it; passing resumes on the next real play"
        : autopassEnabled
          ? "autopass ON — every time priority lands on you, it passes; click to turn off, or pin a phase icon to stop at just that step"
          : "autopass OFF — click to pass every priority window (bypasses stops and smart-skip; a pinned phase icon still stops you)") +
        keyHint(keys.toggleAutopass)}
    >
      {autopassPaused ? "autopass ⏸" : autopassEnabled ? "autopass ✓" : "autopass"}
    </button>
    <!-- ADR 0111 §5: always shown, set up or not. -->
    <BluffChip keyHint={keyHint(keys.toggleBluff)} keyShortcuts={ariaKeys(keys.toggleBluff)} />
  </div>

  <!-- The action bar. Its slots never move: secondary on the left, the
       one primary in the very corner, so the pointer's resting place is
       always the button that moves the game on. -->
  <div class="dock-bar">
    <button
      type="button"
      class="dock-btn secondary pass-turn"
      disabled={!viewerIsActive}
      aria-keyshortcuts={ariaKeys(keys.passTurn)}
      onclick={onPassTurn}
      onkeydown={(e) => enterPresses(e, viewerIsActive, onPassTurn)}
      title={viewerIsActive
        ? `skip the rest of your turn${keyHint(keys.passTurn)}`
        : `${activePlayerName} is the active player`}
    >
      Pass turn
    </button>
    <button
      type="button"
      class="dock-btn primary next"
      class:viewer-priority={viewerHasPriority}
      disabled={!viewerHasPriority}
      aria-keyshortcuts={ariaKeys(keys.passPriority)}
      onclick={onPassPriority}
      onkeydown={(e) => enterPresses(e, viewerHasPriority, onPassPriority)}
      title={viewerHasPriority
        ? `pass priority — rotates to next seat${keyHint(keys.passPriority)}`
        : "you don't hold priority"}
    >
      next{#if nextCap}<kbd class="cap" aria-hidden="true">{nextCap}</kbd>{/if}
    </button>
  </div>
</section>

<style>
  /* ADR 0111 §1 and §4: the screen's bottom-right corner. Anchored to
     .play-area (Game.svelte), inset by --dock-inset so its edges line up
     with the self panel's content box: the panel keeps an empty cell
     exactly this size under its rail (PlayerPanel's .dock-spacer).
     z 55: above the log drawer (30), the strip and the command bar
     (40) and the vote panel (50); below the card-local menus (60+),
     the full-screen modals (200) and the hover zoom (300). */
  .action-dock {
    position: absolute;
    right: var(--dock-inset, 15px);
    bottom: var(--dock-inset, 15px);
    z-index: 55;
    width: clamp(300px, 26vw, 380px);
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 5px;
    padding: 8px 12px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.04) 0%, rgba(0, 0, 0, 0.2) 100%), var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow-sm);
    color: var(--fg-muted);
    font-size: 0.95em;
  }
  .dock-head {
    display: flex;
    align-items: flex-start;
    gap: 6px;
    min-width: 0;
  }
  .dock-head > :global(.phase-display) {
    flex: 1 1 auto;
  }
  /* Phone only (below). */
  .track-toggle {
    display: none;
  }

  .dock-status {
    min-height: 1.25em;
    font-size: 0.72rem;
    line-height: 1.25;
    min-width: 0;
  }
  .pass-hint,
  .bluff-status {
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .pass-hint {
    color: var(--fg-dim);
  }
  .bluff-status {
    color: var(--magenta);
    opacity: 0.85;
  }
  .loop-notice {
    display: flex;
    align-items: flex-start;
    gap: 6px;
    padding: 6px 8px;
    border: 1px solid rgba(255, 107, 107, 0.45);
    border-radius: 6px;
    background: rgba(255, 107, 107, 0.1);
    line-height: 1.35;
  }
  .loop-label {
    flex: none;
    font-size: 0.66rem;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    font-weight: 700;
    color: var(--danger);
  }
  .loop-text {
    opacity: 0.9;
  }

  /* Toggles: small and quiet, always in the same order. One row that
     never wraps, so the dock keeps its height; a label shrinks before
     the row grows. */
  .dock-toggles {
    display: flex;
    gap: 6px;
    min-width: 0;
  }
  .action {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
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
  /* #628: suspended, not switched off — a distinct look from both
     `on` (gold) and `off` (flat). */
  .action.autopass.paused {
    background: rgba(255, 107, 107, 0.14);
    border-color: rgba(255, 107, 107, 0.5);
    color: var(--danger);
    font-weight: 700;
  }

  /* The action bar: secondary left, primary right. The primary is the
     one gold button on the table (gold = priority everywhere). */
  .dock-bar {
    display: flex;
    gap: 8px;
    padding-top: 5px;
    border-top: 1px solid var(--border);
  }
  .dock-btn {
    min-height: 32px;
    padding: 0 12px;
    border-radius: 7px;
    border: 1px solid var(--border);
    background: rgba(255, 255, 255, 0.04);
    color: var(--fg);
    font-size: 0.9em;
    font-weight: 600;
    white-space: nowrap;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
  }
  .dock-btn.secondary {
    flex: 0 1 auto;
  }
  .dock-btn.primary {
    flex: 1 1 auto;
    margin-left: auto;
    max-width: 60%;
    letter-spacing: 0.04em;
  }
  .dock-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.08);
  }
  .dock-btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .dock-btn:focus-visible,
  .action:focus-visible,
  .track-toggle:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .dock-btn.next.viewer-priority {
    background: var(--accent);
    color: var(--accent-fg);
    font-weight: 700;
    border-color: var(--accent-strong);
  }
  .dock-btn.next.viewer-priority:hover:not(:disabled) {
    background: var(--accent-strong);
    border-color: var(--accent-strong);
  }
  .cap {
    font-family: var(--font-mono);
    font-size: 0.7em;
    font-weight: 600;
    padding: 1px 5px;
    border-radius: 4px;
    border: 1px solid currentColor;
    opacity: 0.6;
    letter-spacing: 0;
  }

  /* ADR 0111 §8: a phone gets a full-width bar on the bottom of the
     play area, inside the 16px gutter (the section's padding plus 6px).
     .play-area pads its bottom by --dock-h, so the board ends above the
     bar instead of under it. */
  @media (max-width: 599px) {
    .action-dock {
      left: 6px;
      right: 6px;
      bottom: 0;
      width: auto;
      gap: 6px;
      padding: 8px 10px;
    }
    .track-toggle {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      flex: 0 0 auto;
      width: 28px;
      height: 24px;
      padding: 0;
      border-radius: 6px;
      border: 1px solid var(--border);
      background: transparent;
      color: var(--fg-muted);
      cursor: pointer;
    }
    .dock-bar {
      gap: 6px;
    }
    .dock-btn {
      min-height: 44px;
    }
    /* A phone has no Space bar to advertise. */
    .cap {
      display: none;
    }
    .dock-btn.secondary,
    .dock-btn.primary {
      flex: 1 1 50%;
      max-width: none;
      margin-left: 0;
    }
  }
</style>

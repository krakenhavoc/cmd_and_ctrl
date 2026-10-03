<script lang="ts">
  // TutorialCoach — the tutorial on the practice table (ADR 0076 §2.3,
  // §2.4; #1079). Game.svelte mounts it inside `.play-area` while this
  // tab's practice game (lib/practiceTable.ts) is on screen.
  //
  // It owns three things:
  //
  //   1. The step machine (lib/tutorial.ts), fed the game view and the
  //      tutorialBus events.
  //   2. The spotlight. While a step asks for something, its anchor is
  //      measured every POLL_MS and the scrim's hole follows it (a hand
  //      that lifts on hover, a row that re-centres). An anchor that is
  //      missing for ANCHOR_GRACE_MS (it gets a moment to render) makes
  //      the step advance itself and log. A tutorial never waits on an
  //      element that is not there. The same poll times a hover step's
  //      rest (steps 2 and 4, "hover ≥ 600ms"), and reads an anchor
  //      that comes off the board (step 7's card) afresh each tick.
  //   3. Placement. The card docks bottom-left (§2.3, amended
  //      2026-10-02). It publishes its live size through `onSize`, and
  //      Game.svelte turns that into --coach-w / --coach-h: on a desktop
  //      the viewer's own panel keeps an empty cell that size at the left
  //      of its bottom row (PlayerPanel's .coach-spacer, the twin of the
  //      dock's .dock-spacer), so the hand centres in what is left and
  //      nothing sits under the card. On a phone the card is a strip
  //      stacked on the dock bar, and the play area's bottom padding
  //      grows by its height.
  //
  // Skip tutorial and Finish hide the card and leave the player at the
  // practice table; the table is still a game. Replay opens a fresh one.

  import { onDestroy, untrack } from "svelte";
  import type { GameView } from "../../protocol";
  import { subscribe as onTutorialEvent } from "../../tutorialBus";
  import {
    ANCHOR_GRACE_MS,
    POLL_MS,
    TOUCH_HOVER_STEP_MS,
    TUTORIAL_STEP_COUNT,
    copyText,
    createTutorialRun,
    spotAnchors,
    spotlit,
    statusText,
    type CopyContext,
    type TutorialStep,
  } from "../../tutorial";
  import { anchorRect, resolveAnchor, sameRect, type AnchorRect } from "../../tutorialAnchor";
  import { TUTORIAL_STEPS } from "../../tutorialSteps";
  import { settings } from "../../settings";
  import { effectiveBindings, formatChord, isMacLike } from "../../shortcuts";
  import { navigate } from "../../router";
  import CoachCard from "./CoachCard.svelte";
  import TutorialScrim from "./TutorialScrim.svelte";

  interface Props {
    view: GameView | null;
    viewerID: string | null;
    /** The script; the real one unless a test hands it another. */
    steps?: TutorialStep[];
    /** The card's live size; (0, 0) once it is hidden. */
    onSize?: (width: number, height: number) => void;
    /** Replay. Defaults to opening a fresh practice table. */
    onReplay?: () => void;
    pollMs?: number;
    anchorGraceMs?: number;
    /** Where self-advances are logged. Defaults to console.warn. */
    log?: (msg: string) => void;
    /**
     * Whether this device has a pointer that hovers. Defaults to the
     * `(hover: hover)` media query; true where there is no matchMedia.
     */
    canHover?: boolean;
    /** Whether the pointer is over an element. Defaults to `:hover`. */
    isHovering?: (el: Element) => boolean;
    /** A hover step on a device that cannot hover gives up after this. */
    touchHoverStepMs?: number;
    /** The dock's autopass toggle (Game.svelte's session state), for step 9. */
    autopass?: boolean;
  }

  const {
    view,
    viewerID,
    steps = TUTORIAL_STEPS,
    onSize,
    onReplay = () => navigate("#/practice"),
    pollMs = POLL_MS,
    anchorGraceMs = ANCHOR_GRACE_MS,
    log,
    canHover = typeof matchMedia === "function" ? matchMedia("(hover: hover)").matches : true,
    isHovering = (el: Element) => el.matches(":hover"),
    touchHoverStepMs = TOUCH_HOVER_STEP_MS,
    autopass = false,
  }: Props = $props();

  const run = untrack(() =>
    createTutorialRun(steps, { viewerID, view, log, client: { autopass } }),
  );
  let snap = $state(run.current());
  const unsubRun = run.subscribe((s) => (snap = s));

  // The game moved: check the step.
  $effect(() => {
    run.observe(view);
  });
  // The autopass toggle never reaches the wire; step 9 reads it here.
  $effect(() => {
    run.observeClient({ autopass });
  });
  // A hover or a card-local menu: the three things the snapshot cannot
  // see (ADR 0076 §2.5). On a device with no hover, a hover step's own
  // event is the whole gesture (a touch on the hand or a pile); with a
  // mouse the step waits for the pointer to rest (below).
  const unsubBus = onTutorialEvent((e) => {
    const step = run.current().step;
    if (!canHover && step.hover?.event === e) run.hovered(step.id);
    run.observe(view, e);
  });

  // A hover step on a touch screen has no hover to teach and may have no
  // event to wait for (one Forest is not a pile), so it gives up after a
  // read rather than wait on something that cannot happen (§2.4).
  // Derived, so a republish of the same step (a hint, a detour) does
  // not restart the timer.
  const currentStep = $derived(snap.step);
  const shown = $derived(snap.visible);
  $effect(() => {
    const step = currentStep;
    if (canHover || !step.hover || !shown) return;
    const t = setTimeout(
      () => run.advance(step.id, "cannot be hovered on this device"),
      touchHoverStepMs,
    );
    return () => clearTimeout(t);
  });

  onDestroy(() => {
    unsubBus();
    unsubRun();
    run.destroy();
  });

  // ---- The spotlight ----
  let rect: AnchorRect | null = $state(null);
  $effect(() => {
    const step = snap.step;
    const detour = snap.detour;
    const lit = snap.visible && spotlit(snap.coach);
    // A step with no anchor at all (none in the script) points at
    // nothing. One whose anchor is read off the board is measured every
    // tick, and naming nothing then is a missing anchor.
    if (!lit || (step.anchor === undefined && !detour?.anchor)) {
      rect = null;
      return;
    }
    // A hover step completes once the pointer has rested on its anchor
    // for step.hover.ms (§2.1). Measured here, on the poll that already
    // reads the anchor: the bus says when the pointer arrives, but not
    // that it stayed, and a pointer crossing the hand on its way to the
    // dock is not reading it. Not while a detour points elsewhere.
    const dwell = canHover && !detour ? step.hover : undefined;
    let hoverSince: number | null = null;
    let missingSince: number | null = null;
    const measure = (): void => {
      const anchors = spotAnchors(step, detour, run.context());
      const r = anchors.length > 0 ? anchorRect(anchors) : null;
      if (r === null) {
        rect = null;
        const now = Date.now();
        if (missingSince === null) missingSince = now;
        if (now - missingSince >= anchorGraceMs) run.anchorMissing(step.id);
        return;
      }
      missingSince = null;
      if (dwell) {
        const over = anchors.some((a) => {
          const el = resolveAnchor(a);
          return !!el && isHovering(el);
        });
        const now = Date.now();
        if (!over) hoverSince = null;
        else if (hoverSince === null) hoverSince = now;
        else if (now - hoverSince >= dwell.ms) run.hovered(step.id);
      }
      if (
        !sameRect(
          r,
          untrack(() => rect),
        )
      )
        rect = r;
    };
    measure();
    const t = setInterval(measure, pollMs);
    return () => clearInterval(t);
  });

  // ---- Copy ----
  const keys = $derived(effectiveBindings($settings.shortcuts.bindings));
  const copyCtx: CopyContext = $derived({
    helpKey: keys.toggleHelp ? formatChord(keys.toggleHelp, isMacLike()) : "",
    settingsKey: keys.openSettings ? formatChord(keys.openSettings, isMacLike()) : "",
    nextKey: keys.passPriority ? formatChord(keys.passPriority, isMacLike()) : "",
  });
  // What the card says, strongest first: a detour (what to do before
  // the step can happen), then the recovered copy, then the step's own.
  const recovering = $derived(snap.coach === "recovered" && !!snap.step.recover);
  const said = $derived(snap.detour ?? (recovering ? snap.step.recover : undefined) ?? snap.step);
  const title = $derived(copyText(said.title, copyCtx));
  const body = $derived(copyText(said.body, copyCtx));
  // A detour has no hint: it already says what to do.
  const hint = $derived(
    snap.detour ? "" : copyText(recovering ? snap.step.recover?.hint : snap.step.hint, copyCtx),
  );
  const status = $derived(statusText(snap.step, { ...run.context(), view, client: { autopass } }));

  // ---- Placement ----
  // The phone strip can fold to one line; a desktop card cannot, so a
  // fold made on a phone does not survive a widened window.
  let phone = $state(false);
  let folded = $state(false);
  $effect(() => {
    const mq = typeof matchMedia === "function" ? matchMedia("(max-width: 599px)") : null;
    if (!mq) return;
    const read = (): void => {
      phone = mq.matches;
    };
    read();
    mq.addEventListener?.("change", read);
    return () => mq.removeEventListener?.("change", read);
  });
  const minimised = $derived(phone && folded);

  // `onSize` is called untracked: the parent's handler reads its own
  // size state to skip no-op writes, and a tracked read there would make
  // this effect depend on what it writes (and its cleanup's 0 × 0 would
  // then loop with the next report).
  let root: HTMLElement | null = $state(null);
  $effect(() => {
    const el = root;
    const cb = onSize;
    if (!cb) return;
    const send = (w: number, h: number): void => untrack(() => cb(w, h));
    if (!el) {
      send(0, 0);
      return;
    }
    const report = (): void => send(el.offsetWidth, el.offsetHeight);
    report();
    if (typeof ResizeObserver === "undefined") return () => send(0, 0);
    const ro = new ResizeObserver(report);
    ro.observe(el);
    return () => {
      ro.disconnect();
      send(0, 0);
    };
  });
</script>

{#if snap.visible}
  {#if rect}
    <TutorialScrim {rect} />
  {/if}
  <div class="coach-slot" bind:this={root}>
    <CoachCard
      state={snap.coach}
      n={snap.step.n}
      total={TUTORIAL_STEP_COUNT}
      {title}
      {body}
      {hint}
      {status}
      {minimised}
      onStart={() => run.start()}
      onSkipTutorial={() => run.close()}
      onSkipStep={() => run.skipStep()}
      onReplay={() => {
        run.close();
        onReplay();
      }}
      onFinish={() => run.close()}
      onToggleMinimise={() => (folded = !folded)}
    />
  </div>
{/if}

<style>
  /* ADR 0076 §2.3 (amended 2026-10-02): bottom-left of the play area,
     inset by the dock's --dock-inset so its edges line up with the self
     panel's content box, exactly where PlayerPanel's .coach-spacer
     keeps the cell empty. ~300px at 1280 wide, so a seven-card fan
     still reads beside it. z 57: over the scrim (56) and the dock (55);
     under the card-local menus (60+), the modals (200) and the hover
     zoom (300). */
  .coach-slot {
    position: absolute;
    left: var(--dock-inset, 15px);
    bottom: var(--dock-inset, 15px);
    z-index: 57;
    width: clamp(280px, 24vw, 340px);
  }
  /* §2.3 on a phone: a strip stacked on the dock bar, inside the same
     16px gutter. Game.svelte pads the play area by its height too, so
     the board ends above it. */
  @media (max-width: 599px) {
    .coach-slot {
      left: 6px;
      right: 6px;
      bottom: calc(var(--dock-h, 0px) + 6px);
      width: auto;
    }
  }
</style>

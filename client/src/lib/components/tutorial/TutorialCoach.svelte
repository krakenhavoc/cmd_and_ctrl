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
  //      rest (steps 3, 5 and 10, "hover ≥ 600ms"), and reads an anchor
  //      that comes off the board (step 2's roll, step 9's card) afresh
  //      each tick.
  //   3. Placement. The card docks bottom-left (§2.3, amended
  //      2026-10-02). It publishes its live size through `onSize`, and
  //      Game.svelte turns that into --coach-w / --coach-h: on a desktop
  //      the viewer's own panel keeps an empty cell that size at the left
  //      of its bottom row (PlayerPanel's .coach-spacer, the twin of the
  //      dock's .dock-spacer), so the hand centres in what is left and
  //      nothing sits under the card. The card is taller than that row,
  //      so the piles and the back row start right of it too (#2990). On a phone the card is a strip
  //      stacked on the dock bar, and the play area's bottom padding
  //      grows by its height.
  //
  // Skip tutorial and Finish hide the card and leave the player at the
  // practice table; the table is still a game. Replay opens a fresh one.
  //
  // A step that completes marks the first-use hints it teaches as seen
  // (ADR 0125 §5.3), through `onTaught`; a skipped step marks nothing.

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
    skipNoteText,
    spotAnchors,
    spotlit,
    statusText,
    tutorialRunning,
    type CopyContext,
    type TutorialStep,
  } from "../../tutorial";
  import { anchorRects, resolveAnchor, sameRects, type AnchorRect } from "../../tutorialAnchor";
  import { TUTORIAL_STEPS, abilityCardName } from "../../tutorialSteps";
  import { HINTS } from "../../hints";
  import type { HintID } from "../../hints/hint";
  import { markHintsTaught } from "../../hints/runtime";
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
    /** The dock's autopass toggle (Game.svelte's session state), for step 12. */
    autopass?: boolean;
    /**
     * A completed step's `teaches`. Defaults to marking those hints seen
     * at their current version (ADR 0125 §5.3).
     */
    onTaught?: (ids: readonly HintID[]) => void;
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
    onTaught = (ids: readonly HintID[]) => markHintsTaught(ids, HINTS),
  }: Props = $props();

  const run = untrack(() =>
    createTutorialRun(steps, {
      viewerID,
      view,
      log,
      client: { autopass },
      onComplete: (step) => {
        if (step.teaches?.length) onTaught(step.teaches);
      },
    }),
  );
  let snap = $state(run.current());
  const unsubRun = run.subscribe((s) => (snap = s));

  // The game moved: check the step.
  $effect(() => {
    run.observe(view);
  });
  // The autopass toggle never reaches the wire; step 12 reads it here.
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
  // event to wait for (one Forest is not a pile; the command zone has no
  // event at all, since a tap there casts), so it gives up after a read
  // rather than wait on something that cannot happen (§2.4).
  // Derived, so a republish of the same step (a hint, a detour) does
  // not restart the timer.
  const currentStep = $derived(snap.step);
  const shown = $derived(snap.visible);
  // While the opening roll is open nothing is dealt, so the read starts
  // once the hand is (ADR 0125 §5.2, heldByOpeningRoll).
  const held = $derived(snap.held);
  const detouring = $derived(snap.detour != null);
  $effect(() => {
    const step = currentStep;
    // Not while a detour is up either: the step is current behind
    // "First, keep your hand", and its read starts once that is done.
    if (canHover || !step.hover || !shown || held || detouring) return;
    const t = setTimeout(
      () => run.advance(step.id, "cannot be hovered on this device"),
      touchHoverStepMs,
    );
    return () => clearTimeout(t);
  });

  // While the card is up, the ability menu leaves out its Sandbox rows
  // (Card.svelte), so a right-click shows only the card's own abilities
  // (ADR 0076 §2.2 as amended 2026-10-11, #2992).
  $effect(() => {
    tutorialRunning.set(snap.visible);
  });

  onDestroy(() => {
    unsubBus();
    unsubRun();
    run.destroy();
    tutorialRunning.set(false);
  });

  // ---- The card preview ----
  // A step that points at the portraits moves the hover preview to the
  // left (HoverZoomOverlay reads the attribute), so it does not cover them.
  $effect(() => {
    if (typeof document === "undefined") return;
    if (!(snap.visible && snap.step.previewLeft && !snap.detour)) return;
    const root = document.documentElement;
    root.dataset.tutorialPreview = "left";
    return () => {
      delete root.dataset.tutorialPreview;
    };
  });

  // ---- The spotlight ----
  // Where the pointer last moved, from the window's pointermove. :hover
  // alone cannot say a pointer moved: a hand that slides in under a
  // resting pointer is :hover too (#3005).
  let lastPointer: { x: number; y: number } | null = null;
  $effect(() => {
    const onMove = (e: PointerEvent | MouseEvent): void => {
      lastPointer = { x: e.clientX, y: e.clientY };
    };
    window.addEventListener("pointermove", onMove, { passive: true });
    return () => window.removeEventListener("pointermove", onMove);
  });
  // A pointer must travel this far from where it was when the step's own
  // copy appeared before a rest on the anchor counts.
  const MOVE_PX = 8;
  // Hover steps whose pointer has really moved since their copy showed.
  // Outside the effect so a republish of the same step does not forget
  // it. The origin is taken when the copy first shows (no detour).
  const pointerMoved = new Set<string>();
  const pointerOrigin = new Map<string, { x: number; y: number } | null>();
  let rects: AnchorRect[] = $state([]);
  $effect(() => {
    const step = snap.step;
    const detour = snap.detour;
    const lit = snap.visible && spotlit(snap.coach);
    // A step with no anchor at all (none in the script) points at
    // nothing. One whose anchor is read off the board is measured every
    // tick, and naming nothing then is a missing anchor.
    if (!lit || (step.anchor === undefined && !detour?.anchor)) {
      rects = [];
      return;
    }
    // A hover step completes once the pointer has rested on its anchor
    // for step.hover.ms (§2.1). Measured here, on the poll that already
    // reads the anchor: the bus says when the pointer arrives, but not
    // that it stayed, and a pointer crossing the hand on its way to the
    // dock is not reading it. Not while a detour points elsewhere.
    const dwell = canHover && !detour ? step.hover : undefined;
    if (dwell && !pointerOrigin.has(step.id)) pointerOrigin.set(step.id, lastPointer);
    let hoverSince: number | null = null;
    let missingSince: number | null = null;
    const measure = (): void => {
      const anchors = spotAnchors(step, detour, run.context());
      const r = anchors.length > 0 ? anchorRects(anchors) : [];
      if (r.length === 0) {
        rects = [];
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
        // A rest counts only once the pointer has really moved since this
        // step's copy appeared: Keep hand leaves it resting where the
        // hand lands, and that is not the gesture the step teaches.
        if (!pointerMoved.has(step.id) && lastPointer) {
          const o = pointerOrigin.get(step.id);
          if (!o || Math.hypot(lastPointer.x - o.x, lastPointer.y - o.y) > MOVE_PX)
            pointerMoved.add(step.id);
        }
        if (!over || !pointerMoved.has(step.id)) hoverSince = null;
        else if (hoverSince === null) hoverSince = now;
        else if (now - hoverSince >= dwell.ms) run.hovered(step.id);
      }
      if (
        !sameRects(
          r,
          untrack(() => rects),
        )
      )
        rects = r;
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
    cardName: abilityCardName(view, viewerID),
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
  // The steps the tutorial moved past by itself on the way here, in one
  // line (#2992): "Skipped step 7: the game is over."
  const note = $derived(skipNoteText(snap.skipped));

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
  {#if rects.length > 0}
    <TutorialScrim {rects} />
  {/if}
  <!-- data-bottom-left-fixture: the stack pile stays above it (ADR 0119 §1). -->
  <div class="coach-slot" data-bottom-left-fixture bind:this={root}>
    <CoachCard
      state={snap.coach}
      n={snap.step.n}
      total={TUTORIAL_STEP_COUNT}
      {title}
      {body}
      {hint}
      {status}
      {note}
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

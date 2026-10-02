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
  //   3. Toggles row — hold, autopass, bluff and the one Undo, with
  //      its count (`group "priority controls"`).
  //   4. Prompt area — the open request (lib/dock.ts, PR 3), if any:
  //      its question line, its row of options and a refusal it can
  //      answer, all inside one non-modal `role="dialog"` named as the
  //      control it replaced was.
  //   5. Action bar — the secondary on the left, ONE primary on the
  //      right. Pass turn and `next`, unless a request that takes the
  //      bar is open (a block declaration, a combat selection): then its
  //      buttons, inside its dialog, until it closes.
  //
  // `next` and Pass turn are rendered disabled, never hidden, when the
  // viewer cannot press them: the e2e suite reads `isEnabled()` on both
  // and the tutorial anchors to them.
  //
  // Keys (ADR 0111 §1, ADR 0047): Space is still "pass priority" through
  // the shortcut layer and still defers to a focused button. Enter on a
  // focused action-bar button presses that button and nothing else (it
  // is stopped here, so the window handler below does not also fire).
  //
  // PR 4: this is the table's ONE Enter / Escape handler for prompts.
  // Enter presses the open request's primary and Escape its cancel, as
  // lib/dock.ts's dockKeyFor decides: only a button that advertises the
  // key, never while a modal layer or a text field has the keyboard,
  // never Enter on a focused control. It replaced Game.svelte's
  // window-level targeting Enter / Escape and its combat-selection
  // Escape. Enter with focus on the board never presses `next`, and
  // Escape never touches it: `next` is not a request's button. Focus
  // does not move on an ordinary priority frame: that would steal it on
  // every pass.
  //
  // PR 6: a request may be a SHEET (owner decision 2): its body is drawn
  // in a panel that grows up out of the dock, inside the same dialog as
  // its bar buttons, with a minimise control on its top edge. The rest
  // of the board stays visible and clickable (no backdrop); focus goes
  // into the sheet when it opens; minimised, the request stays open and
  // a restore chip takes the prompt area. The sheet's own modal layer
  // ("sheet") stands the global shortcuts down but not this handler.

  import type { GameView } from "../../protocol";
  import PhaseDisplay from "./PhaseDisplay.svelte";
  import BluffChip from "./BluffChip.svelte";
  import Icon from "../Icon.svelte";
  import {
    activeDockRequest,
    dockKeyFor,
    isTypingTarget,
    sheetKey,
    sheetMaxHeight,
    sheetWidth,
    takesBar,
    type DockAction,
  } from "../../dock";
  import { foreignModalOpen } from "../../modalLayers";
  import { tick } from "svelte";
  import { formatUndoCount, isUnlimitedUndo } from "../../tableSettings";
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
    // The one Undo (ADR 0111 §1, PR 3). `undosLeft` is the seat's
    // undos_remaining (-1 on a table with no limit); `canUndo` is the
    // server's budget rule as Game.svelte reads it (hasUndoBudget).
    undosLeft?: number;
    canUndo?: boolean;
    onUndo?: () => void;
    // The dock's live size, for --dock-w / --dock-h (ADR 0111 §4).
    onSize?: (width: number, height: number) => void;
    // PR 6: the open sheet's width (0 when none is open, or it is
    // minimised), so the hover zoom can move left of it (§4).
    onSheet?: (width: number) => void;
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
    undosLeft = 0,
    canUndo = false,
    onUndo = () => {},
    onSize,
    onSheet,
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

  // ---- the open request (lib/dock.ts) -----------------------------------
  const req = $derived($activeDockRequest);
  const reqTakesBar = $derived(takesBar(req));
  // A key from the binding map (gated by the shortcut switch), or one
  // the prompt always owns (Escape).
  function actionKeys(a: DockAction): string | undefined {
    return a.keyShortcuts ?? ariaKeys(a.chord ?? "");
  }
  function actionTitle(a: DockAction): string | undefined {
    if (a.title === undefined) return undefined;
    return a.title + (a.chord ? keyHint(a.chord) : "");
  }

  // The one Enter / Escape handler (PR 4). The dock's own buttons stop
  // Enter before it gets here.
  // PR 6: a sheet's own modal layer is not "a modal over the dock", so
  // the handler reads foreignModalOpen, not modalOpen.
  function onWindowKey(e: KeyboardEvent): void {
    const a = dockKeyFor(e, req, { modalOpen: $foreignModalOpen });
    if (!a) return;
    e.preventDefault();
    a.onPress();
  }

  // ---- sheets (ADR 0111 §3, Delivery PR 6) ------------------------------
  // A request with `sheet` draws its body in a panel that grows up out
  // of the dock. It can be minimised to its question line; the request
  // stays open, its buttons stay in the bar, and a restore chip takes
  // the prompt area. A new question (a new sheet key) comes back up.
  const sheetOpen = $derived(!!req?.sheet && !!req.body);
  const currentSheetKey = $derived(sheetKey(req));
  let minimisedKey: string | null = $state(null);
  const minimised = $derived(
    sheetOpen && minimisedKey !== null && minimisedKey === currentSheetKey,
  );
  const sheetTitle = $derived(req?.sheet?.title ?? req?.label ?? "");
  let sheetEl: HTMLElement | null = $state(null);
  let restoreEl: HTMLButtonElement | null = $state(null);

  async function minimise(): Promise<void> {
    minimisedKey = currentSheetKey;
    await tick();
    restoreEl?.focus();
  }
  async function restore(): Promise<void> {
    minimisedKey = null;
    await tick();
    focusIntoSheet();
  }

  // Focus goes into the sheet when it opens (ADR 0111 PR 6): to the
  // field or control the body marks `data-sheet-focus` (the creature-
  // type filter), else to the sheet itself, which is not a control, so
  // Enter there presses the sheet's confirm through the handler above.
  function focusIntoSheet(): void {
    const el = sheetEl;
    if (!el) return;
    const target = el.querySelector<HTMLElement>("[data-sheet-focus]") ?? el;
    target.focus();
  }
  let sheetFocusedFor: string | null = null;
  $effect(() => {
    const key = currentSheetKey;
    const el = sheetEl;
    if (!key || !el) {
      if (!key) sheetFocusedFor = null;
      return;
    }
    if (sheetFocusedFor === key) return;
    sheetFocusedFor = key;
    // Never out of a text field the player is typing in elsewhere.
    const active = document.activeElement;
    if (active && active !== document.body && !el.contains(active) && isTypingTarget(active)) {
      return;
    }
    focusIntoSheet();
  });

  // The play area's height, for the sheet's ceiling (60% of it on a
  // desktop, 70% on a phone, and never past the room above the dock).
  let playH = $state(0);
  let dockH = $state(0);
  let phone = $state(false);
  $effect(() => {
    const el = root;
    if (!el) return;
    const parent = el.offsetParent as HTMLElement | null;
    const mq = typeof matchMedia === "function" ? matchMedia("(max-width: 599px)") : null;
    const read = (): void => {
      playH = parent?.clientHeight ?? 0;
      dockH = el.offsetHeight;
      phone = mq?.matches ?? false;
    };
    read();
    mq?.addEventListener?.("change", read);
    if (typeof ResizeObserver === "undefined") {
      return () => mq?.removeEventListener?.("change", read);
    }
    const ro = new ResizeObserver(read);
    if (parent) ro.observe(parent);
    ro.observe(el);
    return () => {
      ro.disconnect();
      mq?.removeEventListener?.("change", read);
    };
  });
  const sheetMaxH = $derived(sheetMaxHeight(playH, dockH, phone));

  // The open sheet's width, for the hover zoom (§4: it moves left of
  // an open sheet, so hovering a card in a scry shows it beside it).
  $effect(() => {
    const el = sheetEl;
    if (!onSheet) return;
    if (!el || !sheetOpen || minimised) {
      onSheet(0);
      return;
    }
    const report = (): void => onSheet(el.offsetWidth);
    report();
    if (typeof ResizeObserver === "undefined") return () => onSheet(0);
    const ro = new ResizeObserver(report);
    ro.observe(el);
    return () => {
      ro.disconnect();
      onSheet(0);
    };
  });

  // ADR 0111 §1, keyboard focus: when a request the game waits on opens
  // (a choice, a block declaration) and focus is on the body, its
  // primary takes focus, so a keyboard player answers it at once and a
  // screen reader lands on it. Once per request, and never on an
  // ordinary priority frame: that would steal focus every pass.
  //
  // PR 5: a request may ask for its dialog to take focus instead
  // (`focus: "dialog"`): a yes/no question, whose "Yes" must not be one
  // stray Enter away. The dialog is focusable for that (tabindex -1)
  // and is not a control, so Enter on it presses nothing.
  let primaryEl: HTMLButtonElement | null = $state(null);
  let dialogEl: HTMLElement | null = $state(null);
  let focusedFor: string | null = null;
  $effect(() => {
    const r = req;
    const el = r?.focus === "dialog" ? dialogEl : primaryEl;
    if (!r) {
      focusedFor = null;
      return;
    }
    // A sheet takes focus into itself instead (below).
    if (!el || r.sheet || (r.rank !== "choice" && r.rank !== "blocks")) return;
    if (focusedFor === r.label) return;
    focusedFor = r.label;
    const active = document.activeElement;
    if (!active || active === document.body) el.focus();
  });

  // ---- the one Undo ------------------------------------------------------
  const undoUnlimited = $derived(isUnlimitedUndo(undosLeft));
  const undoCount = $derived(formatUndoCount(undosLeft));
  const undoTitle = $derived(
    (undoUnlimited
      ? "undo your most recent action — this table has no undo limit"
      : !canUndo
        ? "no undos remaining this turn (refreshes on your next untap)"
        : `undo your most recent action — ${undoCount} left this turn`) + keyHint(keys.undo),
  );

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

<svelte:window onkeydown={onWindowKey} />

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
    {:else if hint && !reqTakesBar}
      <!-- What `next` will do — not while a request has taken the bar
           and `next` is not on it. -->
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
    <!-- ADR 0111 PR 3: the one Undo, out of the ⋯ menu and the attack
         row. Disabled, not hidden, when the budget is spent; the server
         keeps the rules, this only reads them. -->
    <button
      type="button"
      class="action undo"
      disabled={!canUndo}
      aria-label={undoUnlimited ? "Undo (no limit)" : `Undo (${undoCount} left)`}
      aria-keyshortcuts={ariaKeys(keys.undo)}
      onclick={onUndo}
      title={undoTitle}
    >
      <Icon name="undo" size={13} /><span class="undo-word">Undo</span><span class="undo-count"
        >{undoCount}</span
      >
    </button>
  </div>

  {#snippet actionButton(a: DockAction, cls: string)}
    <button
      type="button"
      class={cls}
      class:emphasis={a.emphasis}
      class:align-end={a.alignEnd}
      disabled={a.disabled}
      aria-pressed={a.pressed}
      aria-label={a.ariaLabel}
      aria-keyshortcuts={actionKeys(a)}
      title={actionTitle(a)}
      onclick={a.onPress}
      onkeydown={(e) => enterPresses(e, !a.disabled, a.onPress)}
    >
      {#if a.icon}<Icon name={a.icon} size={12} />{/if}
      {#if a.seatColor}<span class="seat-dot" style="background:{a.seatColor}"></span>{/if}
      {a.label}
      {#if a.note}<span class="note">{a.note}</span>{/if}
      {#if a.cap}<kbd class="cap" aria-hidden="true">{a.cap}</kbd>{/if}
    </button>
  {/snippet}

  <!-- The open request (lib/dock.ts): one non-modal dialog holding its
       question, its options and, when it takes the bar, its buttons. -->
  {#if req}
    <div
      class="dock-request"
      role="dialog"
      aria-label={req.label}
      data-rank={req.rank}
      tabindex={req.focus === "dialog" ? -1 : undefined}
      bind:this={dialogEl}
    >
      {#if sheetOpen && req.body}
        <!-- ADR 0111 §3: the sheet grows up out of the dock, right-
             aligned with it. It is inside the request's dialog, so its
             body and the bar's buttons share one dialog (and one name).
             Minimised, it stays mounted (the picker keeps its scroll
             and its fields) and is hidden; the restore chip below
             brings it back. -->
        <div
          class="dock-sheet"
          class:minimised
          hidden={minimised}
          tabindex="-1"
          bind:this={sheetEl}
          style:--sheet-want="{sheetWidth(req)}px"
          style:max-height={sheetMaxH > 0 ? `${sheetMaxH}px` : undefined}
        >
          <header class="sheet-head">
            <h2 class="sheet-title">
              {sheetTitle}{#if req.sheet?.src}<span class="prompt-src" aria-hidden="true"
                  >{req.sheet.src}</span
                >{/if}
            </h2>
            <button
              type="button"
              class="sheet-min"
              aria-label="minimise"
              aria-expanded="true"
              title="minimise — fold this down to look at the board; it stays open"
              onclick={minimise}><Icon name="chevron-down" size={14} /></button
            >
          </header>
          <div class="sheet-body">
            {@render req.body()}
          </div>
          {#if req.sheet?.count}
            <div class="sheet-foot">
              <span class="prompt-count">{req.sheet.count}</span>
            </div>
          {/if}
        </div>
      {/if}
      <div class="dock-prompt" role={req.group ? "group" : undefined} aria-label={req.group}>
        {#if sheetOpen && minimised}
          <!-- The minimised sheet's question line: one click (or Enter
               on it) brings the sheet back. -->
          <button
            type="button"
            class="sheet-restore"
            aria-label={`restore: ${sheetTitle}`}
            aria-expanded="false"
            title="bring the sheet back up"
            bind:this={restoreEl}
            onclick={restore}
          >
            <Icon name="chevron-up" size={13} />
            <span class="restore-title">{sheetTitle}</span>
            {#if req.sheet?.count}<span class="q-detail">{req.sheet.count}</span>{/if}
          </button>
        {/if}
        {#if req.question}
          <div
            class="dock-question"
            role={req.live ? "status" : undefined}
            aria-live={req.live ? "polite" : undefined}
          >
            {#if req.tag}<span class="q-tag tone-{req.tone ?? 'plain'}">{req.tag}</span>{/if}
            <span class="q-text"
              >{req.question}{#if req.detail}<span class="q-detail">{` · ${req.detail}`}</span
                >{/if}</span
            >
          </div>
        {/if}
        {#if req.hint}
          <p class="dock-hint" class:warn={req.hintWarn}>{req.hint}</p>
        {/if}
        {#if req.body && !req.sheet}
          {@render req.body()}
        {/if}
        {#if req.row && req.row.length > 0}
          <div class="dock-row" class:stack={req.rowLayout === "stack"}>
            {#if req.rowLead}<span class="row-lead">{req.rowLead}</span>{/if}
            {#each req.row as a (a.id)}
              {@render actionButton(a, "dock-btn row-btn")}
            {/each}
          </div>
        {/if}
        {#if req.refusal}
          <div class="dock-refusal" class:danger={req.refusal.tone === "danger"} role="alert">
            <span class="q-tag tone-{req.refusal.tone ?? 'gold'}">{req.refusal.tag}</span>
            <span class="q-text"
              ><strong>{req.refusal.text}</strong>{#if req.refusal.detail}<span class="q-detail"
                  >{` · ${req.refusal.detail}`}</span
                >{/if}</span
            >
            <span class="refusal-actions">
              {#each req.refusal.actions as a (a.id)}
                {@render actionButton({ ...a, emphasis: true }, "dock-btn row-btn")}
              {/each}
              {#if req.refusal.onDismiss}
                <button
                  type="button"
                  class="refusal-close"
                  aria-label="dismiss"
                  onclick={req.refusal.onDismiss}><Icon name="x" size={12} /></button
                >
              {/if}
            </span>
          </div>
        {/if}
      </div>
      {#if reqTakesBar && (req.primary || (req.secondary?.length ?? 0) > 0)}
        <!-- The request's own bar: its secondaries left, its one primary
             in the corner. `next` and Pass turn give way until it closes. -->
        <div class="dock-bar">
          {#each req.secondary ?? [] as a (a.id)}
            {@render actionButton(a, "dock-btn secondary")}
          {/each}
          {#if req.primary}
            {@const p = req.primary}
            <button
              type="button"
              class="dock-btn primary request-primary"
              bind:this={primaryEl}
              disabled={p.disabled}
              aria-label={p.ariaLabel}
              aria-keyshortcuts={actionKeys(p)}
              title={actionTitle(p)}
              onclick={p.onPress}
              onkeydown={(e) => enterPresses(e, !p.disabled, p.onPress)}
            >
              {p.label}{#if p.cap}<kbd class="cap" aria-hidden="true">{p.cap}</kbd>{/if}
            </button>
          {/if}
        </div>
      {/if}
    </div>
  {/if}

  <!-- The action bar. Its slots never move: secondary on the left, the
       one primary in the very corner, so the pointer's resting place is
       always the button that moves the game on. -->
  {#if !reqTakesBar}
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
  {/if}
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
    flex-wrap: wrap;
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
  /* A sheet's verb can be a sentence ("Don't discard — put Mox Diamond
     into its owner's graveyard"): it wraps rather than clips. */
  .dock-btn.request-primary,
  .dock-bar .dock-btn.secondary {
    white-space: normal;
    line-height: 1.2;
    padding-block: 4px;
    text-align: center;
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

  /* The one Undo: an icon, the word, and the count left. */
  .action.undo {
    flex: 0 1 auto;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
  }
  .undo-count {
    font-variant-numeric: tabular-nums;
    opacity: 0.8;
  }

  /* The open request (lib/dock.ts). Non-modal: it sits in the dock and
     blocks nothing on the board. */
  .dock-request {
    display: flex;
    flex-direction: column;
    gap: 5px;
    min-width: 0;
  }
  .dock-prompt {
    display: flex;
    flex-direction: column;
    gap: 5px;
    min-width: 0;
  }
  .dock-question {
    display: flex;
    align-items: baseline;
    gap: 6px;
    font-size: 0.8rem;
    line-height: 1.35;
    color: var(--fg);
    min-width: 0;
  }
  .q-tag {
    flex: none;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 0.64rem;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    font-weight: 700;
    color: var(--fg-muted);
  }
  .q-tag.tone-danger {
    color: var(--danger);
  }
  .q-tag.tone-gold {
    color: var(--accent-strong);
  }
  .q-text {
    min-width: 0;
  }
  .q-detail {
    color: var(--fg-dim);
  }
  .dock-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
  }
  .row-lead {
    font-size: 0.75rem;
    color: var(--fg-muted);
  }
  .dock-btn.row-btn {
    min-height: 28px;
    padding: 0 10px;
    font-size: 0.82em;
    white-space: normal;
    text-align: left;
  }
  /* The row's emphasised option (Attack … with all). Accent outline,
     not gold fill: gold is the action bar's one primary. */
  .dock-btn.emphasis {
    border-color: color-mix(in srgb, var(--accent) 60%, transparent);
    color: var(--accent-strong);
    background: var(--accent-soft);
  }
  .dock-btn.emphasis:hover:not(:disabled) {
    background: rgba(217, 180, 92, 0.24);
  }
  .note {
    color: var(--fg-dim);
    font-weight: 500;
  }
  .seat-dot {
    flex: none;
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
  /* An inline choice's hint (PR 5): what the answer does. */
  .dock-hint {
    margin: 0;
    font-size: 0.72rem;
    line-height: 1.35;
    color: var(--fg-dim);
  }
  .dock-hint.warn {
    color: var(--danger);
  }
  .dock-request:focus {
    outline: none;
  }
  .dock-request:focus-visible {
    outline: 1px solid color-mix(in srgb, var(--accent) 70%, transparent);
    outline-offset: 2px;
    border-radius: 4px;
  }
  /* option_pick: one option per line, full width. */
  .dock-row.stack {
    flex-direction: column;
    align-items: stretch;
  }
  .dock-row.stack .dock-btn.row-btn {
    justify-content: flex-start;
    min-height: 30px;
    padding: 4px 10px;
  }
  .dock-btn.align-end {
    margin-left: auto;
  }
  .dock-btn[aria-pressed="true"] {
    border-color: var(--accent);
  }
  .dock-refusal.danger {
    border-color: color-mix(in srgb, var(--danger) 45%, transparent);
    background: color-mix(in srgb, var(--danger) 10%, transparent);
  }
  .dock-refusal {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 6px;
    padding: 6px 8px;
    border: 1px solid rgba(217, 180, 92, 0.45);
    border-radius: 6px;
    background: rgba(217, 180, 92, 0.08);
    font-size: 0.78rem;
    line-height: 1.35;
    color: var(--fg);
  }
  .refusal-actions {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin-left: auto;
  }
  .refusal-close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    padding: 0;
    border: none;
    border-radius: 5px;
    background: transparent;
    color: var(--fg-muted);
    cursor: pointer;
  }
  .refusal-close:hover {
    background: rgba(255, 255, 255, 0.08);
  }
  .refusal-close:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  /* A request's primary (No blocks, Done blocking) is the gold button
     in the corner, as `next` is with priority. */
  .dock-btn.request-primary:not(:disabled) {
    background: var(--accent);
    color: var(--accent-fg);
    font-weight: 700;
    border-color: var(--accent-strong);
  }
  .dock-btn.request-primary:hover:not(:disabled) {
    background: var(--accent-strong);
  }

  /* ADR 0111 §3 (PR 6): a sheet grows up out of the dock, right-aligned
     with it: as wide as its body asks (--sheet-want, capped at 720px
     and at the screen less 24px, never narrower than the dock) and as
     tall as it needs up to 60% of the play area (max-height, set inline
     from the measured play area; 60vh until it is measured), scrolling
     inside. A dim edge, not a backdrop: the rest of the board stays
     visible, unblurred and clickable. It is inside the dock (z 55), so
     card-local menus, the full-screen modals and the hover zoom stay
     above it. */
  .dock-sheet {
    position: absolute;
    right: -1px;
    bottom: calc(100% + 6px);
    box-sizing: border-box;
    width: min(var(--sheet-want, 560px), calc(100vw - 24px));
    min-width: calc(100% + 2px);
    max-height: 60vh;
    display: flex;
    flex-direction: column;
    background: var(--surface);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg, 12px);
    box-shadow:
      0 0 0 1px rgba(0, 0, 0, 0.35),
      -18px -6px 40px rgba(0, 0, 0, 0.45),
      var(--shadow-lg);
    color: var(--fg);
    animation: sheet-up 180ms var(--ease);
  }
  .dock-sheet[hidden] {
    display: none;
  }
  .dock-sheet:focus {
    outline: none;
  }
  .dock-sheet:focus-visible {
    outline: 1px solid color-mix(in srgb, var(--accent) 45%, transparent);
    outline-offset: 2px;
  }
  @keyframes sheet-up {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  .sheet-head {
    flex: none;
    display: flex;
    align-items: flex-start;
    gap: 8px;
    padding: 12px 10px 6px 16px;
  }
  .sheet-title {
    flex: 1 1 auto;
    min-width: 0;
    margin: 0;
    font-family: var(--font-display);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: -0.01em;
    text-transform: none;
    line-height: 1.25;
    color: var(--fg);
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 4px 10px;
  }
  .sheet-min {
    flex: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 26px;
    padding: 0;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--fg-muted);
    cursor: pointer;
    box-shadow: none;
  }
  .sheet-min:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--fg);
  }
  .sheet-body {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
    overflow-x: hidden;
    padding: 2px 16px 12px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .sheet-foot {
    flex: none;
    padding: 6px 16px 10px;
    border-top: 1px solid var(--border);
  }
  .sheet-restore {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
    min-height: 30px;
    padding: 4px 10px;
    border-radius: 7px;
    border: 1px dashed color-mix(in srgb, var(--accent) 55%, transparent);
    background: var(--accent-soft);
    color: var(--accent-strong);
    font-size: 0.8rem;
    font-weight: 600;
    text-align: left;
    cursor: pointer;
    box-shadow: none;
  }
  .restore-title {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .sheet-min:focus-visible,
  .sheet-restore:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
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
    /* One row: the secondaries share the left half, the primary the
       right (§8). */
    .dock-bar {
      flex-wrap: nowrap;
      gap: 6px;
    }
    .dock-btn {
      min-height: 44px;
    }
    .dock-btn.row-btn {
      min-height: 36px;
    }
    /* Four toggles share 358px: Undo becomes an icon chip with its
       count; its name is its aria-label either way. */
    .undo-word {
      display: none;
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
    /* §8: three or more secondaries move into a row of their own above
       the primary, which then takes the full width below them. */
    .dock-bar:has(.dock-btn.secondary + .dock-btn.secondary + .dock-btn.secondary) {
      flex-wrap: wrap;
    }
    .dock-bar:has(.dock-btn.secondary + .dock-btn.secondary + .dock-btn.secondary)
      .dock-btn.primary {
      flex-basis: 100%;
    }
    /* §8: a bottom sheet, the dock's full width, up to 70% of the play
       area (set inline), above the bar. */
    .dock-sheet {
      left: -1px;
      right: -1px;
      width: auto;
      min-width: 0;
      max-height: 70vh;
      border-radius: 12px 12px 8px 8px;
    }
    .sheet-head {
      padding: 10px 8px 4px 12px;
    }
    .sheet-body {
      padding: 2px 12px 10px;
    }
    .sheet-foot {
      padding: 6px 12px 8px;
    }
  }
</style>

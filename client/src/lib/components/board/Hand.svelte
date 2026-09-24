<script lang="ts">
  // Hand renders the bottom-center hand strip for one player.
  //
  //   - Self: cards are face-up, click → cast_spell. Layout is a
  //     centered fan, each card rotated proportional to its offset
  //     from the centre, like the old Pixi version.
  //   - Opponents: cards are face-down. The server's FilterViewFor
  //     zeroes out the contents and the count is the only signal
  //     of how many they hold; we render `count` placeholder
  //     face-down cards in a tighter fan so the player count is
  //     visible at a glance without leaking content.
  //   - #1524: the viewer's own hand shows in THEIR order — dragged
  //     sideways or sorted from the hand menu, saved in this browser
  //     (lib/handOrder.ts). Opponents' hands keep the server's order.

  import type { Action } from "svelte/action";
  import type { CardView, GameView, ZoneView } from "../../protocol";
  import Card from "./Card.svelte";
  import { dealIn, dealOut } from "../../animations";
  import { fanAngle, fanLift, handOverlap } from "../../handFan";
  import { play } from "../../sounds";
  import { settings } from "../../settings";
  import { canCastFromHand, type Legality } from "../../timing";
  import { fetchAutoTapPreview, type AutoTapPreview } from "../../api";
  import { cardImageURL } from "../../cardImage";
  import {
    HAND_SORTS,
    applyHandOrder,
    handOrderKey,
    idsOf,
    loadHandOrder,
    moveCard,
    saveHandOrder,
    sameOrder,
    sortHand,
    type HandSort,
  } from "../../handOrder";
  import {
    CAST_ZONE_MARGIN_PX,
    IDLE,
    SNAP_REASON_MS,
    autoTapHighlight,
    clearAutoTapHighlight,
    dragVerdict,
    previewSourceIDs,
    step as stepDrag,
    wantsPreview,
    type DragInput,
    type DragOutcome,
    type DragState,
  } from "../../dragCast";

  interface Props {
    hand: ZoneView;
    isSelf: boolean;
    onPlayCard?: (card: CardView) => void;
    // #1508: drag a card out of the hand onto the table to cast it.
    // Called once the card is released past the line while castable;
    // the parent runs the same cast chain a click does, with the drag
    // flag (strict + auto-tap). Undefined turns the gesture off, which
    // it is for every hand but the viewer's own.
    onDragCast?: (card: CardView) => void;
    // #660: a card in hand can have activated abilities that function
    // THERE — cycling, typecycling (CR 702.29a/e). They ride
    // `zone_abilities` on the wire and open the same popover a
    // permanent's abilities do. Self hands only: an opponent's hand
    // cards carry no abilities on the wire in the first place.
    onActivateAbility?: (card: CardView, abilityIndex: number) => void;
    // #1228: and the MANA abilities that function from a hand —
    // "Exile this card from your hand: Add {R}" (CR 113.6). They ride
    // `zone_mana_abilities` and open the same popover, but they fire
    // `activate_mana_ability` rather than `activate_ability`, which is
    // why this is a second callback rather than a second index on the
    // first.
    onActivateManaAbility?: (card: CardView, abilityIndex: number) => void;
    // Why the CR 307.1 sorcery-speed window is shut, or "" when it is
    // open. Passed through to the popover exactly as the battlefield
    // rows pass it; no cycling ability is sorcery-speed today, but a
    // hand ability that is would grey correctly.
    sorcerySpeedBlocked?: string;
    // S13.3 — when present, each hand card is checked for legality
    // and rendered greyed-out with a reason tooltip if illegal.
    // Self hands only; opponent hands always render face-down with
    // no per-card legality (we don't know what they are).
    snap?: GameView | null;
    viewerID?: string | null;
  }

  const {
    hand,
    isSelf,
    onPlayCard,
    onDragCast,
    onActivateAbility,
    onActivateManaAbility,
    sorcerySpeedBlocked = "",
    snap = null,
    viewerID = null,
  }: Props = $props();

  // ---- #1524: the viewer's own order --------------------------------
  //
  // What is saved is read once per (game, viewer); a reorder or a sort
  // made on this page overrides it for the same key. Everything else —
  // the reconcile, the sorts, the storage and its failures — is
  // handOrder.ts.
  const gameID = $derived(snap?.id ?? null);
  const orderKey = $derived(gameID && viewerID ? handOrderKey(gameID, viewerID) : "");
  const storedOrder = $derived(isSelf ? loadHandOrder(gameID, viewerID) : null);
  let orderOverride = $state<{ key: string; ids: string[] } | null>(null);
  const savedOrder = $derived(
    orderOverride && orderOverride.key === orderKey ? orderOverride.ids : storedOrder,
  );

  function setOrder(ids: string[]): void {
    orderOverride = { key: orderKey, ids };
    saveHandOrder(gameID, viewerID, ids);
  }

  // Opponent hand composition:
  //   - hand.cards contains any cards the server has revealed to the
  //     viewer (Thoughtseize-style reveals, sticky per S13.5). Those
  //     render face-up.
  //   - The remaining (hand.count - revealed) cards render as face-
  //     down placeholders synthesized here.
  //   - settings.display.showOpponentHandCount collapses the full
  //     fan to a single face-down stand-in when the viewer prefers a
  //     compact opponent hand; revealed cards still render in full
  //     so a Thoughtseize peek isn't obscured by the setting.
  const cards = $derived.by((): CardView[] => {
    if (isSelf) return applyHandOrder(hand.cards, savedOrder);
    const revealed = hand.cards;
    const hidden = Math.max(0, hand.count - revealed.length);
    const showCount = $settings.display.showOpponentHandCount;
    // Cap the back-count when compact-opponent-hand mode is on, but
    // only if we also have no revealed cards; mixing a revealed card
    // with a single back stand-in reads cleaner than hiding the
    // revealed card entirely.
    const hiddenToShow = showCount ? hidden : Math.min(hidden, revealed.length > 0 ? hidden : 1);
    const out: CardView[] = [...revealed];
    for (let i = 0; i < hiddenToShow; i++) {
      out.push({
        instance_id: `opp-hand-${i}`,
        name: "",
        owner: "",
        controller: "",
      });
    }
    return out;
  });
  const layout = $derived($settings.display.handLayout);

  // #956 — the resting overlap for this layout, tightened for large
  // hands so the fan cannot outgrow its panel and get clipped. The
  // three resting values are the ones the CSS used to hard-code per
  // layout; handOverlap only ever tightens them.
  const overlapBase = $derived(layout === "stacked" ? 0.85 : isSelf ? 0.5 : 0.62);
  const overlap = $derived(handOverlap(cards.length, overlapBase));

  // Once an order is saved, keep it in step with the hand: a card that
  // left drops out and a new one is recorded at the right end, so a
  // card that comes BACK later (a bounce, an undo) is appended like any
  // other new card rather than returning to a stale slot. Quiet: no
  // write at all while the two already agree, and nothing is saved for
  // a hand that was never rearranged.
  $effect(() => {
    if (!isSelf || !savedOrder) return;
    const ids = idsOf(cards);
    if (!sameOrder(ids, savedOrder)) setOrder(ids);
  });

  // The hand menu's one-time sorts (owner decision 3).
  let sortMenuOpen = $state(false);
  let sortButton = $state<HTMLButtonElement | null>(null);
  let sortMenu = $state<HTMLElement | null>(null);
  const canSort = $derived(isSelf && hand.cards.length > 1);

  function sortBy(key: HandSort): void {
    setOrder(sortHand(cards, key));
    closeSortMenu(true);
  }

  function openSortMenu(): void {
    sortMenuOpen = true;
    queueMicrotask(() => sortMenu?.querySelector<HTMLElement>('[role="menuitem"]')?.focus());
  }

  function closeSortMenu(refocus: boolean): void {
    if (!sortMenuOpen) return;
    sortMenuOpen = false;
    if (refocus) sortButton?.focus();
  }

  function onSortMenuKey(ev: KeyboardEvent): void {
    const items = [...(sortMenu?.querySelectorAll<HTMLElement>('[role="menuitem"]') ?? [])];
    const at = items.indexOf(document.activeElement as HTMLElement);
    if (ev.key === "Escape") {
      ev.preventDefault();
      ev.stopPropagation();
      closeSortMenu(true);
    } else if (ev.key === "ArrowDown" || ev.key === "ArrowUp") {
      ev.preventDefault();
      const d = ev.key === "ArrowDown" ? 1 : -1;
      items[(at + d + items.length) % items.length]?.focus();
    } else if (ev.key === "Tab") {
      closeSortMenu(false);
    }
  }

  // A press anywhere outside the button and the menu closes it.
  $effect(() => {
    if (!sortMenuOpen) return;
    const onDown = (ev: PointerEvent) => {
      const t = ev.target as Node | null;
      if (t && (sortMenu?.contains(t) || sortButton?.contains(t))) return;
      closeSortMenu(false);
    };
    window.addEventListener("pointerdown", onDown, true);
    return () => window.removeEventListener("pointerdown", onDown, true);
  });

  function handleCardClick(card: CardView): void {
    if (!isSelf) return;
    onPlayCard?.(card);
  }

  // legalityFor computes the cast-from-hand legality for a single
  // card. Cheap enough to call on every render: since S31 the
  // verdict is a scan of the server's own `legal_moves` list rather
  // than a derivation. Returns a "spectator" Legality for opponent
  // hands so the .timing-disabled class stays off (we never grey
  // opponent hands — face-down already says "not yours", and the
  // wire ships no move list for anyone but the viewer anyway).
  function legalityFor(c: CardView): Legality {
    if (!isSelf) return { legal: true };
    return canCastFromHand(c, snap, viewerID);
  }

  // dealIn / dealOut are Svelte transitions (no mount/destroy callback
  // surface), so this action piggy-backs on the same element's
  // lifecycle: mount = "card entered hand" (draw), destroy = "card
  // left hand" (play). Mass events (opening-hand 7-card deal,
  // mulligan 7-card replacement) collapse to a single sound via the
  // 40ms same-name rate limit in sounds.ts — exactly what we want.
  const handLifecycle: Action<HTMLElement> = () => {
    play("draw");
    return {
      destroy() {
        play("play");
      },
    };
  };

  // ---- #1508: drag to cast, #1524: drag to reorder ---------------
  //
  // dragCast.ts owns every decision (the activation distance, the
  // line, the reorder band and its insertion index, the gold / red
  // verdict, reorder vs cast vs snap back); this block owns the DOM:
  // pointer listeners, the ghost that follows the pointer, the gap in
  // the fan, the drop zone over the table, the auto-tap preview fetch
  // and the board highlight it drives.
  //
  // Clicking and the keyboard are untouched. A press that never
  // travels DRAG_ACTIVATION_PX is a click and the Card's own handler
  // casts it; only a real drag swallows the click that follows it.

  let drag = $state<DragState>(IDLE);
  // The card being dragged. The verdict re-reads the LIVE card from
  // the hand, so a snapshot that lands mid-drag (priority moving on, a
  // spell resolving) re-colours the ghost.
  let dragCardID = $state<string | null>(null);
  let preview = $state<AutoTapPreview | null>(null);
  let previewToken = 0;
  // The ghost: the card, where it is, its size, and the hand slot it
  // snaps back to.
  let ghost = $state<{
    card: CardView;
    x: number;
    y: number;
    w: number;
    h: number;
    originX: number;
    originY: number;
    snapping: boolean;
  } | null>(null);
  // Where inside the card the pointer grabbed it.
  let grabX = 0;
  let grabY = 0;
  let snapReason = $state<{ text: string; x: number; y: number } | null>(null);
  let snapReasonTimer: ReturnType<typeof setTimeout> | null = null;
  let snapTimer: ReturnType<typeof setTimeout> | null = null;
  let dragSlot: HTMLElement | null = null;
  let dragPointerID: number | null = null;
  // The auto-tap preview is asked for once per drag, and only once the
  // card leaves the hand's band: a sideways reorder never fetches one.
  let previewRequested = false;
  // A drag that ended (cast, snapped back or cancelled) swallows the
  // one click the browser may still deliver for the same press.
  let swallowClick = false;

  const dragEnabled = $derived(isSelf && !!onDragCast);
  const dragging = $derived(drag.phase === "dragging");
  const reordering = $derived(dragging && drag.mode === "reorder");
  // gapShift: in reorder mode, the other cards slide apart around the
  // insertion point — those left of it one way, the rest the other.
  // The dragged card keeps its (dimmed) slot. 0 when not reordering.
  function gapShift(i: number, c: CardView): number {
    if (!reordering || c.instance_id === dragCardID) return 0;
    const from = cards.findIndex((x) => x.instance_id === dragCardID);
    const j = from >= 0 && i > from ? i - 1 : i;
    return j < drag.insertAt ? -1 : 1;
  }
  const liveDragCard = $derived(
    dragCardID ? (cards.find((c) => c.instance_id === dragCardID) ?? null) : null,
  );
  const verdict = $derived(
    liveDragCard
      ? dragVerdict(liveDragCard, legalityFor(liveDragCard), preview)
      : { castable: false, reason: "That card left your hand" },
  );

  // The ghost, the drop zone and the reason are position: fixed, and a
  // transformed ancestor (the fan's rotated slot, the hand's hover
  // lift) would make "fixed" mean "fixed to that ancestor". So they
  // are moved to <body>.
  function portal(node: HTMLElement): { destroy(): void } {
    document.body.appendChild(node);
    return {
      destroy() {
        node.remove();
      },
    };
  }

  function listen(): void {
    window.addEventListener("pointermove", onWindowMove);
    window.addEventListener("pointerup", onWindowUp);
    window.addEventListener("pointercancel", onWindowCancel);
    window.addEventListener("keydown", onWindowKey, true);
  }

  function unlisten(): void {
    window.removeEventListener("pointermove", onWindowMove);
    window.removeEventListener("pointerup", onWindowUp);
    window.removeEventListener("pointercancel", onWindowCancel);
    window.removeEventListener("keydown", onWindowKey, true);
  }

  function samePointer(ev: PointerEvent): boolean {
    return dragPointerID === null || ev.pointerId === undefined || ev.pointerId === dragPointerID;
  }

  function onSlotPointerDown(ev: PointerEvent, c: CardView, index: number): void {
    swallowClick = false;
    if (!dragEnabled || drag.phase !== "idle") return;
    if (ev.button !== 0 || ev.isPrimary === false) return;
    // A press inside the card's ability popover belongs to the popover.
    if ((ev.target as Element | null)?.closest?.('[role="menu"]')) return;
    const slot = ev.currentTarget as HTMLElement;
    const handEl = slot.closest(".hand") as HTMLElement | null;
    const handRect = (handEl ?? slot).getBoundingClientRect();
    const handTop = handRect.top;
    // #1524: the reorder band is the hand itself, and the insertion
    // points are the other cards' centres, both as they are now — the
    // gap that opens later must not move the targets under the pointer.
    const slotCenters: number[] = [];
    for (const el of handEl?.querySelectorAll<HTMLElement>(".hand-slot") ?? []) {
      if (el === slot) continue;
      const sr = el.getBoundingClientRect();
      slotCenters.push(sr.left + sr.width / 2);
    }
    const cardEl = (slot.querySelector(".card") as HTMLElement | null) ?? slot;
    const r = cardEl.getBoundingClientRect();
    grabX = ev.clientX - r.left;
    grabY = ev.clientY - r.top;
    dragSlot = slot;
    dragPointerID = ev.pointerId ?? null;
    dragCardID = c.instance_id;
    ghost = {
      card: c,
      x: r.left,
      y: r.top,
      w: cardEl.offsetWidth || r.width,
      h: cardEl.offsetHeight || r.height,
      originX: r.left,
      originY: r.top,
      snapping: false,
    };
    apply({
      type: "down",
      cardID: c.instance_id,
      x: ev.clientX,
      y: ev.clientY,
      handTop,
      handBottom: handEl ? handRect.bottom : undefined,
      slotCenters,
      fromIndex: index,
    });
    listen();
  }

  function onWindowMove(ev: PointerEvent): void {
    if (!samePointer(ev)) return;
    const was = drag.phase;
    const wasMode = drag.mode;
    apply({ type: "move", x: ev.clientX, y: ev.clientY });
    if (drag.phase !== "dragging") return;
    ev.preventDefault();
    if (was === "pressed") startDrag();
    if (was === "pressed" || wasMode !== drag.mode) syncCastFeedback();
    if (ghost) ghost = { ...ghost, x: ev.clientX - grabX, y: ev.clientY - grabY };
  }

  function onWindowUp(ev: PointerEvent): void {
    if (!samePointer(ev)) return;
    const card = liveDragCard;
    const out = apply({ type: "up", x: ev.clientX, y: ev.clientY, verdict });
    finish(out, card, ev.clientX, ev.clientY);
  }

  function onWindowCancel(): void {
    finish(apply({ type: "cancel" }), null, 0, 0);
  }

  function onWindowKey(ev: KeyboardEvent): void {
    if (ev.key !== "Escape") return;
    if (drag.phase === "dragging") {
      // The Escape belongs to the drag, not to whatever else on the
      // page listens for it (a modal, the targeting banner).
      ev.preventDefault();
      ev.stopPropagation();
    }
    onWindowCancel();
  }

  function apply(input: DragInput): DragOutcome {
    const res = stepDrag(drag, input);
    drag = res.state;
    return res.outcome;
  }

  // startDrag runs once, as a press becomes a drag: capture the pointer
  // so the gesture keeps its events wherever it goes.
  function startDrag(): void {
    if (dragSlot && dragPointerID !== null) {
      try {
        dragSlot.setPointerCapture?.(dragPointerID);
      } catch {
        // A pointer that is already gone cannot be captured; the
        // window listeners still see its events.
      }
    }
  }

  // syncCastFeedback runs when the drag starts and whenever it changes
  // band. Inside the hand's band it is a reorder, and the board shows
  // no auto-tap highlight; out of it, the auto-tapper is asked which
  // sources it would spend — once per drag, never per move — and its
  // answer drives the highlight.
  function syncCastFeedback(): void {
    if (drag.mode === "reorder") {
      clearAutoTapHighlight();
      return;
    }
    if (preview) {
      autoTapHighlight.set(new Set(previewSourceIDs(preview)));
      return;
    }
    if (previewRequested) return;
    previewRequested = true;
    const card = liveDragCard;
    if (!card || !snap?.id || !wantsPreview(card, legalityFor(card))) return;
    const token = ++previewToken;
    fetchAutoTapPreview(snap.id, card.instance_id)
      .then((p) => {
        if (token !== previewToken || drag.phase !== "dragging") return;
        preview = p;
        if (drag.mode !== "reorder") autoTapHighlight.set(new Set(previewSourceIDs(p)));
      })
      .catch(() => {
        // No preview, no highlight: the drag works without one, and the
        // server is the real answer on release.
      });
  }

  function finish(out: DragOutcome, card: CardView | null, x: number, y: number): void {
    unlisten();
    if (dragSlot && dragPointerID !== null) {
      try {
        dragSlot.releasePointerCapture?.(dragPointerID);
      } catch {
        // Already released.
      }
    }
    dragSlot = null;
    dragPointerID = null;
    dragCardID = null;
    previewToken++;
    previewRequested = false;
    preview = null;
    clearAutoTapHighlight();

    if (out.kind === "click" || out.kind === "none") {
      ghost = null;
      return;
    }
    swallowClick = true;
    if (out.kind === "cast") {
      ghost = null;
      if (card) onDragCast?.(card);
      return;
    }
    if (out.kind === "reorder") {
      ghost = null;
      // Re-find the card in the live hand: a snapshot may have landed
      // mid-drag. A card that left the hand has nothing to move.
      const ids = idsOf(cards);
      const from = ids.indexOf(out.cardID);
      if (from >= 0) setOrder(moveCard(ids, from, out.to));
      return;
    }
    if (out.reason) showSnapReason(out.reason, x, y);
    snapBack();
  }

  // snapBack returns the ghost to the card's place in the hand, or just
  // removes it under reduced motion.
  function snapBack(): void {
    const g = ghost;
    if (!g || $settings.accessibility.reduceMotion) {
      ghost = null;
      return;
    }
    ghost = { ...g, snapping: true, x: g.originX, y: g.originY };
    if (snapTimer !== null) clearTimeout(snapTimer);
    snapTimer = setTimeout(() => {
      snapTimer = null;
      ghost = null;
    }, 180);
  }

  function showSnapReason(text: string, x: number, y: number): void {
    snapReason = { text, x, y };
    if (snapReasonTimer !== null) clearTimeout(snapReasonTimer);
    snapReasonTimer = setTimeout(() => {
      snapReasonTimer = null;
      snapReason = null;
    }, SNAP_REASON_MS);
  }

  function onSlotClickCapture(ev: MouseEvent): void {
    if (!swallowClick) return;
    swallowClick = false;
    ev.preventDefault();
    ev.stopPropagation();
  }

  // Unmounted mid-drag (leaving the table): drop the listeners and the
  // board highlight so neither outlives the hand.
  $effect(() => {
    return () => {
      unlisten();
      clearAutoTapHighlight();
      if (snapTimer !== null) clearTimeout(snapTimer);
      if (snapReasonTimer !== null) clearTimeout(snapReasonTimer);
    };
  });
</script>

{#if canSort}
  <!-- #1524: the hand menu — one-time sorts. The result becomes the
       saved order; nothing keeps the hand sorted afterwards. -->
  <div class="hand-sort">
    <button
      bind:this={sortButton}
      type="button"
      class="hand-sort-button"
      aria-label="Sort hand"
      title="Sort hand"
      aria-haspopup="menu"
      aria-expanded={sortMenuOpen}
      onclick={() => (sortMenuOpen ? closeSortMenu(false) : openSortMenu())}
    >
      <svg viewBox="0 0 16 16" width="14" height="14" aria-hidden="true">
        <path
          d="M2 3.5h12M2 8h8M2 12.5h4"
          stroke="currentColor"
          stroke-width="1.6"
          stroke-linecap="round"
          fill="none"
        />
      </svg>
    </button>
    {#if sortMenuOpen}
      <div
        bind:this={sortMenu}
        class="hand-sort-menu"
        role="menu"
        aria-label="Sort hand by"
        tabindex="-1"
        onkeydown={onSortMenuKey}
      >
        {#each HAND_SORTS as s (s.key)}
          <button type="button" role="menuitem" onclick={() => sortBy(s.key)}>{s.label}</button>
        {/each}
      </div>
    {/if}
  </div>
{/if}
<div
  class="hand"
  class:opponent={!isSelf}
  class:stacked={layout === "stacked"}
  class:reordering
  style:--hand-overlap={overlap}
  aria-label={isSelf ? "your hand" : "opponent hand"}
>
  {#each cards as c, i (c.instance_id)}
    {@const leg = legalityFor(c)}
    {@const shift = gapShift(i, c)}
    {@const gapX = shift === 0 ? "" : `translateX(calc(var(--card-w, 80px) * ${shift * 0.3})) `}
    <!-- #1508: the pointer handler is the drag-to-cast gesture, a
         pointer-only enhancement. The Card inside is the button, and
         clicking or pressing Enter on it casts exactly as before. -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="hand-slot"
      class:timing-disabled={isSelf && !leg.legal}
      class:draggable={dragEnabled}
      class:drag-source={dragging && dragCardID === c.instance_id}
      class:gap-left={shift < 0}
      class:gap-right={shift > 0}
      title={isSelf && !leg.legal ? leg.reason : undefined}
      onpointerdown={dragEnabled ? (ev) => onSlotPointerDown(ev, c, i) : undefined}
      onclickcapture={dragEnabled ? onSlotClickCapture : undefined}
      style:transform={layout === "stacked"
        ? gapX || "none"
        : `${gapX}rotate(${fanAngle(i, cards.length)}deg) translateY(${fanLift(i, cards.length)}px)`}
    >
      <!-- Inner wrapper carries the deal-in / deal-out transforms so
           they don't fight the .hand-slot's fan-layout transform. -->
      <div class="deal-wrap" in:dealIn out:dealOut use:handLifecycle>
        <Card
          card={c}
          faceDown={!isSelf && c.known_by_you !== true}
          showManaCost={isSelf}
          priority={isSelf}
          onActivateAbility={isSelf && (c.zone_abilities?.length ?? 0) > 0
            ? (idx) => onActivateAbility?.(c, idx)
            : undefined}
          onActivateManaAbility={isSelf && (c.zone_mana_abilities?.length ?? 0) > 0
            ? (idx) => onActivateManaAbility?.(c, idx)
            : undefined}
          {sorcerySpeedBlocked}
          onClick={isSelf && leg.legal ? () => handleCardClick(c) : undefined}
        />
      </div>
    </div>
  {/each}
  {#if !isSelf && hand.count === 0}
    <span class="empty">empty</span>
  {/if}
</div>

{#if dragging && drag.mode === "cast"}
  <!-- #1508: the faint "release to cast" zone over the table, shown
       once the dragged card has crossed the line. -->
  <div
    class="drag-cast-zone"
    class:castable={verdict.castable}
    class:blocked={!verdict.castable}
    style:height="{Math.max(0, drag.handTop - CAST_ZONE_MARGIN_PX - 12)}px"
    use:portal
    aria-hidden="true"
  >
    <span class="drag-cast-label">
      {verdict.castable ? "Release to cast" : verdict.reason}
    </span>
  </div>
{/if}
{#if ghost && (dragging || ghost.snapping)}
  {@const art = cardImageURL(ghost.card, "small")}
  <div
    class="drag-ghost"
    class:castable={dragging && !reordering && verdict.castable}
    class:blocked={dragging && !reordering && !verdict.castable}
    class:reordering
    class:snapping={ghost.snapping}
    style:left="{ghost.x}px"
    style:top="{ghost.y}px"
    style:width="{ghost.w}px"
    style:height="{ghost.h}px"
    use:portal
    aria-hidden="true"
  >
    {#if art}
      <img src={art} alt="" draggable="false" />
    {:else}
      <span class="drag-ghost-name">{ghost.card.name}</span>
    {/if}
    {#if dragging && !reordering && !verdict.castable && verdict.reason}
      <span class="drag-ghost-reason">{verdict.reason}</span>
    {/if}
  </div>
{/if}
{#if snapReason}
  <div
    class="drag-snap-reason"
    role="status"
    style:left="{snapReason.x}px"
    style:top="{snapReason.y}px"
    use:portal
  >
    {snapReason.text}
  </div>
{/if}

<style>
  .hand {
    display: flex;
    flex-direction: row;
    align-items: flex-start;
    justify-content: center;
    min-height: 0;
    padding: 4px;
    /* Default "peek" state: clip to the top 62% of a card so the hand
       takes up well under its full vertical footprint while exposing
       the name, mana cost, art, and type line and hiding P/T +
       flavour / rules text. Hover lifts the whole fan up and over the
       board (see .hand:hover) to reveal full cards without pushing
       layout. PlayerPanel's .hand-zone reserves the same 62%. */
    max-height: calc(var(--card-h, 168px) * 0.62);
    /* #956 — never wider than the zone that holds it. The fan's own
       width is bounded by --hand-overlap below; this is the backstop
       for the case where it is not. */
    max-width: 100%;
    overflow: hidden;
    position: relative;
    z-index: 1;
    transition:
      max-height 220ms var(--ease),
      transform 220ms var(--ease);
  }
  /* Self hand expands on hover: overflow goes visible, the whole strip
     translates upward by the hidden 38% so full cards poke over the
     battlefield and the fan's bottom edge stays on the panel edge,
     and z-index jumps so nothing on the board occludes the revealed
     cards. */
  .hand:not(.opponent):hover {
    max-height: none;
    overflow: visible;
    transform: translateY(calc(var(--card-h, 168px) * -0.38));
    z-index: 20;
  }
  /* Opponent hands stay compact — they're face-down anyway and the
     peek/reveal interaction would feel wrong on someone else's hand. */
  .hand.opponent {
    max-height: calc(var(--card-h, 168px) * 0.55);
    overflow: hidden;
  }
  /* #956 — the overlap is handOverlap()'s answer, published by the
     markup as --hand-overlap, rather than three hard-coded per-layout
     values. A constant let the fan's width grow linearly with the
     card count (1 + (n-1)/2 card-widths for the self hand: 4 wide at
     seven cards, 7.5 at fourteen) until .panel's overflow: hidden cut
     it off — and --card-h's 168px floor means the cards do not shrink
     to fit. The resting values are unchanged: 0.5 self, 0.62 for an
     opponent's face-down fan, 0.85 stacked; only large hands tighten.
     The fallback matches the self hand's resting value.

     The overlap is sized relative to the card width (--card-w, which
     cascades from the card-size setting) rather than the container,
     because a % margin-left in flexbox resolves against the flex
     container, pushing cards off-screen. */
  .hand-slot {
    margin-left: calc(var(--card-w, 80px) * -1 * var(--hand-overlap, 0.5));
    transform-origin: bottom center;
    transition: transform 80ms ease;
  }
  .hand-slot:first-child {
    margin-left: 0;
  }
  /* Stacked layout needs left alignment — justify-content: center
     with near-zero total width collapses every slot onto the same
     point, which hides every card but the last. flex-start anchors
     the stack to the left edge so the strip grows visibly. */
  .hand.stacked {
    justify-content: flex-start;
    padding-left: 12px;
  }
  /* Opponent hand sizes are inherited from the parent
     .panel.opponent via --card-w/--card-h, so no explicit override
     here. Their resting fan is tighter than the self hand's (0.62 vs
     0.5, see overlapBase in the script) because face-down stacks read
     better tightly packed. */
  .empty {
    font-size: 11px;
    color: var(--fg-dim);
    align-self: center;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    font-weight: 600;
  }
  /* S13.3 — illegal-cast greying. Hover/zoom still works; only the
     click affordance is muted via the conditional onClick on Card. */
  .hand-slot.timing-disabled {
    opacity: 0.5;
    filter: grayscale(0.5) brightness(0.85);
  }

  /* #1508: drag to cast. A self hand card takes the whole gesture —
     nothing in the hand scrolls, and a vertical drag handed to the
     browser as a pan would cancel the pointer mid-drag. Taps still
     click. */
  .hand-slot.draggable {
    touch-action: none;
  }
  /* The card being dragged keeps its place, dimmed, so the fan does
     not reflow under the pointer and the snap-back has somewhere to
     land. */
  .hand-slot.drag-source {
    opacity: 0.3;
  }
  /* #1524: while reordering, the other cards slide apart to show where
     the dragged one will land (gapShift in the script). A slightly
     longer ease than the resting 80ms so the gap reads as opening. */
  .hand.reordering .hand-slot {
    transition: transform 140ms var(--ease, ease);
  }
  /* The sort menu sits in the hand zone's top-left corner, faint until
     it is wanted. PlayerPanel's .hand-zone is position: relative. */
  .hand-sort {
    position: absolute;
    top: 2px;
    left: 2px;
    z-index: 21;
  }
  .hand-sort-button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    padding: 0;
    border: 1px solid var(--border, rgba(255, 255, 255, 0.2));
    border-radius: 6px;
    background: var(--surface-2, #222);
    color: var(--fg, #eee);
    opacity: 0.55;
    cursor: pointer;
  }
  .hand-sort-button:hover,
  .hand-sort-button:focus-visible,
  .hand-sort-button[aria-expanded="true"] {
    opacity: 1;
  }
  .hand-sort-menu {
    position: absolute;
    bottom: calc(100% + 4px);
    left: 0;
    z-index: 30;
    display: flex;
    flex-direction: column;
    min-width: 170px;
    padding: 4px;
    border: 1px solid var(--border, rgba(255, 255, 255, 0.2));
    border-radius: 8px;
    background: var(--surface-2, #222);
    box-shadow: 0 8px 22px rgba(0, 0, 0, 0.5);
  }
  .hand-sort-menu button {
    padding: 6px 10px;
    border: 0;
    border-radius: 5px;
    background: transparent;
    color: var(--fg, #eee);
    font-size: 12px;
    text-align: left;
    white-space: nowrap;
    cursor: pointer;
  }
  .hand-sort-menu button:hover,
  .hand-sort-menu button:focus-visible {
    background: var(--accent-soft, rgba(255, 212, 0, 0.18));
    outline: none;
  }
  /* A reordering ghost is neither gold nor red: nothing is being cast. */
  .drag-ghost.reordering {
    transform: scale(1.04);
    box-shadow:
      0 0 0 2px rgba(255, 255, 255, 0.35),
      0 14px 30px rgba(0, 0, 0, 0.55);
  }
  /* The ghost, the zone and the reason are portalled to <body>. Scoped
     rules still reach them: the scoping class is on the element, and
     moving it does not take it off. */
  .drag-ghost {
    position: fixed;
    z-index: 1000;
    pointer-events: none;
    border-radius: 8px;
    transform: rotate(-3deg) scale(1.04);
    box-shadow: 0 14px 30px rgba(0, 0, 0, 0.55);
  }
  .drag-ghost img {
    display: block;
    width: 100%;
    height: 100%;
    border-radius: 8px;
    object-fit: cover;
  }
  .drag-ghost-name {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    padding: 6px;
    box-sizing: border-box;
    border-radius: 8px;
    background: var(--surface-2, #222);
    color: var(--fg, #eee);
    font-size: 12px;
    text-align: center;
  }
  .drag-ghost.castable {
    box-shadow:
      0 0 0 2px var(--gold),
      0 0 26px rgba(255, 208, 122, 0.75),
      0 14px 30px rgba(0, 0, 0, 0.55);
  }
  .drag-ghost.blocked {
    box-shadow:
      0 0 0 2px var(--danger),
      0 0 22px rgba(255, 107, 107, 0.6),
      0 14px 30px rgba(0, 0, 0, 0.55);
  }
  .drag-ghost.blocked img {
    filter: grayscale(0.5) brightness(0.65);
  }
  .drag-ghost-reason {
    position: absolute;
    left: 50%;
    bottom: -8px;
    transform: translate(-50%, 100%);
    white-space: nowrap;
    padding: 3px 8px;
    border-radius: 6px;
    background: rgba(40, 10, 10, 0.92);
    color: #ffd6d6;
    font-size: 12px;
    font-weight: 600;
  }
  .drag-ghost.snapping {
    transition:
      left 180ms var(--ease, ease),
      top 180ms var(--ease, ease),
      opacity 180ms var(--ease, ease);
    opacity: 0.4;
  }
  .drag-cast-zone {
    position: fixed;
    left: 12px;
    right: 12px;
    top: 12px;
    z-index: 999;
    pointer-events: none;
    display: flex;
    align-items: flex-end;
    justify-content: center;
    padding-bottom: 14px;
    box-sizing: border-box;
    border: 2px dashed rgba(217, 180, 92, 0.35);
    border-radius: 14px;
    background: rgba(217, 180, 92, 0.05);
  }
  .drag-cast-zone.blocked {
    border-color: rgba(255, 107, 107, 0.35);
    background: rgba(255, 107, 107, 0.05);
  }
  .drag-cast-label {
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--gold-strong, #f1d38a);
  }
  .drag-cast-zone.blocked .drag-cast-label {
    color: #ffb3b3;
  }
  .drag-snap-reason {
    position: fixed;
    z-index: 1001;
    pointer-events: none;
    transform: translate(-50%, -140%);
    padding: 4px 10px;
    border-radius: 6px;
    background: rgba(40, 10, 10, 0.94);
    color: #ffd6d6;
    font-size: 12px;
    font-weight: 600;
    white-space: nowrap;
  }
  @media (prefers-reduced-motion: reduce) {
    .hand.reordering .hand-slot {
      transition: none;
    }
    .drag-ghost,
    .drag-ghost.snapping {
      transition: none;
      transform: none;
    }
  }
</style>

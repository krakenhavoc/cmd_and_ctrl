<script lang="ts">
  // ExileStrip — the castable-from-exile strip beside the viewer's hand
  // (#1389). Every exiled card the viewer holds a permission over sits
  // here like a second hand: same card size, the same peek-and-lift,
  // the same hover zoom (Card writes the shared hover store), and a
  // click hands the card to the Board's one cast chain exactly as the
  // zone browser's exile button does (#874) — X, modes, targets, costs.
  //
  // What goes in, in what state, and what the corner badge says are
  // all decided in exileStrip.ts from server answers; this file only
  // lays them out.
  //
  //   - castable now: lit, clickable.
  //   - waiting on its window (a plotted card outside its main phase,
  //     a sorcery in an end step): greyed like an uncastable hand card.
  //   - waiting on a later turn (warp, plot, foretell on the turn they
  //     were set up): dimmed, with a "next turn" caption.
  //
  // The badge is a price tag on the top-right corner, where the card
  // prints its own cost, and it only appears when the price is NOT the
  // printed cost: 0 for a plotted card, 2 for airbend, the foretell
  // cost. Hovering it lists the printed cost and any other price.
  //
  // #1622: a card may also be DRAGGED onto the table to cast it, the
  // hand's gesture (#1508) with the same thresholds, ghost, gold / red
  // verdict and auto-tap preview; every decision is dragCast.ts's. The
  // drop hands the card to the same cast chain a click does, with the
  // exile zone, the grant's face and the drag flag (strict + auto-tap).
  // The strip has no reorder band. Every card here is one the viewer
  // holds a permission over; a card they do not hold is not in the
  // strip at all, so it cannot be picked up.
  //
  // On a narrow panel the strip collapses to a count chip that opens
  // the cards in a popover, so a phone keeps its hand width.
  //
  // #2202: the strip is the "castable from other zones" strip now. The
  // viewer's commanders sit in it too, first, nearest the hand: lit
  // when castable, greyed when not, with a price tag once the commander
  // tax (or a cost modifier) moves the price off the printed cost. A
  // click or a drag hands the commander to the same cast chain with
  // the command zone as its zone, exactly as the command zone panel's
  // own click does. castStrip.ts decides which cards and what tag; the
  // command zone panel keeps showing the commander, since it lives
  // there.

  import type { CardView, GameView } from "../../protocol";
  import Card from "./Card.svelte";
  import ManaSymbol from "./ManaSymbol.svelte";
  import { pipRun } from "../../manaSymbol";
  import { dealIn, dealOut } from "../../animations";
  import { handOverlap } from "../../handFan";
  import { settings } from "../../settings";
  import { fetchAutoTapPreview, type AutoTapPreview } from "../../api";
  import { cardImageURL } from "../../cardImage";
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
  import {
    castStripBadge,
    castStripCastAnywayBlocked,
    castStripEntries,
    castStripLegality,
    castStripOffersCastAnyway,
    type CastStripEntry,
    type CastStripZone,
  } from "../../castStrip";
  import { requestCastAnyway } from "../../castAnyway";
  import type { CastSourceZone } from "../../targeting";
  import type { Legality } from "../../timing";
  import { NO_LEGAL_ACTIONS, type LegalActions } from "../../legalActions";

  interface Props {
    view: GameView;
    viewerID: string | null;
    onCastCard?: (card: CardView, fromZone: CastSourceZone, face?: number) => void;
    // #1622: drag a card onto the table to cast it. Called once the
    // card is released past the line while castable; the parent runs
    // the same cast chain a click does, with the drag flag. Undefined
    // turns the gesture off.
    onDragCast?: (card: CardView, fromZone: CastSourceZone, face?: number) => void;
    // ADR 0105 (#1789): the frame's legal-action lookup, "nothing"
    // while highlights are off or autopass is about to pass. Drives
    // the ready ring and the chip's ready count; the click gate
    // (legalityFor) never reads it.
    legal?: LegalActions;
  }

  const { view, viewerID, onCastCard, onDragCast, legal = NO_LEGAL_ACTIONS }: Props = $props();

  const entries = $derived(castStripEntries(view, viewerID));
  const commanderCount = $derived(entries.filter((e) => e.zone === "command").length);
  const exileCount = $derived(entries.length - commanderCount);
  // ADR 0105: what the server's move list says the viewer can cast
  // from exile this instant, mana included. The chip's cyan count, and
  // (sub-PR 6, §7) the count its accessible name says. It used to say
  // how many cards castable_here opened a window for, which counts a
  // card the viewer cannot pay for. With no digest (highlights off,
  // autopass passing, or no decision owed) the label says no count at
  // all rather than one the board is not drawing.
  //
  // #2202: plus the commanders the move list can cast from the command
  // zone. With no commander in the strip the label reads exactly as it
  // did; with one it names both (labels are a contract: added, never
  // renamed).
  const readyCount = $derived(
    legal.readyCount("exile") +
      entries.filter(
        (e) => e.zone === "command" && legal.castableFrom(e.card.instance_id, "command"),
      ).length,
  );
  const chipLabel = $derived(
    stripCountPhrase(commanderCount, exileCount) +
      " you may cast" +
      (legal.known ? `, ${readyCount} ready` : ""),
  );
  function stripCountPhrase(commanders: number, exiled: number): string {
    const ex = `${exiled} exiled ${exiled === 1 ? "card" : "cards"}`;
    if (commanders === 0) return ex;
    const cmd = `${commanders} ${commanders === 1 ? "commander" : "commanders"}`;
    return exiled === 0 ? cmd : `${cmd} and ${ex}`;
  }
  // The strip's own name and the chip's word: unchanged while it holds
  // only exile cards.
  const stripLabel = $derived(
    commanderCount === 0 ? "castable from exile" : "castable from other zones",
  );
  const chipWord = $derived(commanderCount === 0 ? "exile" : exileCount === 0 ? "cmd" : "cast");
  const tagText = $derived(
    commanderCount === 0 ? "from exile" : exileCount === 0 ? "commander" : "other zones",
  );
  // Two cards sit side by side; from three on they overlap like the
  // hand, tightening with handOverlap so a long strip cannot outgrow
  // the panel (#956's rule).
  const overlap = $derived(entries.length <= 2 ? -0.05 : handOverlap(entries.length, 0.4));

  // Collapsed-mode popover. Irrelevant on a wide panel, where the CSS
  // shows the cards inline whatever this says.
  let open = $state(false);
  $effect(() => {
    if (entries.length === 0) open = false;
  });

  // #1406: extracted to exileStrip.ts as exileEntryLegality, shared
  // with the zone browser's exile button so the two surfaces read the
  // same verdict rather than deriving it a second way.
  // #2202: castStripLegality, which asks a commander the command zone
  // panel's own gate.
  function legalityFor(e: CastStripEntry): Legality {
    return castStripLegality(e, view, viewerID);
  }

  function cast(e: CastStripEntry): void {
    if (!onCastCard) return;
    open = false;
    onCastCard(e.card, e.zone, e.face);
  }

  // ADR 0118 §2: with strict payment on, a commander and an exile entry
  // whose verb is "cast" offer "Cast anyway (don't pay)" in their
  // popover, payable or not. The row opens the dock's confirmation
  // (castAnyway.ts), which starts the same cast chain a click does.
  function castAnywayHere(e: CastStripEntry): boolean {
    return !!onCastCard && $settings.gameplay.strictMana && castStripOffersCastAnyway(e);
  }
  function castAnyway(e: CastStripEntry): void {
    open = false;
    requestCastAnyway(e.card, e.zone, e.face);
  }
  // ---- #1622: drag to cast -------------------------------------------
  let drag = $state<DragState>(IDLE);
  let dragCardID = $state<string | null>(null);
  let dragZone = $state<CastStripZone>("exile");
  let preview = $state<AutoTapPreview | null>(null);
  let previewToken = 0;
  let previewRequested = false;
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
  let grabX = 0;
  let grabY = 0;
  let snapReason = $state<{ text: string; x: number; y: number } | null>(null);
  let snapReasonTimer: ReturnType<typeof setTimeout> | null = null;
  let snapTimer: ReturnType<typeof setTimeout> | null = null;
  let dragSlot: HTMLElement | null = null;
  let dragPointerID: number | null = null;
  // A drag that ended swallows the one click the browser may still
  // deliver for the same press.
  let swallowClick = false;

  const dragEnabled = $derived(!!onDragCast);
  const dragging = $derived(drag.phase === "dragging");
  const liveEntry = $derived(
    dragCardID ? (entries.find((e) => e.card.instance_id === dragCardID) ?? null) : null,
  );
  const verdict = $derived(
    liveEntry
      ? dragVerdict(liveEntry.card, legalityFor(liveEntry), preview)
      : {
          castable: false,
          reason:
            dragZone === "command" ? "That card left the command zone" : "That card left exile",
        },
  );

  // The ghost, the drop zone and the reason are position: fixed, and a
  // transformed ancestor (the strip's hover lift) would make "fixed"
  // mean "fixed to that ancestor", so they are moved to <body>.
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

  function apply(input: DragInput): DragOutcome {
    const res = stepDrag(drag, input);
    drag = res.state;
    return res.outcome;
  }

  function onSlotPointerDown(ev: PointerEvent, e: CastStripEntry): void {
    swallowClick = false;
    if (!dragEnabled || drag.phase !== "idle") return;
    if (ev.button !== 0 || ev.isPrimary === false) return;
    const slot = ev.currentTarget as HTMLElement;
    const row = (slot.closest(".strip-cards") as HTMLElement | null) ?? slot;
    const cardEl = (slot.querySelector(".card") as HTMLElement | null) ?? slot;
    const r = cardEl.getBoundingClientRect();
    grabX = ev.clientX - r.left;
    grabY = ev.clientY - r.top;
    dragSlot = slot;
    dragPointerID = ev.pointerId ?? null;
    dragCardID = e.card.instance_id;
    dragZone = e.zone;
    ghost = {
      card: e.card,
      x: r.left,
      y: r.top,
      w: cardEl.offsetWidth || r.width,
      h: cardEl.offsetHeight || r.height,
      originX: r.left,
      originY: r.top,
      snapping: false,
    };
    // No handBottom: the strip has no reorder band, so the only
    // outcomes are cast and snap back.
    apply({
      type: "down",
      cardID: e.card.instance_id,
      x: ev.clientX,
      y: ev.clientY,
      handTop: row.getBoundingClientRect().top,
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
    const entry = liveEntry;
    const out = apply({ type: "up", x: ev.clientX, y: ev.clientY, verdict });
    finish(out, entry, ev.clientX, ev.clientY);
  }

  function onWindowCancel(): void {
    finish(apply({ type: "cancel" }), null, 0, 0);
  }

  function onWindowKey(ev: KeyboardEvent): void {
    if (ev.key !== "Escape") return;
    if (drag.phase === "dragging") {
      ev.preventDefault();
      ev.stopPropagation();
    }
    onWindowCancel();
  }

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

  // The auto-tapper is asked which sources it would spend once per
  // drag, once the card is past the line, and its answer drives the
  // board highlight.
  function syncCastFeedback(): void {
    if (drag.mode !== "cast") {
      clearAutoTapHighlight();
      return;
    }
    if (preview) {
      autoTapHighlight.set(new Set(previewSourceIDs(preview)));
      return;
    }
    if (previewRequested) return;
    previewRequested = true;
    const entry = liveEntry;
    if (!entry || !view.id || !wantsPreview(entry.card, legalityFor(entry))) return;
    const token = ++previewToken;
    // #2202: priced for the zone the card is cast from — a commander
    // carries its tax only when the preview is told it comes out of
    // the command zone, and an exile grant's face and price likewise.
    fetchAutoTapPreview(view.id, entry.card.instance_id, {
      cast: { fromZone: entry.zone, face: entry.face },
    })
      .then((p) => {
        if (token !== previewToken || drag.phase !== "dragging") return;
        preview = p;
        if (drag.mode === "cast") autoTapHighlight.set(new Set(previewSourceIDs(p)));
      })
      .catch(() => {
        // No preview, no highlight: the server is the real answer.
      });
  }

  function finish(out: DragOutcome, entry: CastStripEntry | null, x: number, y: number): void {
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
      if (entry) {
        open = false;
        onDragCast?.(entry.card, entry.zone, entry.face);
      }
      return;
    }
    if (out.kind === "snapBack" && out.reason) showSnapReason(out.reason, x, y);
    snapBack();
  }

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

  // Unmounted mid-drag: drop the listeners and the board highlight.
  $effect(() => {
    return () => {
      unlisten();
      clearAutoTapHighlight();
      if (snapTimer !== null) clearTimeout(snapTimer);
      if (snapReasonTimer !== null) clearTimeout(snapReasonTimer);
    };
  });
</script>

{#if entries.length > 0}
  <div class="exile-strip" class:open aria-label={stripLabel}>
    <button
      type="button"
      class="strip-toggle"
      aria-expanded={open}
      aria-label={chipLabel}
      onclick={() => (open = !open)}
    >
      <span class="toggle-label">{chipWord}</span>
      <span class="toggle-count">{entries.length}</span>
      {#if readyCount > 0}
        <span class="toggle-count ready" aria-hidden="true">{readyCount} ready</span>
      {/if}
    </button>
    <div class="strip-body" style:--strip-overlap={overlap}>
      <span class="strip-tag" aria-hidden="true">{tagText}</span>
      <div class="strip-cards">
        {#each entries as e (e.card.instance_id)}
          {@const leg = legalityFor(e)}
          {@const badge = castStripBadge(e)}
          <!-- #1622: the pointer handler is the drag-to-cast gesture, a
               pointer-only enhancement. The Card inside is the button;
               click and Enter cast exactly as before. -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div
            class="strip-slot"
            class:draggable={dragEnabled}
            class:drag-source={dragging && dragCardID === e.card.instance_id}
            onpointerdown={dragEnabled ? (ev) => onSlotPointerDown(ev, e) : undefined}
            onclickcapture={dragEnabled ? onSlotClickCapture : undefined}
            class:blocked={!leg.legal}
            class:later={e.state === "later"}
            class:castable={leg.legal}
            title={leg.legal ? undefined : leg.reason}
          >
            <div class="deal-wrap" in:dealIn out:dealOut>
              <Card
                card={e.card}
                showManaCost={badge === null}
                ready={legal.castableFrom(e.card.instance_id, e.zone)}
                readyZone={e.zone}
                {legal}
                onClick={leg.legal && onCastCard ? () => cast(e) : undefined}
                onCastAnyway={castAnywayHere(e) ? () => castAnyway(e) : undefined}
                castAnywayBlocked={castAnywayHere(e)
                  ? castStripCastAnywayBlocked(e, view, viewerID)
                  : ""}
              />
              {#if badge}
                <span class="cost-tag" title={badge.title} aria-label={badge.label}>
                  {#each pipRun(badge.symbols) as p, i (i)}
                    <ManaSymbol symbol={p.symbol} size={15} />
                  {/each}
                  {#if badge.life}
                    <span class="life">+{badge.life}♥</span>
                  {/if}
                </span>
              {/if}
              {#if e.hint}
                <span class="hint">{e.hint}</span>
              {:else if e.verb === "play" && leg.legal}
                <span class="hint verb">play</span>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    </div>
  </div>
{/if}

{#if dragging && drag.mode === "cast"}
  <!-- #1622: the faint "release to cast" zone over the table. -->
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
    class:castable={dragging && verdict.castable}
    class:blocked={dragging && !verdict.castable}
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
    {#if dragging && !verdict.castable && verdict.reason}
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
  .exile-strip {
    position: relative;
    flex: 0 1 auto;
    min-width: 0;
    align-self: flex-end;
    display: flex;
    align-items: flex-end;
    /* Same resting footprint as the self hand beside it (PlayerPanel's
       .hand-zone): 62% of a card, so the two rows line up. */
    height: calc(var(--card-h, 168px) * 0.62);
  }

  /* ---- inline (wide panel) ------------------------------------ */

  .strip-toggle {
    display: none;
  }
  .strip-body {
    position: relative;
    display: flex;
    align-items: flex-end;
    height: 100%;
    /* A gold hairline to the left: this is a second hand, and it
       should read as one — the same cards, set apart. */
    padding: 0 6px 0 12px;
    border-left: 1px dashed rgba(217, 180, 92, 0.45);
  }
  .strip-tag {
    position: absolute;
    left: 3px;
    bottom: 4px;
    writing-mode: vertical-rl;
    transform: rotate(180deg);
    font-family: var(--font-mono);
    font-size: 8.5px;
    font-weight: 700;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--gold);
    opacity: 0.75;
    pointer-events: none;
  }
  .strip-cards {
    display: flex;
    flex-direction: row;
    align-items: flex-start;
    /* Room above the cards for the price tag, which sits on the
       corner rather than inside the art. */
    padding: 7px 7px 0 4px;
    max-height: 100%;
    max-width: 100%;
    overflow: hidden;
    position: relative;
    z-index: 1;
    transition:
      max-height 220ms var(--ease),
      transform 220ms var(--ease);
  }
  /* The hand's lift, exactly: overflow goes visible and the row rises
     by the hidden 38% so whole cards show over the board. */
  .strip-cards:hover {
    max-height: none;
    overflow: visible;
    transform: translateY(calc(var(--card-h, 168px) * -0.38));
    z-index: 20;
  }
  .strip-slot {
    position: relative;
    margin-left: calc(var(--card-w, 80px) * -1 * var(--strip-overlap, 0.42));
    transition: transform 120ms var(--ease);
  }
  .strip-slot:first-child {
    margin-left: 0;
  }
  .strip-slot:last-child .cost-tag {
    right: -5px;
  }
  .strip-slot.castable:hover {
    transform: translateY(-6px);
    z-index: 2;
  }
  .deal-wrap {
    position: relative;
  }
  /* ADR 0105: the faint gold glow a castable card used to wear is the
     board-wide cyan ready ring now (Card's `ready`), drawn from the
     server's move list — mana included — and only while highlights
     are live. */
  /* Dimmed with a filter rather than opacity: the cards overlap, and a
     translucent card lets the one beneath it — its badges included —
     show through, which reads as two cards smeared together. */
  /* On the card alone, so the caption and the tag stay legible. */
  .strip-slot.blocked .deal-wrap :global(.card) {
    filter: grayscale(0.5) brightness(0.6);
  }
  .strip-slot.later .deal-wrap :global(.card) {
    filter: grayscale(0.85) brightness(0.45);
  }

  /* ---- the price tag ------------------------------------------ */

  .cost-tag {
    position: absolute;
    top: -6px;
    /* The top-right of the part of the card you can SEE: the next card
       in the row covers this one's right-hand share, and a tag on the
       true corner would sit on top of its neighbour and read as the
       neighbour's price. The last card has no neighbour. */
    right: calc(var(--card-w, 80px) * max(var(--strip-overlap, 0), 0) - 5px);
    z-index: 6;
    display: inline-flex;
    align-items: center;
    gap: 1px;
    padding: 2px 3px;
    border-radius: 999px;
    /* The table's one accent (app.css): the tag is the thing on the
       card to read, and gold is how this board says so. */
    background: #1c1503;
    border: 1.5px solid var(--gold-strong);
    box-shadow:
      0 2px 8px rgba(0, 0, 0, 0.55),
      0 0 0 1px rgba(0, 0, 0, 0.4);
    cursor: help;
  }
  .life {
    margin-left: 2px;
    font-size: 9px;
    font-weight: 700;
    color: #ffb3b3;
  }

  /* ---- captions ----------------------------------------------- */

  .hint {
    position: absolute;
    left: 50%;
    /* Inside the visible 62% peek, so it reads at rest. */
    top: calc(var(--card-h, 168px) * 0.62 - 28px);
    transform: translateX(-50%);
    z-index: 5;
    padding: 2px 7px;
    border-radius: 999px;
    white-space: nowrap;
    font-family: var(--font-mono);
    font-size: 8.5px;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: #b9d8ff;
    background: rgba(12, 22, 44, 0.92);
    border: 1px solid rgba(145, 195, 255, 0.55);
    pointer-events: none;
  }
  .hint.verb {
    color: var(--gold-strong);
    background: rgba(40, 30, 6, 0.92);
    border-color: rgba(217, 180, 92, 0.6);
  }

  /* ---- collapsed (narrow panel) -------------------------------
     .panel is a size container (PlayerPanel.svelte), so this asks
     how wide the viewer's own panel is, not the window. Below 640px
     the strip gives its width back to the hand and becomes a chip;
     the cards open in a popover above it that scrolls sideways inside
     itself, never the page. */
  @container (max-width: 640px) {
    .exile-strip {
      /* Static, so the popover below positions against .panel (which
         is position: relative) and can span the panel's width rather
         than hang off a chip near its right edge. */
      position: static;
      height: auto;
      align-self: flex-end;
    }
    .strip-toggle {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      height: 30px;
      padding: 0 8px 0 10px;
      border-radius: 999px;
      border: 1px solid rgba(217, 180, 92, 0.55);
      background: var(--surface-raised);
      color: var(--gold-strong);
      font-family: var(--font-mono);
      font-size: 10px;
      font-weight: 700;
      letter-spacing: 0.12em;
      text-transform: uppercase;
    }
    .toggle-count {
      display: inline-grid;
      place-items: center;
      min-width: 18px;
      height: 18px;
      border-radius: 999px;
      background: var(--surface-sunken, rgba(0, 0, 0, 0.35));
      color: var(--fg-muted);
      letter-spacing: 0;
    }
    .toggle-count.ready {
      padding: 0 6px;
      background: var(--ready);
      color: var(--ready-ink);
      letter-spacing: 0.04em;
    }
    .strip-body {
      display: none;
      position: absolute;
      left: 8px;
      right: 8px;
      bottom: calc(var(--card-h, 168px) * 0.62 + 16px);
      z-index: 40;
      height: auto;
      padding: 10px 10px 8px;
      border: 1px solid var(--border-strong);
      border-radius: 12px;
      background: var(--surface);
      box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
      /* Sideways scroll INSIDE the popover, never the page. */
      overflow-x: auto;
      overflow-y: hidden;
    }
    /* A phone-sized card: the panel's --card-h is sized for the board
       and would leave the popover room for one card. Set on the row
       rather than the popover so the popover's own offset above still
       reads the panel's value. */
    .strip-cards {
      --card-h: 154px;
      --card-w: 110px;
    }
    .exile-strip.open .strip-body {
      display: block;
    }
    .strip-tag {
      display: none;
    }
    .strip-cards,
    .strip-cards:hover {
      width: max-content;
      max-height: none;
      overflow: visible;
      transform: none;
    }
    .strip-slot {
      margin-left: 6px;
    }
    /* Side by side here, so the true corner is the visible one. */
    .cost-tag {
      right: -5px;
    }
    .hint {
      top: auto;
      bottom: 6px;
    }
  }
  .strip-slot.draggable {
    touch-action: none;
  }
  .strip-slot.drag-source {
    opacity: 0.3;
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
    .drag-ghost,
    .drag-ghost.snapping {
      transition: none;
      transform: none;
    }
  }
</style>

<script lang="ts">
  // TargetingArrows — the source glow and the arrows drawn while the
  // viewer chooses targets (ADR 0119 §4). The decisions are in
  // lib/targetingArrows.ts; this file finds the elements, measures
  // them against the board, and draws.
  //
  // A board-sized SVG at z 36, beside CombatArrows' layer and like it
  // with no pointer events, so it never takes a click meant for a
  // target. Every lookup goes through boardAnchor, so while a seat's
  // board is expanded over the table (ADR 0120 §3) an arrow ends on
  // the overlay's copy of a card or avatar.
  //
  // The source glows through `data-targeting-source`, set on its
  // element while the prompt is open. The glow is a drop-shadow filter
  // on the element itself, so it follows the hand fan's tilt and a
  // tapped card's turn, and it leaves the card's own outline and
  // box-shadow rings (ready, selected, a stack target) as they are.
  //
  // Re-measured when the prompt or the snapshot changes, on board
  // resize, and on every animation frame in which a mouse or pen moved
  // while a pick is open (the follow arrow).

  import { onDestroy, untrack } from "svelte";
  import type { GameView } from "../../protocol";
  import {
    allPicks,
    isLegalCardTarget,
    isLegalPlayerTarget,
    isPicked,
    targeting,
    type TargetingState,
  } from "../../targeting";
  import { findAnchor, findCardAnchor, findSeatAnchor, hasSize } from "../../boardAnchor";
  import { boardExpandLayout } from "../../boardExpand";
  import { curvePath, relativeTo, type Box, type Point } from "../../stackArrows";
  import { L } from "../../labels";
  import {
    FOLLOW_ARROW_COLOR,
    FOLLOW_SNAP_COLOR,
    PICK_ARROW_COLOR,
    followsPointer,
    pickOpen,
    planPickArrows,
    planTargetingArrows,
    type PickArrowPlan,
    type PickZones,
    type TargetingArrow,
  } from "../../targetingArrows";

  interface Props {
    view: GameView;
    boardEl: HTMLElement | null;
  }

  const { view, boardEl }: Props = $props();

  /** The attribute the source element carries while the prompt is open. */
  const SOURCE_ATTR = "data-targeting-source";

  const uid = `targeting-${Math.random().toString(36).slice(2, 9)}`;

  let arrows = $state<TargetingArrow[]>([]);
  let sourceEl: HTMLElement | null = null;

  // The last pointer position, in client coordinates, and its type.
  // Not reactive: a move schedules a measure, it does not re-render.
  let pointer: { x: number; y: number; type: string } | null = null;

  const zones = $derived<PickZones>({
    battlefield: new Set(view.battlefield.cards.map((c) => c.instance_id)),
    stack: new Set((view.stack_items ?? []).map((it) => it.id)),
    players: new Set(view.seats.map((s) => s.id)),
  });

  function cssEscape(s: string): string {
    return typeof CSS !== "undefined" && typeof CSS.escape === "function"
      ? CSS.escape(s)
      : s.replace(/["\\]/g, "\\$&");
  }

  const stackItemSelector = (id: string) => `[data-stack-item-id="${cssEscape(id)}"]`;
  const sized = { accept: hasSize };

  function findSource(board: HTMLElement, t: TargetingState): HTMLElement | null {
    const id = t.card.instance_id;
    if (!id) return null;
    return findCardAnchor(board, id, sized) ?? findAnchor(board, stackItemSelector(id), sized);
  }

  function findPick(board: HTMLElement, plan: PickArrowPlan): HTMLElement | null {
    switch (plan.kind) {
      case "permanent":
        return findCardAnchor(board, plan.targetID, sized);
      case "player":
        return findSeatAnchor(board, plan.targetID, sized);
      case "stack":
        return findAnchor(board, stackItemSelector(plan.targetID), sized);
    }
  }

  /** The dock's question line ("Select target for X"), or the dock itself. */
  function findDock(board: HTMLElement): HTMLElement | null {
    const doc = board.ownerDocument;
    const dock = doc.querySelector<HTMLElement>(`[aria-label="${L.actions}"]`);
    if (!dock) return null;
    const q = dock.querySelector<HTMLElement>(".dock-question");
    if (q && hasSize(q)) return q;
    return hasSize(dock) ? dock : null;
  }

  /** The legal target under the pointer, if any: its element. */
  function snapTarget(t: TargetingState, x: number, y: number): HTMLElement | null {
    const doc = boardEl?.ownerDocument;
    if (!doc || typeof doc.elementFromPoint !== "function") return null;
    const hit = doc.elementFromPoint(x, y);
    const el = hit?.closest<HTMLElement>(
      "[data-instance-id], [data-seat-id], [data-stack-item-id]",
    );
    if (!el) return null;
    const seatID = el.dataset.seatId;
    if (seatID !== undefined) {
      return zones.players.has(seatID) && isLegalPlayerTarget(t, seatID) && !isPicked(t, seatID)
        ? el
        : null;
    }
    const id = el.dataset.instanceId ?? el.dataset.stackItemId;
    if (!id || !(zones.battlefield.has(id) || zones.stack.has(id))) return null;
    return isLegalCardTarget(t, id) && !isPicked(t, id) ? el : null;
  }

  function setSource(el: HTMLElement | null): void {
    if (sourceEl === el) return;
    sourceEl?.removeAttribute(SOURCE_ATTR);
    el?.setAttribute(SOURCE_ATTR, "");
    sourceEl = el;
  }

  function clear(): void {
    setSource(null);
    if (arrows.length > 0) arrows = [];
  }

  function measure(): void {
    const board = boardEl;
    const t = $targeting;
    if (!board || !t) {
      clear();
      return;
    }
    const b = board.getBoundingClientRect();
    const box = (el: Element): Box => relativeTo(el.getBoundingClientRect(), b);

    const src = findSource(board, t);
    setSource(src);
    const source = src ? box(src) : null;
    const dockEl = src ? null : findDock(board);
    const dock = dockEl ? box(dockEl) : null;

    const picks = planPickArrows(allPicks(t), zones).map((plan) => {
      const el = findPick(board, plan);
      return { plan, box: el ? box(el) : null, isSource: el !== null && el === src };
    });

    let follow: { pointer: Point; snap: Box | null } | null = null;
    const p = pointer;
    if (p && followsPointer(p.type) && pickOpen(t)) {
      const inBoard = p.x >= b.left && p.x <= b.right && p.y >= b.top && p.y <= b.bottom;
      if (inBoard) {
        const snapEl = snapTarget(t, p.x, p.y);
        follow = {
          pointer: { x: p.x - b.left, y: p.y - b.top },
          snap: snapEl ? box(snapEl) : null,
        };
      }
    }

    const plan = planTargetingArrows({ board: b, source, dock, picks, follow });
    arrows = plan.arrows;
  }

  const raf: (cb: () => void) => number =
    typeof requestAnimationFrame === "function"
      ? (cb) => requestAnimationFrame(cb)
      : (cb) => setTimeout(cb, 16) as unknown as number;
  const cancelRaf: (h: number) => void =
    typeof cancelAnimationFrame === "function"
      ? (h) => cancelAnimationFrame(h)
      : (h) => clearTimeout(h);
  let pending = 0;
  let scheduled = false;
  function schedule(): void {
    if (scheduled) return;
    scheduled = true;
    pending = raf(() => {
      scheduled = false;
      untrack(measure);
    });
  }

  // The prompt, the snapshot or the board changed: measure now (the
  // DOM is current when effects run) and again next frame, for a card
  // still laying out.
  // ADR 0120 §3: and when the expanded board opens, closes, changes
  // seat or resizes, because its copy of a card is the anchor then.
  $effect(() => {
    void $targeting;
    void view;
    void $boardExpandLayout;
    const board = boardEl;
    untrack(measure);
    if (!board || !$targeting) return;
    schedule();
    if (typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(schedule);
    ro.observe(board);
    return () => ro.disconnect();
  });

  // The pointer, only while a prompt is open. A move is throttled to
  // one measure per animation frame; touch records its type, so a
  // touch tap after a mouse move drops the follow arrow.
  $effect(() => {
    if (!$targeting || typeof window === "undefined") return;
    const onMove = (e: PointerEvent) => {
      pointer = { x: e.clientX, y: e.clientY, type: e.pointerType };
      schedule();
    };
    const onOut = (e: PointerEvent) => {
      if (e.relatedTarget !== null) return;
      pointer = null;
      schedule();
    };
    window.addEventListener("pointermove", onMove, { passive: true });
    window.addEventListener("pointerdown", onMove, { passive: true });
    window.addEventListener("pointerout", onOut, { passive: true });
    return () => {
      window.removeEventListener("pointermove", onMove);
      window.removeEventListener("pointerdown", onMove);
      window.removeEventListener("pointerout", onOut);
    };
  });

  onDestroy(() => {
    cancelRaf(pending);
    setSource(null);
  });
</script>

{#if arrows.length > 0}
  <svg class="targeting-arrows" aria-hidden="true">
    <defs>
      <marker
        id="{uid}-pick"
        viewBox="0 0 10 10"
        refX="8"
        refY="5"
        markerWidth="6"
        markerHeight="6"
        orient="auto-start-reverse"
      >
        <path d="M 0 0 L 10 5 L 0 10 z" fill={PICK_ARROW_COLOR} />
      </marker>
      <marker
        id="{uid}-follow"
        viewBox="0 0 10 10"
        refX="8"
        refY="5"
        markerWidth="6"
        markerHeight="6"
        orient="auto-start-reverse"
      >
        <path d="M 0 0 L 10 5 L 0 10 z" fill={FOLLOW_ARROW_COLOR} />
      </marker>
      <marker
        id="{uid}-snap"
        viewBox="0 0 10 10"
        refX="8"
        refY="5"
        markerWidth="6"
        markerHeight="6"
        orient="auto-start-reverse"
      >
        <path d="M 0 0 L 10 5 L 0 10 z" fill={FOLLOW_SNAP_COLOR} />
      </marker>
    </defs>
    {#each arrows as a (a.id)}
      <path
        class="arrow {a.kind}"
        class:snapped={a.snapped}
        d={curvePath(a)}
        stroke={a.color}
        marker-end={`url(#${uid}-${a.kind === "pick" ? "pick" : a.snapped ? "snap" : "follow"})`}
        data-targeting-arrow={a.kind}
        data-arrow-target={a.kind === "pick" ? a.id.slice(a.id.indexOf(":") + 1) : undefined}
      />
    {/each}
  </svg>
{/if}

<style>
  .targeting-arrows {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    z-index: 36;
    overflow: visible;
  }
  .arrow {
    fill: none;
    stroke-width: 3;
    stroke-linecap: round;
    /* Slight shadow so the arrow reads against busy battlefield art. */
    filter: drop-shadow(0 0 4px rgba(0, 0, 0, 0.6));
  }
  .arrow.follow {
    stroke-width: 2.5;
    stroke-dasharray: 7 5;
    opacity: 0.85;
  }
  .arrow.follow.snapped {
    stroke-dasharray: none;
    opacity: 1;
  }
  /* The source's glow (ADR 0119 §4): gold, so it never reads as the
     green "legal target" ring. */
  :global([data-targeting-source]) {
    filter: drop-shadow(0 0 5px color-mix(in srgb, var(--pick) 95%, transparent))
      drop-shadow(0 0 14px color-mix(in srgb, var(--pick) 60%, transparent));
    animation: -global-targeting-source-pulse 1.4s ease-in-out infinite alternate;
  }
  @keyframes -global-targeting-source-pulse {
    from {
      filter: drop-shadow(0 0 3px color-mix(in srgb, var(--pick) 80%, transparent))
        drop-shadow(0 0 8px color-mix(in srgb, var(--pick) 40%, transparent));
    }
    to {
      filter: drop-shadow(0 0 6px var(--pick))
        drop-shadow(0 0 16px color-mix(in srgb, var(--pick) 70%, transparent));
    }
  }
  @media (prefers-reduced-motion: reduce) {
    :global([data-targeting-source]) {
      animation: none;
    }
  }
  :global(:root[data-reduce-motion="1"] [data-targeting-source]) {
    animation: none;
  }
</style>

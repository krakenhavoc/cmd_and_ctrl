<script lang="ts">
  // CombatArrows is the full-board SVG overlay that draws a curved
  // line from each declared attacker to its target — opponent header
  // for an attack, attacker card for a block. Anchored to the .board
  // container with `position: absolute; pointer-events: none;` so it
  // never intercepts clicks.
  //
  // Source/target positions are read from the live DOM via
  // `getBoundingClientRect()` on `[data-instance-id="…"]` (cards) and
  // `[data-seat-id="…"]` (player headers), then translated into
  // board-relative coordinates. Re-measured on every snapshot tick
  // and on board resize, since opponent rotation, the hand fan, and
  // the auto-sized rows can all reflow card positions independently
  // of the combat state itself.
  //
  // The arrows themselves use a quadratic bezier curve with a control
  // point pulled toward the table centre, which reads as "an arc
  // across the table" rather than a straight gunfire line. Attacks
  // get a warm red; blocks get a cool blue. A small triangle marker
  // is placed at the target end so the direction is unambiguous.

  import type { CardView, GameView } from "../../protocol";
  import { gsap } from "gsap";

  interface Props {
    view: GameView;
    boardEl: HTMLElement | null;
  }

  const { view, boardEl }: Props = $props();

  type Pair =
    | { kind: "attack"; id: string; fromCardID: string; toSeatID: string }
    | { kind: "block"; id: string; fromCardID: string; toCardID: string };

  const pairs = $derived.by((): Pair[] => {
    const out: Pair[] = [];
    const byID = new Map<string, CardView>();
    for (const c of view.battlefield.cards) byID.set(c.instance_id, c);
    for (const c of view.battlefield.cards) {
      if (c.attacking_target) {
        out.push({
          kind: "attack",
          id: `atk-${c.instance_id}`,
          fromCardID: c.instance_id,
          toSeatID: c.attacking_target,
        });
      }
      if (c.blocking_target && byID.has(c.blocking_target)) {
        out.push({
          kind: "block",
          id: `blk-${c.instance_id}`,
          fromCardID: c.instance_id,
          toCardID: c.blocking_target,
        });
      }
    }
    return out;
  });

  interface ArrowGeo {
    id: string;
    kind: "attack" | "block";
    x1: number;
    y1: number;
    x2: number;
    y2: number;
    // cx/cy is the control point for the quadratic bezier — pulled
    // perpendicular to the chord by a fraction of the chord length
    // so the arc bows outward instead of sitting flat.
    cx: number;
    cy: number;
  }

  let arrows = $state<ArrowGeo[]>([]);

  function midpointOffset(
    x1: number,
    y1: number,
    x2: number,
    y2: number,
  ): { cx: number; cy: number } {
    // Curve bows perpendicular to the chord by ~20% of the chord
    // length, biased upward so attack arcs don't dive through the
    // hand fan when both endpoints are near the bottom of the board.
    const mx = (x1 + x2) / 2;
    const my = (y1 + y2) / 2;
    const dx = x2 - x1;
    const dy = y2 - y1;
    const len = Math.hypot(dx, dy) || 1;
    const nx = -dy / len;
    const ny = dx / len;
    const bow = len * 0.18;
    // Always bow upward (negative Y on screen). The sign of nx*ny
    // depends on chord direction; we just compare and pick the
    // candidate with the smaller Y.
    const candA = { cx: mx + nx * bow, cy: my + ny * bow };
    const candB = { cx: mx - nx * bow, cy: my - ny * bow };
    return candA.cy < candB.cy ? candA : candB;
  }

  function rectIn(boardRect: DOMRect, sel: string): { x: number; y: number } | null {
    if (!boardEl) return null;
    const el = boardEl.querySelector(sel) as HTMLElement | null;
    if (!el) return null;
    const r = el.getBoundingClientRect();
    return {
      x: r.left + r.width / 2 - boardRect.left,
      y: r.top + r.height / 2 - boardRect.top,
    };
  }

  function recomputeArrows(): void {
    if (!boardEl) {
      arrows = [];
      return;
    }
    const boardRect = boardEl.getBoundingClientRect();
    const next: ArrowGeo[] = [];
    for (const p of pairs) {
      const from = rectIn(boardRect, `[data-instance-id="${cssEscape(p.fromCardID)}"]`);
      if (!from) continue;
      let to: { x: number; y: number } | null = null;
      if (p.kind === "attack") {
        to = rectIn(boardRect, `[data-seat-id="${cssEscape(p.toSeatID)}"]`);
      } else {
        to = rectIn(boardRect, `[data-instance-id="${cssEscape(p.toCardID)}"]`);
      }
      if (!to) continue;
      const ctrl = midpointOffset(from.x, from.y, to.x, to.y);
      next.push({
        id: p.id,
        kind: p.kind,
        x1: from.x,
        y1: from.y,
        x2: to.x,
        y2: to.y,
        cx: ctrl.cx,
        cy: ctrl.cy,
      });
    }
    arrows = next;
  }

  // CSS.escape is the spec-correct selector escape but isn't typed on
  // every TS lib target; this thin wrapper keeps the call sites short
  // and falls back to identity on the (vanishingly unlikely) chance
  // it's missing.
  function cssEscape(s: string): string {
    return typeof CSS !== "undefined" && typeof CSS.escape === "function" ? CSS.escape(s) : s;
  }

  // Recompute on snapshot change (pairs change → DOM positions may
  // have shifted because the new card just mounted) AND on board
  // resize. requestAnimationFrame the first measurement so newly-
  // mounted DOM nodes have laid out before we read their rects.
  $effect(() => {
    void pairs;
    void view;
    if (!boardEl) return;
    let raf = requestAnimationFrame(recomputeArrows);
    const obs = new ResizeObserver(() => {
      cancelAnimationFrame(raf);
      raf = requestAnimationFrame(recomputeArrows);
    });
    obs.observe(boardEl);
    return () => {
      cancelAnimationFrame(raf);
      obs.disconnect();
    };
  });

  // Animate each new path in by tweening its stroke-dashoffset from
  // pathLength → 0, so arrows "draw on" rather than popping in. Keyed
  // by ARROW id so re-mounts (a card replays its attack declaration)
  // re-fire the draw.
  function drawOn(node: SVGPathElement): void {
    const len = node.getTotalLength();
    gsap.fromTo(
      node,
      { strokeDasharray: len, strokeDashoffset: len },
      {
        strokeDashoffset: 0,
        duration: 0.34,
        ease: "power2.out",
        overwrite: "auto",
      },
    );
  }
</script>

{#if arrows.length > 0}
  <svg class="arrows" aria-hidden="true">
    <defs>
      <marker
        id="arrowhead-attack"
        viewBox="0 0 10 10"
        refX="8"
        refY="5"
        markerWidth="6"
        markerHeight="6"
        orient="auto-start-reverse"
      >
        <path d="M 0 0 L 10 5 L 0 10 z" fill="#ff7a7a" />
      </marker>
      <marker
        id="arrowhead-block"
        viewBox="0 0 10 10"
        refX="8"
        refY="5"
        markerWidth="6"
        markerHeight="6"
        orient="auto-start-reverse"
      >
        <path d="M 0 0 L 10 5 L 0 10 z" fill="#9ec7ff" />
      </marker>
    </defs>
    {#each arrows as a (a.id)}
      <path
        d={`M ${a.x1} ${a.y1} Q ${a.cx} ${a.cy} ${a.x2} ${a.y2}`}
        class={a.kind}
        marker-end={a.kind === "attack" ? "url(#arrowhead-attack)" : "url(#arrowhead-block)"}
        use:drawOn
      />
    {/each}
  </svg>
{/if}

<style>
  .arrows {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    z-index: 35;
    overflow: visible;
  }
  path {
    fill: none;
    stroke-width: 3;
    stroke-linecap: round;
    /* Slight glow so the arrow reads against busy battlefield art. */
    filter: drop-shadow(0 0 4px rgba(0, 0, 0, 0.55));
  }
  path.attack {
    stroke: #ff7a7a;
  }
  path.block {
    stroke: #9ec7ff;
  }
</style>

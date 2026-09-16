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
  //
  // #187 / ADR 0053: combat damage beats. When a combat had a
  // first-strike step, the log tags each combat damage entry with the
  // step that dealt it, and this overlay cues the two steps one after
  // the other: a pulse on each live arrow the beat's damage travelled
  // along, a fading ghost arrow from cached geometry where an endpoint
  // has already left the battlefield, and a text cue ("First strike",
  // "Regular damage") that is real text in a live region. All the
  // rules — which entries are new, beat membership, the pause, live vs
  // ghost — live in lib/combatBeats.ts and are unit-tested there; this
  // file only measures, draws and renders. Cues are overlays: the board
  // already shows the frame's end state and input stays live under
  // them (pointer-events: none throughout).

  import { onDestroy, untrack } from "svelte";
  import { get } from "svelte/store";
  import type { CardView, GameView } from "../../protocol";
  import { gsap } from "gsap";
  import { settings } from "../../settings";
  import {
    BeatSequencer,
    arrowRender,
    beatMode,
    cueAnchor,
    cueDetail,
    emptyBeatTracker,
    ghostGeometry,
    keepArrowCache,
    midpointOffset,
    planFrame,
    type ArrowGeometry,
    type BeatTag,
    type CachedArrow,
    type Point,
    type ScheduledCue,
  } from "../../combatBeats";

  interface Props {
    view: GameView;
    boardEl: HTMLElement | null;
    // Changing this asks the beat tracker to prime again on the next
    // frame, as on a first frame: Game.svelte changes it across a
    // reconnect and a replay toggle, where the board stays mounted but
    // the frames in between were never watched live.
    beatsPrimeKey?: string;
  }

  const { view, boardEl, beatsPrimeKey = "" }: Props = $props();

  type Pair =
    | { kind: "attack"; id: string; fromCardID: string; toSeatID: string }
    | { kind: "block"; id: string; fromCardID: string; toCardID: string }
    | { kind: "stack-target-player"; id: string; fromStackID: string; toSeatID: string }
    | { kind: "stack-target-card"; id: string; fromStackID: string; toCardID: string };

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
    // S14: stack-item target arrows. Each StackItem carries its
    // announce-time targets[]; draw one arrow per player / card slot
    // so the board shows who each spell is aimed at. Self / none
    // slots don't produce arrows (no visual referent).
    for (const item of view.stack_items ?? []) {
      for (let i = 0; i < (item.targets?.length ?? 0); i++) {
        const t = item.targets![i];
        if (t.kind === "player" && t.id) {
          out.push({
            kind: "stack-target-player",
            id: `stk-${item.id}-p-${i}`,
            fromStackID: item.id,
            toSeatID: t.id,
          });
        } else if (t.kind === "card" && t.id) {
          out.push({
            kind: "stack-target-card",
            id: `stk-${item.id}-c-${i}`,
            fromStackID: item.id,
            toCardID: t.id,
          });
        }
      }
    }
    return out;
  });

  type ArrowKind = "attack" | "block" | "stack-target-player" | "stack-target-card";

  interface ArrowGeo {
    id: string;
    kind: ArrowKind;
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
      // Source: battlefield card (attack / block) or stack item
      // (stack-target-*). Route through the matching data attribute
      // in each case.
      let from: { x: number; y: number } | null = null;
      if (p.kind === "attack" || p.kind === "block") {
        from = rectIn(boardRect, `[data-instance-id="${cssEscape(p.fromCardID)}"]`);
      } else {
        from = rectIn(boardRect, `[data-stack-item-id="${cssEscape(p.fromStackID)}"]`);
      }
      if (!from) continue;
      let to: { x: number; y: number } | null = null;
      if (p.kind === "attack" || p.kind === "stack-target-player") {
        to = rectIn(boardRect, `[data-seat-id="${cssEscape(p.toSeatID)}"]`);
      } else {
        // Card targets can live in two places: the battlefield (via
        // data-instance-id on Card.svelte) or the stack (via
        // data-stack-item-id on StackOverlay's item div). For spell
        // stack items, id and instance_id are the same UUID (see
        // server/internal/game/stack.go StackItem.ID doc) so either
        // attribute resolves. Try the battlefield first; fall back
        // to the stack to cover Counterspell-style stack-on-stack
        // targeting.
        to =
          rectIn(boardRect, `[data-instance-id="${cssEscape(p.toCardID)}"]`) ??
          rectIn(boardRect, `[data-stack-item-id="${cssEscape(p.toCardID)}"]`);
      }
      if (!to) continue;
      const ctrl = midpointOffset(from.x, from.y, to.x, to.y);
      const geo: ArrowGeo = {
        id: p.id,
        kind: p.kind,
        x1: from.x,
        y1: from.y,
        x2: to.x,
        y2: to.y,
        cx: ctrl.cx,
        cy: ctrl.cy,
      };
      next.push(geo);
      // ADR 0053: keep the last measured geometry of every combat
      // arrow, so a beat can still draw it after an endpoint has left.
      if (p.kind === "attack" || p.kind === "block") {
        geoCache.set(p.id, {
          kind: p.kind,
          fromCardID: p.fromCardID,
          toSeatID: p.kind === "attack" ? p.toSeatID : undefined,
          toCardID: p.kind === "block" ? p.toCardID : undefined,
          geo: { x1: geo.x1, y1: geo.y1, x2: geo.x2, y2: geo.y2, cx: geo.cx, cy: geo.cy },
          board: { w: boardRect.width, h: boardRect.height },
        });
      }
    }
    arrows = next;
  }

  // ---- Combat damage beats (#187, ADR 0053) ----

  // geoCache is each attack / block arrow's last measured geometry,
  // keyed by arrow ID. Deliberately not reactive: nothing renders from
  // it directly. Kept through combat and until the last scheduled cue
  // has played (keepArrowCache), cleared on prime.
  const geoCache = new Map<string, CachedArrow>();

  // One pulse (on a live arrow) or ghost (from cached geometry). Each
  // removes itself when its tween completes.
  interface BeatEffect {
    key: string;
    style: "pulse" | "ghost";
    kind: "attack" | "block";
    geo: ArrowGeometry;
    durationMs: number;
  }
  let effects = $state<BeatEffect[]>([]);
  let effectCounter = 0;

  // One text cue per combat_damage step, positioned next to the
  // damage it describes. Lines join as beats are cued.
  interface CueBox {
    stepSeq: number;
    x: number;
    y: number;
    lines: { tag: BeatTag; label: string; detail: string }[];
  }
  let cueBoxes = $state<CueBox[]>([]);

  let tracker = emptyBeatTracker();
  let reprimeNext = false;

  const sequencer = new BeatSequencer({
    onCue: playCue,
    onHide: (stepSeq) => {
      cueBoxes = cueBoxes.filter((b) => b.stepSeq !== stepSeq);
    },
  });
  onDestroy(() => sequencer.dispose());

  function dropCacheIfDone(): void {
    if (!keepArrowCache(view.turn?.step, sequencer.pending)) geoCache.clear();
  }

  // measure is an endpoint's board-relative centre, or null when it is
  // not mounted. A card counts only while it is on the battlefield: a
  // dead creature's instance can still be drawn in a graveyard pile,
  // and a ghost must not point there.
  function measure(boardRect: DOMRect, cardID?: string, seatID?: string): Point | null {
    if (seatID) return rectIn(boardRect, `[data-seat-id="${cssEscape(seatID)}"]`);
    if (!cardID || !view.battlefield.cards.some((c) => c.instance_id === cardID)) return null;
    return rectIn(boardRect, `[data-instance-id="${cssEscape(cardID)}"]`);
  }

  function playCue(cue: ScheduledCue): void {
    if (!boardEl) return;
    // Measure now: the frame that carried the beat has rendered, so
    // the live set is the end state's, not the previous frame's.
    recomputeArrows();
    const boardRect = boardEl.getBoundingClientRect();
    const board = { w: boardRect.width, h: boardRect.height };
    const live = new Map(arrows.map((a) => [a.id, a]));

    const drawn: {
      id: string;
      style: "pulse" | "ghost";
      kind: "attack" | "block";
      geo: ArrowGeometry;
    }[] = [];
    for (const id of cue.arrowIDs) {
      const render = arrowRender(id, live, geoCache);
      if (render === "live") {
        const a = live.get(id)!;
        if (a.kind === "attack" || a.kind === "block")
          drawn.push({ id, style: "pulse", kind: a.kind, geo: a });
      } else if (render === "ghost") {
        const c = geoCache.get(id)!;
        const geo = ghostGeometry(
          c,
          board,
          measure(boardRect, c.fromCardID),
          measure(boardRect, c.toCardID, c.toSeatID),
        );
        // Null: the board changed size, so the cached point is a guess.
        // The text cue carries the beat.
        if (geo) drawn.push({ id, style: "ghost", kind: c.kind, geo });
      }
    }

    // Motion is decided again at cue time, so turning reduce-motion on
    // mid-sequence starts no further tween.
    const s = get(settings);
    const motion =
      cue.motion &&
      beatMode(s.animations.enabled, s.animations.damagePopups, s.accessibility.reduceMotion) ===
        "full";
    if (motion && drawn.length > 0) {
      effects = [
        ...effects,
        ...drawn.map((d) => ({
          key: `${cue.stepSeq}-${cue.tag}-${d.id}-${effectCounter++}`,
          style: d.style,
          kind: d.kind,
          geo: d.geo,
          durationMs: cue.effectMs,
        })),
      ];
    }

    if (cue.label !== null) {
      const line = { tag: cue.tag, label: cue.label, detail: cueDetail(cue) };
      const existing = cueBoxes.find((b) => b.stepSeq === cue.stepSeq);
      if (existing) {
        cueBoxes = cueBoxes.map((b) =>
          b.stepSeq === cue.stepSeq
            ? { ...b, lines: [...b.lines.filter((l) => l.tag !== cue.tag), line] }
            : b,
        );
      } else {
        const at = cueAnchor(
          drawn.map((d) => d.geo),
          board,
        );
        cueBoxes = [...cueBoxes, { stepSeq: cue.stepSeq, x: at.x, y: at.y, lines: [line] }];
      }
    }
    dropCacheIfDone();
  }

  // A prime-key change (reconnect, replay toggle) primes the next
  // frame. Declared before the frame effect so a key and a view that
  // change together prime on that same frame.
  $effect(() => {
    void beatsPrimeKey;
    untrack(() => {
      reprimeNext = true;
    });
  });

  // Plan each frame's beats. Only `view` is tracked: settings are read
  // at plan time, and nothing a later frame does cancels a cue.
  $effect(() => {
    const log = view.log;
    untrack(() => {
      const s = get(settings);
      const plan = planFrame(tracker, log, {
        mode: beatMode(
          s.animations.enabled,
          s.animations.damagePopups,
          s.accessibility.reduceMotion,
        ),
        speed: s.animations.speed,
        reprime: reprimeNext,
      });
      reprimeNext = false;
      tracker = plan.tracker;
      if (plan.primed) geoCache.clear();
      sequencer.play(plan.cues);
      dropCacheIfDone();
    });
  });

  // beatEffect runs one pulse or ghost: in and out within its beat,
  // then removes itself. Only ever mounted in "full" mode.
  function beatEffect(node: SVGPathElement, effect: BeatEffect) {
    const peak = effect.style === "pulse" ? 0.85 : 0.5;
    const tween = gsap.fromTo(
      node,
      { opacity: 0 },
      {
        opacity: peak,
        duration: effect.durationMs / 2000,
        ease: "sine.inOut",
        yoyo: true,
        repeat: 1,
        onComplete: () => {
          effects = effects.filter((e) => e.key !== effect.key);
        },
      },
    );
    return {
      destroy() {
        tween.kill();
      },
    };
  }

  function arcPath(g: ArrowGeometry): string {
    return `M ${g.x1} ${g.y1} Q ${g.cx} ${g.cy} ${g.x2} ${g.y2}`;
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
  function markerFor(kind: ArrowKind): string {
    switch (kind) {
      case "attack":
        return "url(#arrowhead-attack)";
      case "block":
        return "url(#arrowhead-block)";
      case "stack-target-player":
      case "stack-target-card":
        return "url(#arrowhead-target)";
    }
  }

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

{#if arrows.length > 0 || effects.length > 0}
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
      <marker
        id="arrowhead-target"
        viewBox="0 0 10 10"
        refX="8"
        refY="5"
        markerWidth="6"
        markerHeight="6"
        orient="auto-start-reverse"
      >
        <path d="M 0 0 L 10 5 L 0 10 z" fill="#ffd07a" />
      </marker>
    </defs>
    {#each arrows as a (a.id)}
      <path
        d={`M ${a.x1} ${a.y1} Q ${a.cx} ${a.cy} ${a.x2} ${a.y2}`}
        class={a.kind}
        marker-end={markerFor(a.kind)}
        use:drawOn
      />
    {/each}
    <!-- ADR 0053 beat effects. A pulse is a wide soft stroke over a
         live arrow. A ghost is its own style — thin, dashed, no glow,
         no arrowhead — so it never reads as a live arrow, and it fades
         out within its beat. -->
    {#each effects as e (e.key)}
      <path
        d={arcPath(e.geo)}
        class="beat-{e.style} {e.kind}"
        style="opacity: 0"
        use:beatEffect={e}
      />
    {/each}
  </svg>
{/if}

<!-- The combat damage beat cue: real text, so it is in the
     accessibility tree (the SVG above is aria-hidden). The live region
     is always mounted, empty between beats, so a screen reader
     announces each cue as it joins. Shown in both modes; only the
     arrow effects are motion. -->
<div class="beat-cues" role="status" aria-live="polite">
  {#each cueBoxes as box (box.stepSeq)}
    <div class="beat-cue" style:left="{box.x}px" style:top="{box.y}px">
      {#each box.lines as line (line.tag)}
        <span class="beat-line {line.tag}"
          >{line.label}<span class="sr-only">: {line.detail}</span></span
        >
      {/each}
    </div>
  {/each}
</div>

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
  /* S14: stack-target arrows use the same gold palette as the
     AUTO badge + targeting banner — reads as "a catalog spell is
     pointing at this thing." Dashed to distinguish from combat's
     solid red/blue when both are drawn simultaneously. */
  path.stack-target-player,
  path.stack-target-card {
    stroke: #ffd07a;
    stroke-dasharray: 6 4;
  }
  /* ADR 0053 beat effects. Opacity is driven by the tween. */
  path.beat-pulse {
    stroke-width: 9;
    filter: blur(2px);
  }
  path.beat-ghost {
    stroke-width: 2;
    stroke-dasharray: 3 7;
    filter: none;
  }
  .beat-cues {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 36;
    overflow: hidden;
  }
  .beat-cue {
    position: absolute;
    transform: translate(-50%, -50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }
  .beat-line {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    font-weight: 700;
    white-space: nowrap;
    padding: 4px 9px;
    border-radius: 999px;
    background: color-mix(in srgb, var(--surface) 92%, transparent);
    box-shadow: var(--shadow-lg);
    color: var(--fg);
    border: 1px solid rgba(255, 255, 255, 0.18);
  }
  .beat-line.first_strike {
    border-color: rgba(217, 180, 92, 0.7);
    color: var(--gold-strong);
  }
  .beat-line.regular {
    border-color: rgba(255, 122, 122, 0.6);
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
</style>

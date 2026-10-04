<script lang="ts">
  // DiceLayer draws every die a card rolls and every coin it flips, and
  // every die or coin a player rolls at the table (ADR 0121 §5), at the
  // roller's seat, for everyone at the table (ADR 0121 §7).
  //
  // The result is already drawn when the frame arrives (ADR 0054): this
  // is presentation. The schedule (what plays when, the six-die cap, the
  // burst rule, the deterministic tumble) is lib/dice.ts, pure and
  // unit-tested; the shared queue is lib/diceQueue.svelte.ts, so the
  // attention strip can hold a roll's text until its die settles. This
  // file measures, draws and announces.
  //
  // Display only. Nothing waits for a die: the layer is aria-hidden,
  // takes no pointer events, and sits at z 41, above the strip (40), the
  // pile (38) and the expanded board (32), below the dock (55), menus
  // and modals. Each die settles beside its roller's avatar, found
  // through boardAnchor (so an expanded board's copy wins), on the side
  // toward the table's centre. A seat whose avatar is not on screen
  // gets its die under the attention strip instead.

  import { onDestroy, untrack } from "svelte";
  import { get } from "svelte/store";
  import type { GameView } from "../../protocol";
  import { settings } from "../../settings";
  import { seatColor } from "../../colors";
  import { findSeatAnchor, hasSize } from "../../boardAnchor";
  import { attentionStrip, stripContentBottom } from "../../stackPile";
  import {
    COIN_PX,
    DIE_PX,
    coinTurns,
    coinWon,
    diceGroupSize,
    diceLabel,
    diceMotion,
    dicePhase,
    diceShown,
    diceThrow,
    faceAt,
    placeDice,
    rollFromLog,
    rollsFromLogs,
    visiblePlays,
    type DicePlacement,
    type DicePlay,
    type Rect,
  } from "../../dice";
  import { DiceQueue, useDiceQueue } from "../../diceQueue.svelte";
  import { latestOpeningDice } from "../../openingRoll";

  interface Props {
    view: GameView;
    boardEl: HTMLElement | null;
    // Changing this primes the next frame, as on a first frame: a
    // reconnect or a replay toggle (the CombatArrows contract).
    beatsPrimeKey?: string;
    // The game screen's queue. Defaults to the one Game.svelte provides;
    // a layer mounted alone makes its own.
    queue?: DiceQueue;
  }

  const { view, boardEl, beatsPrimeKey = "", queue: given }: Props = $props();

  // Fixed for the life of the layer: the queue is never swapped.
  const provided = useDiceQueue();
  const initial = untrack(() => given);
  const own = provided === null && initial === undefined ? new DiceQueue() : null;
  const queue: DiceQueue = initial ?? provided ?? (own as DiceQueue);
  onDestroy(() => own?.dispose());

  // The wall clock the faces are read at. Moved by the queue's own
  // wake-ups (a die starts, settles, fades) and, only while something
  // tumbles, by animation frames.
  let now = $state(Date.now());
  $effect(() => {
    void queue.tick;
    untrack(() => (now = Date.now()));
  });

  // Roll keys already read out (or primed): each roll is announced once.
  const announced = new Set<string>();

  let reprimeNext = false;
  $effect(() => {
    void beatsPrimeKey;
    untrack(() => {
      reprimeNext = true;
    });
  });

  // Fold each frame into the schedule. Only `view.log` is tracked;
  // settings are read at plan time, so turning motion off mid-roll
  // changes the next roll, not the one in the air.
  let firstFrame = true;
  $effect(() => {
    const log = view.log;
    untrack(() => {
      const rolls = rollsFromLogs(log);
      if (firstFrame || reprimeNext) {
        firstFrame = false;
        reprimeNext = false;
        queue.prime(rolls);
        for (const r of rolls) announced.add(r.key);
        return;
      }
      const s = get(settings);
      queue.ingest(rolls, Date.now(), {
        motion: diceMotion({
          enabled: s.animations.enabled,
          dice: s.animations.dice,
          reduceMotion: s.accessibility.reduceMotion,
        }),
        speed: s.animations.speed,
      });
    });
  });

  const plays = $derived(visiblePlays({ plays: queue.plays }, now));

  // ADR 0121 §7: the opening roll's dice do not fade. While the roll is
  // open, each seat's latest opening die stays settled beside its seat,
  // as the standings, until the winner chooses. A die still tumbling or
  // holding is drawn by its play (and its fade is held off while the
  // roll is open); once that play is over, the same die is drawn here,
  // under the same key, so it is the same element and nothing flickers.
  // A die the queue has not settled yet (or not seen yet) is not drawn
  // here, so a standing never gives a result away before its tumble. A
  // reconnect primes the window, and primed dice are settled at once.
  const openingOpen = $derived(!!view.opening_roll);
  const standings = $derived.by((): DicePlay[] => {
    if (!view.opening_roll) return [];
    void queue.tick;
    const busy = new Set(plays.map((p) => p.roll.seat));
    const out: DicePlay[] = [];
    for (const log of latestOpeningDice(view.log)) {
      if (busy.has(log.seat) || !queue.released(log.seq, now)) continue;
      const roll = rollFromLog(log);
      if (!roll) continue;
      out.push({
        roll,
        startAt: 0,
        settleAt: 0,
        fadeAt: Number.POSITIVE_INFINITY,
        endAt: Number.POSITIVE_INFINITY,
        motion: false,
        slotMs: 0,
        releaseAt: 0,
      });
    }
    return out;
  });
  const drawn = $derived([...plays, ...standings]);

  function phaseOf(p: DicePlay, at: number): ReturnType<typeof dicePhase> {
    const phase = dicePhase(p, at);
    return phase === "fade" && openingOpen && p.roll.source === "opening" ? "hold" : phase;
  }

  // Animation frames, only while a die is in the air: the faces change
  // on a seeded schedule (lib/dice.ts tumbleIndex).
  $effect(() => {
    void queue.tick;
    const tumbling = queue.plays.some(
      (p) => p.motion && p.startAt <= Date.now() && Date.now() < p.settleAt,
    );
    if (!tumbling || typeof requestAnimationFrame !== "function") return;
    let raf = 0;
    const loop = () => {
      now = Date.now();
      if (queue.plays.some((p) => p.motion && p.startAt <= now && now < p.settleAt)) {
        raf = requestAnimationFrame(loop);
      }
    };
    raf = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(raf);
  });

  // ---------------------------------------------------------------- //
  // Where each group goes. Measured when the set on screen changes, on
  // every frame, and when the board resizes, never per animation frame.

  let placements = $state.raw<Map<string, DicePlacement>>(new Map());
  let resized = $state(0);
  const visibleKeys = $derived(drawn.map((p) => p.roll.key).join("|"));

  $effect(() => {
    const el = boardEl;
    if (!el || typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(() => resized++);
    ro.observe(el);
    return () => ro.disconnect();
  });

  function relRect(board: DOMRect, el: Element): Rect {
    const r = el.getBoundingClientRect();
    return { left: r.left - board.left, top: r.top - board.top, width: r.width, height: r.height };
  }

  function onBoard(board: DOMRect, r: Rect): boolean {
    return (
      r.left + r.width > 0 && r.top + r.height > 0 && r.left < board.width && r.top < board.height
    );
  }

  function stripRect(board: HTMLElement, boardRect: DOMRect): Rect | null {
    const strip = attentionStrip(board);
    if (!strip) return null;
    const r = relRect(boardRect, strip);
    const bottom = stripContentBottom(board);
    return { left: r.left, top: bottom ?? r.top, width: r.width, height: 0 };
  }

  $effect(() => {
    void visibleKeys;
    void view;
    void resized;
    const board = boardEl;
    untrack(() => {
      if (!board || drawn.length === 0) {
        if (placements.size > 0) placements = new Map();
        return;
      }
      const boardRect = board.getBoundingClientRect();
      const size = { width: boardRect.width, height: boardRect.height };
      const next = new Map<string, DicePlacement>();
      for (const p of drawn) {
        const seatID = view.seats[p.roll.seat]?.id;
        const anchor = seatID ? findSeatAnchor(board, seatID, { accept: hasSize }) : null;
        let avatar: Rect | null = anchor ? relRect(boardRect, anchor) : null;
        if (avatar && !onBoard(boardRect, avatar)) avatar = null;
        next.set(
          p.roll.key,
          placeDice(size, avatar, diceGroupSize(p.roll), stripRect(board, boardRect)),
        );
      }
      placements = next;
    });
  });

  // ---------------------------------------------------------------- //
  // The announcer: each roll read once, when its die settles (at once
  // with motion off). Always mounted, so additions are read reliably.

  let lines = $state.raw<{ seq: number; text: string }[]>([]);
  $effect(() => {
    void queue.tick;
    const log = view.log;
    untrack(() => {
      const t = Date.now();
      const rolls = rollsFromLogs(log);
      const present = new Set(rolls.map((r) => r.key));
      // A rewound roll that is rolled again is read again. A table roll
      // whose line an undo wrote again keeps its key, so it is not.
      for (const key of [...announced]) if (!present.has(key)) announced.delete(key);
      const fresh: { seq: number; text: string }[] = [];
      for (const r of rolls) {
        if (announced.has(r.key) || !queue.released(r.seq, t)) continue;
        announced.add(r.key);
        if (queue.wasPrimed(r.seq)) continue;
        fresh.push({ seq: r.seq, text: r.text });
      }
      if (fresh.length > 0) lines = [...lines, ...fresh].slice(-3);
    });
  });

  function pips(n: number): [number, number][] {
    const L = 12.5;
    const C = 21;
    const R = 29.5;
    const T = 18.5;
    const M = 27;
    const B = 35.5;
    switch (n) {
      case 1:
        return [[C, M]];
      case 2:
        return [
          [L, T],
          [R, B],
        ];
      case 3:
        return [
          [L, T],
          [C, M],
          [R, B],
        ];
      case 4:
        return [
          [L, T],
          [R, T],
          [L, B],
          [R, B],
        ];
      case 5:
        return [
          [L, T],
          [R, T],
          [C, M],
          [L, B],
          [R, B],
        ];
      case 6:
        return [
          [L, T],
          [L, M],
          [L, B],
          [R, T],
          [R, M],
          [R, B],
        ];
      default:
        return [];
    }
  }

  function numberSize(n: number): number {
    const digits = String(n).length;
    return digits >= 3 ? 9.5 : digits === 2 ? 12.5 : 14;
  }

  function throwStyle(p: DicePlay, i: number): string {
    const t = diceThrow(p.roll.key, i);
    return `--spin:${t.spin}deg;--from-x:${t.fromX}px;--from-y:${t.fromY}px;--tumble:${p.settleAt - p.startAt}ms;--delay:${i * 40}ms`;
  }

  function coinStyle(p: DicePlay, i: number): string {
    const face = p.roll.faces[i];
    const turn = p.motion ? coinTurns(p.roll.key, i, face) * 180 : face === "tails" ? 180 : 0;
    return `--turn:${turn}deg;--tumble:${p.settleAt - p.startAt}ms;--delay:${i * 40}ms`;
  }
</script>

<div class="dice-layer" aria-hidden="true" style="--die-px:{DIE_PX}px;--coin-px:{COIN_PX}px">
  {#each drawn as p (p.roll.key)}
    {@const at = placements.get(p.roll.key)}
    {@const phase = phaseOf(p, now)}
    {@const shown = diceShown(p.roll)}
    {@const box = diceGroupSize(p.roll)}
    {#if at}
      <div
        class="group {at.side}"
        class:moving={p.motion}
        class:settled={phase !== "tumble"}
        class:fading={phase === "fade"}
        data-dice-key={p.roll.key}
        data-dice-phase={phase}
        data-dice-seat={p.roll.seat}
        style:left="{at.left}px"
        style:top="{at.top}px"
        style:width="{box.width}px"
        style:height="{box.height}px"
        style:--seat={seatColor(p.roll.seat)}
      >
        {#if p.roll.call}
          <span class="call">called {p.roll.call}</span>
        {/if}
        <div class="row">
          {#each Array.from({ length: shown.count }, (_, i) => i) as i (i)}
            {#if p.roll.kind === "die"}
              {@const face = faceAt(p, i, now)}
              <span class="die" style={throwStyle(p, i)} data-face={face}>
                {#if p.roll.sides === 6}
                  <svg viewBox="0 0 48 48" width={DIE_PX} height={DIE_PX}>
                    <polygon class="d6-top" points="4,10 10,4 44,4 38,10" />
                    <polygon class="d6-side" points="38,10 44,4 44,38 38,44" />
                    <rect class="d6-front" x="4" y="10" width="34" height="34" rx="5" />
                    {#each pips(face) as [x, y], k (k)}
                      <circle class="pip" cx={x} cy={y} r="3.3" />
                    {/each}
                  </svg>
                {:else}
                  <svg viewBox="0 0 48 48" width={DIE_PX} height={DIE_PX}>
                    <polygon class="poly-body" points="24,2 43,13 43,35 24,46 5,35 5,13" />
                    <polygon class="poly-facet dark" points="24,2 5,13 24,10" />
                    <polygon class="poly-facet dark" points="24,2 43,13 24,10" />
                    <polygon class="poly-facet" points="5,13 11,33 24,10" />
                    <polygon class="poly-facet" points="43,13 37,33 24,10" />
                    <polygon class="poly-facet dark" points="5,13 5,35 11,33" />
                    <polygon class="poly-facet dark" points="43,13 43,35 37,33" />
                    <polygon class="poly-facet" points="5,35 24,46 11,33" />
                    <polygon class="poly-facet" points="43,35 24,46 37,33" />
                    <polygon class="poly-facet dark" points="11,33 37,33 24,46" />
                    <polygon class="poly-face" points="24,10 37,33 11,33" />
                    <text class="num" x="24" y="28.5" font-size={numberSize(face)}>{face}</text>
                  </svg>
                {/if}
              </span>
            {:else}
              {@const won = coinWon(p.roll, i)}
              <span class="coin-slot">
                <span class="coin" style={coinStyle(p, i)}>
                  <span class="coin-face heads">H</span>
                  <span class="coin-face tails">T</span>
                </span>
                {#if phase !== "tumble" && won !== undefined}
                  <span class="verdict" class:won class:lost={!won}>{won ? "won" : "lost"}</span>
                {/if}
              </span>
            {/if}
          {/each}
          {#if shown.more > 0}
            <span class="more">+{shown.more}</span>
          {/if}
        </div>
        <span class="caption">
          {#if phase !== "tumble" && p.roll.kind === "coin" && p.roll.faces.length === 1}
            {p.roll.faces[0]}
          {:else}
            {diceLabel(p.roll)}
          {/if}
        </span>
      </div>
    {/if}
  {/each}
</div>

<!-- Real text for a screen reader: the layer above is aria-hidden. -->
<div class="dice-announcer" role="status" aria-live="polite">
  {#each lines as line (line.seq)}
    <p>{line.text}</p>
  {/each}
</div>

<style>
  .dice-layer {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 41;
    overflow: visible;
  }
  .group {
    position: absolute;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    padding: 6px;
    box-sizing: border-box;
    opacity: 1;
    transition: opacity 200ms linear;
    filter: drop-shadow(0 6px 10px rgba(0, 0, 0, 0.55));
  }
  .group.fading {
    opacity: 0;
  }
  .row {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  /* ---- dice ---- */
  .die {
    display: block;
    width: var(--die-px);
    height: var(--die-px);
    transform-origin: 50% 55%;
  }
  .die svg {
    display: block;
    overflow: visible;
  }
  .group.moving:not(.settled) .die {
    animation: dice-tumble var(--tumble) cubic-bezier(0.18, 0.7, 0.3, 1) var(--delay) both;
  }
  .group.moving.settled .die {
    animation: dice-land 380ms cubic-bezier(0.3, 1.6, 0.5, 1) both;
  }
  .group.settled .die svg {
    filter: drop-shadow(0 0 7px color-mix(in srgb, var(--seat) 75%, transparent));
  }
  .d6-front {
    fill: var(--seat);
    stroke: color-mix(in srgb, var(--seat) 55%, black);
    stroke-width: 1;
  }
  .d6-top {
    fill: color-mix(in srgb, var(--seat) 55%, white);
    stroke: color-mix(in srgb, var(--seat) 55%, black);
    stroke-width: 1;
    stroke-linejoin: round;
  }
  .d6-side {
    fill: color-mix(in srgb, var(--seat) 60%, black);
    stroke: color-mix(in srgb, var(--seat) 45%, black);
    stroke-width: 1;
    stroke-linejoin: round;
  }
  .pip {
    fill: #0e1116;
  }
  .poly-body {
    fill: color-mix(in srgb, var(--seat) 45%, black);
    stroke: color-mix(in srgb, var(--seat) 70%, white);
    stroke-width: 1.4;
    stroke-linejoin: round;
  }
  .poly-facet {
    fill: color-mix(in srgb, var(--seat) 70%, black);
    stroke: color-mix(in srgb, var(--seat) 60%, white);
    stroke-width: 0.8;
    stroke-linejoin: round;
  }
  .poly-facet.dark {
    fill: color-mix(in srgb, var(--seat) 48%, black);
  }
  .poly-face {
    fill: var(--seat);
    stroke: color-mix(in srgb, var(--seat) 55%, white);
    stroke-width: 1;
    stroke-linejoin: round;
  }
  .num {
    fill: #0e1116;
    font-family: var(--font-mono);
    font-weight: 800;
    text-anchor: middle;
    dominant-baseline: middle;
    font-variant-numeric: tabular-nums;
  }

  /* ---- coins ---- */
  .coin-slot {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    perspective: 400px;
  }
  .coin {
    position: relative;
    display: block;
    width: var(--coin-px);
    height: var(--coin-px);
    transform-style: preserve-3d;
    transform: rotateY(var(--turn));
  }
  .group.moving:not(.settled) .coin {
    animation: coin-flip var(--tumble) cubic-bezier(0.2, 0.6, 0.35, 1) var(--delay) both;
  }
  .coin-face {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    backface-visibility: hidden;
    -webkit-backface-visibility: hidden;
    font-family: var(--font-mono);
    font-weight: 800;
    font-size: 21px;
    color: #0e1116;
    border: 3px solid color-mix(in srgb, var(--seat) 70%, black);
    box-sizing: border-box;
  }
  .coin-face.heads {
    background: radial-gradient(
      circle at 35% 30%,
      color-mix(in srgb, var(--gold-strong) 55%, white),
      var(--gold-strong) 60%,
      color-mix(in srgb, var(--gold-strong) 60%, black)
    );
  }
  .coin-face.tails {
    transform: rotateY(180deg);
    background: radial-gradient(
      circle at 35% 30%,
      color-mix(in srgb, var(--seat) 45%, white),
      var(--seat) 60%,
      color-mix(in srgb, var(--seat) 60%, black)
    );
  }
  /* On the faces, not the coin: a filter on the 3D container would
     flatten it and show the wrong side. */
  .group.settled .coin-face {
    box-shadow: 0 0 10px color-mix(in srgb, var(--seat) 70%, transparent);
  }
  .verdict {
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    padding: 1px 6px;
    border-radius: 999px;
    background: color-mix(in srgb, var(--surface) 88%, transparent);
  }
  .verdict.won {
    color: var(--mint);
  }
  .verdict.lost {
    color: var(--danger);
  }

  /* ---- labels ---- */
  .call,
  .caption,
  .more {
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    white-space: nowrap;
    padding: 2px 7px;
    border-radius: 999px;
    color: var(--fg);
    background: color-mix(in srgb, var(--surface) 88%, transparent);
    border: 1px solid color-mix(in srgb, var(--seat) 55%, transparent);
  }
  .more {
    font-size: 12px;
    letter-spacing: 0.02em;
    font-variant-numeric: tabular-nums;
  }

  .dice-announcer {
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

  @keyframes dice-tumble {
    0% {
      opacity: 0;
      transform: translate(var(--from-x), var(--from-y)) rotate(0deg) scale(0.55);
    }
    14% {
      opacity: 1;
    }
    48% {
      transform: translate(calc(var(--from-x) * 0.3), 2px) rotate(calc(var(--spin) * 0.72))
        scale(1.06);
    }
    66% {
      transform: translate(calc(var(--from-x) * 0.12), -7px) rotate(calc(var(--spin) * 0.9))
        scale(1);
    }
    84% {
      transform: translate(0, 1px) rotate(calc(var(--spin) * 0.985)) scale(1.02);
    }
    100% {
      opacity: 1;
      transform: translate(0, 0) rotate(var(--spin)) scale(1);
    }
  }
  @keyframes dice-land {
    0% {
      transform: scale(1);
    }
    40% {
      transform: scale(1.16);
    }
    100% {
      transform: scale(1);
    }
  }
  @keyframes coin-flip {
    0% {
      opacity: 0;
      transform: translateY(6px) rotateY(0deg) scale(0.7);
    }
    10% {
      opacity: 1;
    }
    45% {
      transform: translateY(-34px) rotateY(calc(var(--turn) * 0.6)) scale(1.12);
    }
    100% {
      opacity: 1;
      transform: translateY(0) rotateY(var(--turn)) scale(1);
    }
  }

  /* The settings already gate motion (master switch, the dice toggle,
     reduced motion); this is the same belt the other overlays wear. */
  @media (prefers-reduced-motion: reduce) {
    .group .die,
    .group .coin {
      animation: none !important;
    }
  }
  :global(:root[data-reduce-motion="1"]) .group .die,
  :global(:root[data-reduce-motion="1"]) .group .coin {
    animation: none !important;
  }
</style>

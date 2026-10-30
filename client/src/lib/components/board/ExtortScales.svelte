<script lang="ts">
  // ExtortScales — the Scales of Souls (#4544, from Ian's mock).
  //
  // When an extort trigger resolves and drains, a blindfolded judge holding
  // golden scales appears at the middle of the table, small. A soul leaves
  // each opponent that lost life and flies into the dark pan, which is
  // weighing the scale down; the scale levels (the coin hits, at
  // EXTORT_HIT_MS, where the `extort` sound's coin is cut to land); the
  // light on the other pan bursts and a spark flies to the extort player.
  // About 2.6 s at animations.speed 1. The life itself is already on the
  // avatars (their popups); this is presentation over it, and nothing waits
  // for it (ADR 0209 §1: no input, priority pass or bot seat ever does).
  //
  // Every rule is lib/extort.ts, pure and unit-tested: what an extort drain
  // is (a `resolve` entry labelled " — extort" and the `life` entries of
  // its batch), how a run of them queues (one scale at a time; the drains
  // that wait merge into one more scale), the clock and the geometry. This
  // file measures, draws and plays. The avatars' own damage and heal sounds
  // stand down for a drain it covers (PlayerIdentity.svelte), so the drain
  // is sounded once.
  //
  // Reduced motion, all three ways (DESIGN.md §4): accessibility.reduceMotion,
  // animations.enabled / animations.effects Off, and the OS signal, here and
  // in the @media block below. So does the Short setting. They show a static,
  // levelled scale and a fade, with no flight; the sound plays.
  //
  // Skip: a click anywhere or Escape clears the scene and the one waiting.
  // It is this player's alone (nothing is sent), and the layer takes no
  // pointer events, so the click also reaches the board under it.
  //
  // z 57 on DESIGN.md §4's ladder, the table-effects rung: after
  // TableEffects in the DOM, above the action dock (55), below the hints
  // (58). aria-hidden: the log and the life popups carry the news.

  import { onDestroy, untrack } from "svelte";
  import { get } from "svelte/store";
  import type { GameView } from "../../protocol";
  import { settings } from "../../settings";
  import { findSeatAnchor, hasSize } from "../../boardAnchor";
  import { emptyBeatTracker, track } from "../../combatBeats";
  import { effectMotion } from "../../tableEffects";
  import { play } from "../../sounds";
  import { L } from "../../labels";
  import {
    EXTORT_BEATS,
    EXTORT_MS,
    arrive,
    darkPanAt,
    emptyExtortQueue,
    emptyExtortTracker,
    extortK,
    extortMotion,
    extortMs,
    finish,
    flight,
    foldExtort,
    lightPanAt,
    mergeDrains,
    scaleBox,
    type Box,
    type ExtortDrain,
    type ExtortMotion,
    type ExtortQueue,
  } from "../../extort";

  interface Props {
    view: GameView;
    boardEl: HTMLElement | null;
    // Changing this primes the next frame, as on a first frame: nothing
    // already on the wire is cued (Game.svelte's beatsPrimeKey: a
    // reconnect, a replay toggle).
    beatsPrimeKey?: string;
    // The viewer's seat ID: a seat whose avatar is off screen is drawn at the
    // board's near edge for the viewer and the far edge for everyone else.
    viewerID?: string | null;
  }

  const { view, boardEl, beatsPrimeKey = "", viewerID = null }: Props = $props();

  // A point in board px: a soul's start, a spark's end, a ring.
  interface Spot {
    x: number;
    y: number;
  }
  interface Scene {
    key: number;
    motion: ExtortMotion;
    // Real ms per scene ms, and the scene's length in real ms.
    k: number;
    ms: number;
    box: Box;
    dark: Spot;
    light: Spot;
    // Each soul: where it starts relative to the dark pan, and its stagger.
    souls: { dx: number; dy: number; delay: number; at: Spot }[];
    // Each spark: where it ends relative to the light pan.
    sparks: { dx: number; dy: number; at: Spot }[];
    count: number;
  }

  let scene = $state<Scene | null>(null);
  let serial = 0;
  let destroyed = false;
  let queue: ExtortQueue = emptyExtortQueue;
  let timer: ReturnType<typeof setTimeout> | null = null;

  let tracker = emptyBeatTracker();
  let extort = emptyExtortTracker();
  let reprimeNext = false;

  function osReducesMotion(): boolean {
    return (
      typeof window !== "undefined" &&
      typeof window.matchMedia === "function" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches
    );
  }

  function clearAll(): void {
    if (timer) clearTimeout(timer);
    timer = null;
    queue = emptyExtortQueue;
    scene = null;
  }

  // Where a seat's avatar is, in board px: the centre of its anchor, else
  // the board's near or far edge (a seat whose avatar is off screen).
  function seatSpot(index: number, rect: DOMRect, near: boolean): Spot {
    const id = view.seats?.[index]?.id;
    const el = id && boardEl ? findSeatAnchor(boardEl, id, { accept: hasSize }) : null;
    if (el) {
      const r = el.getBoundingClientRect();
      return { x: r.left + r.width / 2 - rect.left, y: r.top + r.height / 2 - rect.top };
    }
    return { x: rect.width / 2, y: near ? rect.height : 0 };
  }

  function start(drain: ExtortDrain): void {
    const s = get(settings);
    const em = effectMotion({
      enabled: s.animations.enabled,
      effects: s.animations.effects,
      reduceMotion: s.accessibility.reduceMotion,
      osReduceMotion: osReducesMotion(),
    });
    const motion = extortMotion(em);
    const ms = extortMs(em, s.animations.speed);
    // The coin lands about 1.25 s in (EXTORT_HIT_MS): the sound is cut to it.
    play("extort");

    const rect = boardEl?.getBoundingClientRect();
    if (rect && rect.width > 0 && rect.height > 0) {
      const box = scaleBox({ w: rect.width, h: rect.height });
      const dark = darkPanAt(box);
      const light = lightPanAt(box);
      const viewer = viewerID ? (view.seats?.findIndex((st) => st.id === viewerID) ?? -1) : -1;
      scene = {
        key: ++serial,
        motion,
        k: extortK(em, s.animations.speed),
        ms,
        box,
        dark,
        light,
        souls: drain.victims.map((v, i) => {
          const at = seatSpot(v.seat, rect, v.seat === viewer);
          return { ...flight(at, dark, i), at };
        }),
        sparks: drain.gainers.map((g) => {
          const at = seatSpot(g.seat, rect, g.seat === viewer);
          return { dx: Math.round(at.x - light.x), dy: Math.round(at.y - light.y), at };
        }),
        count: drain.count,
      };
    } else {
      // No board to measure: the sound has played; draw nothing.
      scene = null;
    }
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => {
      timer = null;
      if (destroyed) return;
      const next = finish(queue);
      queue = next.queue;
      scene = null;
      if (next.start) start(next.start);
    }, ms);
  }

  function frame(log: GameView["log"]): void {
    const prime = reprimeNext;
    reprimeNext = false;
    const r = track(tracker, log, prime);
    tracker = r.tracker;
    // A reconnect or a replay move: forget what was playing and what was seen.
    if (r.primed) {
      clearAll();
      extort = emptyExtortTracker();
    }
    if (r.rewound) extort = emptyExtortTracker();
    if (r.fresh.length === 0) return;

    const folded = foldExtort(extort, r.fresh);
    extort = folded.tracker;
    // Drains that resolve in one frame are one scale.
    const drain = mergeDrains(folded.drains);
    if (!drain) return;
    const next = arrive(queue, drain);
    queue = next.queue;
    if (next.start) start(next.start);
  }

  // The first frame primes; a prime-key change primes the next one. The
  // key effect is declared first, so a key and a log that change together
  // prime that same frame.
  $effect(() => {
    void beatsPrimeKey;
    untrack(() => {
      reprimeNext = true;
    });
  });
  $effect(() => {
    const log = view.log;
    untrack(() => frame(log));
  });

  // Skip: a click or Escape, while a scene is up. Capture and passive, so
  // the click still reaches what it was aimed at.
  $effect(() => {
    if (!scene || typeof window === "undefined") return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") clearAll();
    };
    const onPointer = () => clearAll();
    window.addEventListener("keydown", onKey, true);
    window.addEventListener("pointerdown", onPointer, { capture: true, passive: true });
    return () => {
      window.removeEventListener("keydown", onKey, true);
      window.removeEventListener("pointerdown", onPointer, true);
    };
  });

  onDestroy(() => {
    destroyed = true;
    if (timer) clearTimeout(timer);
  });

  const caption = (n: number): string => (n > 1 ? L.effectExtortMany(n) : L.effectExtort);
  const beat = (ms: number): string => `${ms}ms`;
</script>

<div class="x-layer" aria-hidden="true" data-extort-scales>
  {#if scene}
    {@const sc = scene}
    <div
      class="x-scene"
      data-extort
      data-extort-motion={sc.motion}
      data-extort-count={sc.count}
      style:--k={sc.k}
      style:--ms={`${sc.ms}ms`}
      style:--soul-from={beat(EXTORT_BEATS.soulsFrom)}
      style:--soul-ms={beat(EXTORT_BEATS.soulsTo - EXTORT_BEATS.soulsFrom + 150)}
      style:--hit={beat(EXTORT_BEATS.hit)}
      style:--spark-from={beat(EXTORT_BEATS.sparkFrom)}
      style:--spark-ms={beat(EXTORT_BEATS.sparkTo - EXTORT_BEATS.sparkFrom)}
      style:--total={beat(EXTORT_MS)}
    >
      <div
        class="x-box"
        style:left={`${sc.box.x}px`}
        style:top={`${sc.box.y}px`}
        style:width={`${sc.box.width}px`}
        style:height={`${sc.box.height}px`}
      >
        <div class="x-glow"></div>
        <!-- Ian's mock: a blindfolded judge, golden scales, the angel over the dark pan
             (which is weighed down until the coin falls) and the demon over the light. -->
        <svg class="x-art" viewBox="0 0 360 320" overflow="visible">
          <defs>
            <linearGradient id="x-gold" x1="0" y1="0" x2="1" y2="1">
              <stop stop-color="#fff3c0" />
              <stop offset=".3" stop-color="#ecbd66" />
              <stop offset=".52" stop-color="#87511d" />
              <stop offset=".78" stop-color="#f8d78c" />
              <stop offset="1" stop-color="#8d5a25" />
            </linearGradient>
            <radialGradient id="x-void">
              <stop offset=".42" stop-color="#000" />
              <stop offset=".7" stop-color="#0c061c" />
              <stop offset=".89" stop-color="#563584" />
              <stop offset="1" stop-color="#bb88ff" />
            </radialGradient>
            <radialGradient id="x-star">
              <stop stop-color="#fff" />
              <stop offset=".25" stop-color="#fff8d7" />
              <stop offset=".65" stop-color="#f3ce75" />
              <stop offset="1" stop-color="#9b6536" stop-opacity="0" />
            </radialGradient>
          </defs>
          <circle
            cx="180"
            cy="147"
            r="139"
            fill="none"
            stroke="#b89b6530"
            stroke-dasharray="3 11"
          />
          <!-- The blindfolded judge -->
          <path
            d="M165 93Q159 69 169 54Q181 42 192 55Q201 73 194 93L200 108L208 142L197 179L217 265H143L164 179L151 142L160 108Z"
            fill="url(#x-gold)"
            stroke="#ffe5aa"
            stroke-width="1.5"
          />
          <ellipse cx="180" cy="80" rx="16" ry="21" fill="#d8b883" />
          <path
            d="M163 76Q180 70 197 76L198 85Q180 81 162 86Z"
            fill="#eee2cd"
            stroke="#a78a65"
            stroke-width="1.5"
          />
          <path
            d="M174 96Q180 101 186 96M166 60Q180 45 195 60"
            fill="none"
            stroke="#5a3b25"
            stroke-width="1.8"
          />
          <circle
            cx="180"
            cy="78"
            r="29"
            fill="none"
            stroke="#f5d28a"
            stroke-width="3"
            stroke-dasharray="7 4"
          />
          <path
            d="M169 121L180 137L191 121M180 137L180 226M154 268H206L218 282H142Z"
            fill="none"
            stroke="#fff0bc"
            stroke-width="3"
          />
          <g class="x-arms">
            <path
              d="M42 102Q180 119 318 102"
              fill="none"
              stroke="#513119"
              stroke-width="13"
              stroke-linecap="round"
            />
            <path
              d="M42 99Q180 116 318 99"
              fill="none"
              stroke="url(#x-gold)"
              stroke-width="9"
              stroke-linecap="round"
            />
            <!-- The angel, on the dark pan's side -->
            <g transform="translate(44 72)">
              <circle cx="0" cy="-24" r="14" fill="none" stroke="#ffedab" stroke-width="3" />
              <path
                d="M-6-8Q-38-45-47-23Q-42-1-11 1Q-33-5-36 9Q-20 17-7 5M6-8Q38-45 47-23Q42-1 11 1Q33-5 36 9Q20 17 7 5"
                fill="#f5d898"
                stroke="#fff4c5"
                stroke-width="1.5"
              />
              <circle cx="0" cy="-9" r="7" fill="#fbe4b7" />
              <path d="M-7 1L0 0L7 1L14 28H-14Z" fill="url(#x-gold)" stroke="#fff0bd" />
            </g>
            <!-- The demon, on the light pan's side -->
            <g transform="translate(316 72)">
              <path
                d="M-5-10Q-18-29-9-38L-1-24L7-24L15-39Q23-27 9-10"
                fill="#34223c"
                stroke="#a27bb7"
                stroke-width="1.5"
              />
              <path
                d="M-8 0Q-38-32-52-12L-35-7L-45 8L-19 4L-8 19M8 0Q38-32 52-12L35-7L45 8L19 4L8 19"
                fill="#29152e"
                stroke="#9267a6"
                stroke-width="1.5"
              />
              <ellipse cx="0" cy="-8" rx="8" ry="11" fill="#291522" />
              <path d="M-7 2L7 2L13 27H-13Z" fill="#39253e" stroke="#b082bf" />
              <circle cx="-3" cy="-9" r="1.6" fill="#e6a2ff" />
              <circle cx="3" cy="-9" r="1.6" fill="#e6a2ff" />
            </g>
            <path
              d="M65 103L34 194M65 103L96 194M295 103L264 194M295 103L326 194"
              stroke="#e7bb6d"
              stroke-width="2.5"
            />
            <path
              d="M17 192Q65 248 113 192Z"
              fill="url(#x-gold)"
              stroke="#ffe4a4"
              stroke-width="3"
            />
            <path d="M247 192Q295 248 343 192Z" fill="#51412f" stroke="#bd92d0" stroke-width="3" />
            <path
              d="M22 196Q65 211 108 196M252 196Q295 211 338 196"
              fill="none"
              stroke="#fff0b3"
              stroke-width="1.5"
            />
            <!-- Left: the light-sucking dark orb -->
            <g class="x-dark">
              <g class="x-orbit a">
                <ellipse
                  cx="65"
                  cy="185"
                  rx="36"
                  ry="13"
                  transform="rotate(-28 65 185)"
                  fill="none"
                  stroke="#ae91e7"
                  stroke-width="3"
                  stroke-dasharray="32 11 19 16"
                />
              </g>
              <g class="x-orbit b">
                <ellipse
                  cx="65"
                  cy="185"
                  rx="31"
                  ry="11"
                  transform="rotate(36 65 185)"
                  fill="none"
                  stroke="#efd18b"
                  stroke-width="2"
                  stroke-dasharray="26 14 16 15"
                />
              </g>
              <circle cx="65" cy="185" r="26" fill="url(#x-void)" />
              <circle cx="65" cy="185" r="16" fill="#000" />
              <circle cx="65" cy="185" r="10" fill="#000" />
              <g class="x-orbit a">
                <circle cx="29" cy="183" r="2.2" fill="#f3d18d" />
                <circle cx="91" cy="161" r="1.7" fill="#cba9fa" />
              </g>
            </g>
            <!-- Right: the radiant orb that bursts when the coin falls -->
            <g class="x-light">
              <circle cx="295" cy="185" r="31" fill="url(#x-star)" opacity=".55" />
              <g class="x-orbit a">
                <ellipse
                  cx="295"
                  cy="185"
                  rx="34"
                  ry="12"
                  transform="rotate(-24 295 185)"
                  stroke="#ffe7a2"
                  fill="none"
                  stroke-width="3"
                />
              </g>
              <g class="x-orbit b">
                <ellipse
                  cx="295"
                  cy="185"
                  rx="29"
                  ry="11"
                  transform="rotate(38 295 185)"
                  stroke="#fff5ce"
                  fill="none"
                  stroke-width="2"
                />
              </g>
              <circle cx="295" cy="185" r="21" fill="url(#x-star)" />
              <circle cx="295" cy="185" r="7" fill="#fffef0" />
              <path
                d="M295 144v22M295 204v22M254 185h22M314 185h22M266 156l16 16M308 198l16 16M324 156l-16 16M282 198l-16 16"
                stroke="#ffe9ae"
                stroke-width="2"
                opacity=".85"
              />
            </g>
          </g>
        </svg>
        <span class="x-cap" data-extort-caption>{caption(sc.count)}</span>
      </div>

      {#if sc.motion === "full"}
        <!-- The souls leaving each drained opponent, into the dark pan. -->
        {#each sc.souls as s, i (i)}
          <span class="x-ring m" style:left={`${s.at.x}px`} style:top={`${s.at.y}px`}></span>
          <span
            class="x-soul m"
            data-extort-soul
            style:left={`${sc.dark.x}px`}
            style:top={`${sc.dark.y}px`}
            style:--dx={`${s.dx}px`}
            style:--dy={`${s.dy}px`}
            style:--delay={`${s.delay}ms`}
          ></span>
        {/each}
        <!-- The light, once the coin has fallen, to the extort player. -->
        {#each sc.sparks as p, i (i)}
          <span
            class="x-spark m"
            data-extort-spark
            style:left={`${sc.light.x}px`}
            style:top={`${sc.light.y}px`}
            style:--dx={`${p.dx}px`}
            style:--dy={`${p.dy}px`}
          ></span>
          <span class="x-ring gain m" style:left={`${p.at.x}px`} style:top={`${p.at.y}px`}></span>
        {/each}
      {/if}
    </div>
  {/if}
</div>

<style>
  .x-layer {
    position: absolute;
    inset: 0;
    pointer-events: none;
    /* The table-effects rung of DESIGN.md §4's ladder: above the action
       dock (55), below the hints (58). */
    z-index: 57;
    overflow: hidden;
  }
  .x-scene {
    position: absolute;
    inset: 0;
  }
  .x-box {
    position: absolute;
    /* The whole scene's life: in, held, out (fades start at EXTORT_BEATS.fadeFrom). */
    animation: x-life calc(var(--ms)) ease both;
    filter: drop-shadow(0 0 10px rgba(198, 154, 76, 0.25));
  }
  /* A soft dark pool behind the art, so the gold reads on any skin. */
  .x-glow {
    position: absolute;
    inset: -14%;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      rgba(10, 8, 18, 0.78) 0%,
      rgba(10, 8, 18, 0.55) 48%,
      transparent 72%
    );
  }
  .x-art {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
  }
  .x-cap {
    position: absolute;
    left: 50%;
    bottom: -4px;
    translate: -50% 100%;
    padding: 2px 9px;
    border-radius: 999px;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    white-space: nowrap;
    color: #ffe1a0;
    background: rgba(16, 13, 23, 0.82);
    box-shadow: 0 0 0 1px rgba(134, 104, 53, 0.7);
  }

  /* x-life, in the scene's own clock (ms = EXTORT_MS x animations.speed): in over
     the first 6 %, out from EXTORT_BEATS.fadeFrom / EXTORT_MS (88.46 %). A static
     scene (--ms is its hold) is the same fade. */
  @keyframes x-life {
    0% {
      opacity: 0;
      scale: 0.92;
    }
    6% {
      opacity: 1;
      scale: 1;
    }
    88.46% {
      opacity: 1;
      scale: 1;
    }
    100% {
      opacity: 0;
      scale: 0.98;
    }
  }
  [data-extort-motion="still"] .x-box {
    animation-name: x-life-still;
  }
  @keyframes x-life-still {
    0% {
      opacity: 0;
    }
    12% {
      opacity: 1;
    }
    75% {
      opacity: 1;
    }
    100% {
      opacity: 0;
    }
  }

  /* The beam leans (the dark pan is down) until the coin: it starts to level at
     EXTORT_BEATS.soulsTo - 50 (32.69 %) and is level at EXTORT_BEATS.hit (48.08 %),
     with the mock's overshoot. Static scales are level. */
  .x-arms {
    transform-origin: 180px 102px;
    animation: x-level calc(var(--total) * var(--k)) both;
  }
  [data-extort-motion="still"] .x-arms {
    animation: none;
  }
  @keyframes x-level {
    0% {
      transform: rotate(-12deg);
    }
    32.69% {
      transform: rotate(-12deg);
      animation-timing-function: cubic-bezier(0.18, 0.86, 0.24, 1.22);
    }
    48.08% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(0deg);
    }
  }

  /* The dark orb pulses as it absorbs (four quick pulses from the souls'
     departure to the coin); the light orb bursts at the coin. */
  .x-dark {
    transform-origin: 65px 185px;
    animation: x-pulse calc(275ms * var(--k)) ease-in-out calc(var(--soul-from) * var(--k)) 4 both;
  }
  .x-light {
    transform-origin: 295px 185px;
    animation: x-burst calc(500ms * var(--k)) ease-out calc(var(--hit) * var(--k)) 1 both;
  }
  [data-extort-motion="still"] .x-dark,
  [data-extort-motion="still"] .x-light {
    animation: none;
  }
  @keyframes x-pulse {
    50% {
      transform: scale(1.14);
    }
  }
  @keyframes x-burst {
    0% {
      transform: scale(1);
      opacity: 0.75;
    }
    50% {
      transform: scale(1.55);
      opacity: 1;
    }
    100% {
      transform: scale(1);
      opacity: 1;
    }
  }
  .x-orbit {
    animation: x-spin 3.4s linear infinite;
  }
  .x-light .x-orbit.a,
  .x-light .x-orbit.b {
    transform-origin: 295px 185px;
  }
  .x-dark .x-orbit.a,
  .x-dark .x-orbit.b {
    transform-origin: 65px 185px;
  }
  .x-orbit.b {
    animation-direction: reverse;
    animation-duration: 4.8s;
  }
  [data-extort-motion="still"] .x-orbit {
    animation: none;
  }
  @keyframes x-spin {
    to {
      transform: rotate(360deg);
    }
  }

  /* A soul: leaves its player, flies into the dark pan (which is at its leaning
     place, the box's own geometry), and is absorbed. It is a zero-size anchor at
     the pan; --dx/--dy start it back at the player. */
  .x-soul {
    position: absolute;
    width: 22px;
    height: 22px;
    margin: -11px 0 0 -11px;
    border-radius: 50%;
    background: radial-gradient(circle, #fff, #ffe29c 45%, #aa792f);
    box-shadow: 0 0 18px 6px rgba(233, 187, 97, 0.6);
    animation: x-soul calc(var(--soul-ms) * var(--k)) cubic-bezier(0.45, 0, 0.25, 1)
      calc((var(--soul-from) + var(--delay)) * var(--k)) both;
  }
  @keyframes x-soul {
    0% {
      translate: var(--dx) var(--dy);
      scale: 0.3;
      opacity: 0;
    }
    12% {
      opacity: 1;
    }
    83% {
      translate: 0 0;
      scale: 0.75;
      opacity: 1;
    }
    100% {
      translate: 0 0;
      scale: 0.2;
      opacity: 0;
    }
  }

  /* The spark: leaves the light pan after the burst and lands on the extort
     player's avatar. */
  .x-spark {
    position: absolute;
    width: 12px;
    height: 12px;
    margin: -6px 0 0 -6px;
    border-radius: 50%;
    background: #fffce0;
    box-shadow: 0 0 18px 7px #eac06a;
    animation: x-spark calc(var(--spark-ms) * var(--k)) ease-in-out
      calc(var(--spark-from) * var(--k)) both;
  }
  @keyframes x-spark {
    0% {
      translate: 0 0;
      scale: 1;
      opacity: 0;
    }
    8% {
      opacity: 1;
    }
    85% {
      opacity: 1;
    }
    100% {
      translate: var(--dx) var(--dy);
      scale: 0.6;
      opacity: 0;
    }
  }

  /* A ring where a soul leaves (dark, at the drained player) and where the
     light lands (gold, at the extort player). */
  .x-ring {
    position: absolute;
    width: 56px;
    height: 56px;
    margin: -28px 0 0 -28px;
    border-radius: 50%;
    border: 2px solid rgba(174, 145, 231, 0.9);
    box-shadow: 0 0 14px rgba(120, 80, 190, 0.7);
    animation: x-ring calc(420ms * var(--k)) ease-out calc(var(--soul-from) * var(--k)) both;
  }
  .x-ring.gain {
    border-color: rgba(255, 233, 174, 0.95);
    box-shadow: 0 0 16px rgba(234, 192, 106, 0.8);
    animation-delay: calc((var(--spark-from) + var(--spark-ms) - 120ms) * var(--k));
  }
  @keyframes x-ring {
    0% {
      scale: 0.3;
      opacity: 0;
    }
    30% {
      opacity: 1;
    }
    100% {
      scale: 1.25;
      opacity: 0;
    }
  }

  /* The OS's reduced-motion signal (DESIGN.md §4, the third way): every
     moving piece goes and the static scale and its fade stay. The script
     reads the same signal and draws only the static scale, so this is the
     belt to its braces. */
  @media (prefers-reduced-motion: reduce) {
    .m {
      display: none;
    }
    .x-arms,
    .x-dark,
    .x-light,
    .x-orbit {
      animation: none;
    }
    .x-box {
      animation-name: x-life-still;
    }
  }
</style>

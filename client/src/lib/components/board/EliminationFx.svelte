<script lang="ts">
  // EliminationFx — a player leaves the game (#4449, ADR 0209).
  //
  // When the log says a player was eliminated, this layer plays the
  // player_eliminated sound once for that player, and shatters their
  // avatar like rock: its wedges fly apart over it, sparks fling out and
  // fall, a ring spreads, and the avatar shakes. The tombstone that then
  // settles on their panel is Tombstone.svelte, drawn from the seat's own
  // flag so it also survives a reload; this layer is only the moment.
  //
  // It works the same for the viewer and for an opponent, for each of up
  // to three players leaving in a four-player game, and for two leaving in
  // one frame (the second is heard a beat after the first).
  //
  // Presentation over the board's own state, never waited on (ADR 0209 §1).
  // The frame's log entries are tracked as TableEffects tracks them, so a
  // reconnect, a replay jump or an undo cues nothing already on the wire.
  // Sound always plays (a hidden tab included). Motion follows the three
  // reduced-motion signals (DESIGN.md §4): when any is on the burst draws
  // nothing at all, no wedges and no sparks, and the stone fades in.
  //
  // z 57, with the other effect cues: above the action dock (55), below the
  // hints (58). aria-hidden, and it takes no pointer events.

  import { onDestroy, untrack } from "svelte";
  import { get } from "svelte/store";
  import type { GameView } from "../../protocol";
  import { settings } from "../../settings";
  import { findSeatAnchor, findStandIn, hasSize } from "../../boardAnchor";
  import { seatColor } from "../../colors";
  import { impactShake } from "../../animations";
  import { play } from "../../sounds";
  import { emptyBeatTracker, track } from "../../combatBeats";
  import { effectMotion, effectsScheduled, type EffectMotion } from "../../tableEffects";
  import {
    BURST_MS,
    ELIMINATION_STAGGER_MS,
    burstShards,
    burstSparks,
    eliminatedSeats,
    rememberCommanders,
    type Shard,
    type Spark,
  } from "../../elimination";

  interface Props {
    view: GameView;
    boardEl: HTMLElement | null;
    // Changing this primes the next frame, as on a first frame: nothing
    // already on the wire is cued (Game.svelte's beatsPrimeKey).
    beatsPrimeKey?: string;
  }

  const { view, boardEl, beatsPrimeKey = "" }: Props = $props();

  interface Burst {
    key: string;
    seat: string;
    // The avatar's centre, in board pixels, and its width.
    x: number;
    y: number;
    d: number;
    color: string;
    ms: number;
    shards: Shard[];
    sparks: Spark[];
  }

  let bursts = $state<Burst[]>([]);
  let serial = 0;
  let destroyed = false;
  let tracker = emptyBeatTracker();
  let reprimeNext = false;
  const timers = new Set<ReturnType<typeof setTimeout>>();

  function later(fn: () => void, ms: number): void {
    const t = setTimeout(() => {
      timers.delete(t);
      if (!destroyed) fn();
    }, ms);
    timers.add(t);
  }

  function osReducesMotion(): boolean {
    return (
      typeof window !== "undefined" &&
      typeof window.matchMedia === "function" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches
    );
  }

  function clearAll(): void {
    for (const t of timers) clearTimeout(t);
    timers.clear();
    bursts = [];
  }

  // The seat's avatar, or the card that stands in for its board in the
  // reflow, as a box in board pixels.
  function avatarBox(seatID: string): Pick<Burst, "x" | "y" | "d"> | null {
    if (!boardEl) return null;
    const el = findSeatAnchor(boardEl, seatID, { accept: hasSize }) ?? findStandIn(boardEl, seatID);
    if (!el) return null;
    const b = boardEl.getBoundingClientRect();
    const r = el.getBoundingClientRect();
    if (r.width <= 0 && r.height <= 0) return null;
    return {
      x: r.left + r.width / 2 - b.left,
      y: r.top + r.height / 2 - b.top,
      d: Math.min(r.width, r.height) || Math.max(r.width, r.height),
    };
  }

  // One player's departure: the sound, then (when motion is on) the burst.
  function cue(seatIndex: number, motion: EffectMotion, speed: number): void {
    const seat = view.seats[seatIndex];
    if (!seat) return;
    // Once for this player, however often the frame is seen again.
    play("player_eliminated", { batch: `${view.id}:out:${seat.id}` });
    const visibility = typeof document === "undefined" ? undefined : document.visibilityState;
    if (motion === "still" || !effectsScheduled(visibility)) return;
    const at = avatarBox(seat.id);
    if (!at) return;
    const key = `out-${++serial}`;
    const ms = Math.round(BURST_MS * speed * (motion === "short" ? 0.6 : 1));
    bursts = [
      ...bursts,
      {
        key,
        seat: seat.id,
        ...at,
        color: seatColor(seat.seat),
        ms,
        shards: burstShards(),
        sparks: burstSparks(seat.id),
      },
    ];
    later(() => {
      bursts = bursts.filter((b) => b.key !== key);
    }, ms + 50);
    const el = boardEl ? (findSeatAnchor(boardEl, seat.id, { accept: hasSize }) ?? null) : null;
    if (el) void impactShake(el, { lethal: true });
  }

  function frame(log: GameView["log"]): void {
    const prime = reprimeNext;
    reprimeNext = false;
    const r = track(tracker, log, prime);
    tracker = r.tracker;
    // A reconnect or a replay move: forget what was playing.
    if (r.primed) clearAll();
    const gone = eliminatedSeats(r.fresh);
    if (gone.length === 0) return;

    const s = get(settings);
    const motion = effectMotion({
      enabled: s.animations.enabled,
      effects: s.animations.effects,
      reduceMotion: s.accessibility.reduceMotion,
      osReduceMotion: osReducesMotion(),
    });
    const speed = s.animations.speed;
    // Two players leaving in one frame are heard apart: play() drops a
    // repeat of one name inside its rate limit.
    gone.forEach((seatIndex, i) => {
      if (i === 0) cue(seatIndex, motion, speed);
      else later(() => cue(seatIndex, motion, speed), i * ELIMINATION_STAGGER_MS);
    });
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

  // Remember each seat's commander while it can be seen, for the tombstone
  // that names it once it has left with the player. This layer is mounted
  // for the whole game, whichever way a seat is drawn (a full panel or a
  // collapsed card), so it sees every frame the seat is still in.
  $effect(() => {
    for (const s of view.seats) rememberCommanders(view, s);
  });

  onDestroy(() => {
    destroyed = true;
    for (const t of timers) clearTimeout(t);
    timers.clear();
  });
</script>

<div class="elim-layer" aria-hidden="true" data-elimination-fx>
  {#each bursts as b (b.key)}
    <div
      class="burst"
      data-burst={b.seat}
      style:left="{b.x}px"
      style:top="{b.y}px"
      style:--d="{b.d}px"
      style:--ms="{b.ms}ms"
      style:--rock="color-mix(in srgb, {b.color} 55%, #4a4650)"
    >
      <span class="ring"></span>
      {#each b.shards as s, i (i)}
        <span
          class="shard"
          style:clip-path={s.clip}
          style:--dx={s.dx}
          style:--dy={s.dy}
          style:--rot="{s.rot}deg"
        ></span>
      {/each}
      {#each b.sparks as p, i (i)}
        <span
          class="spark"
          class:ember={p.ember}
          style:--dx={p.dx}
          style:--dy={p.dy}
          style:width="{p.size}px"
          style:height="{p.size}px"
        ></span>
      {/each}
    </div>
  {/each}
</div>

<style>
  .elim-layer {
    position: absolute;
    inset: 0;
    pointer-events: none;
    /* The table's z-index ladder (DESIGN.md §4): with the other effect
       cues, above the action dock (55), below the hints (58). */
    z-index: 57;
    overflow: hidden;
  }
  /* A zero-size anchor at the avatar's centre; every piece is placed from it. */
  .burst {
    position: absolute;
    width: 0;
    height: 0;
  }
  .ring {
    position: absolute;
    left: calc(var(--d) / -2);
    top: calc(var(--d) / -2);
    width: var(--d);
    height: var(--d);
    box-sizing: border-box;
    border: 3px solid rgba(255, 90, 40, 0.9);
    border-radius: 50%;
    animation: elim-ring var(--ms) cubic-bezier(0.215, 0.61, 0.355, 1) both;
  }
  .shard {
    position: absolute;
    left: calc(var(--d) / -2);
    top: calc(var(--d) / -2);
    width: var(--d);
    height: var(--d);
    border-radius: 50%;
    background: var(--rock);
    animation: elim-shard calc(var(--ms) * 0.75) cubic-bezier(0.25, 0.46, 0.45, 0.94) both;
  }
  .spark {
    position: absolute;
    left: 0;
    top: 0;
    border-radius: 50%;
    background: #ffaa00;
    animation: elim-spark calc(var(--ms) * 0.9) cubic-bezier(0.25, 0.46, 0.45, 0.94) both;
  }
  .spark.ember {
    background: #ff6622;
  }
  @keyframes elim-ring {
    from {
      transform: scale(0.3);
      opacity: 0.9;
    }
    to {
      transform: scale(2.6);
      opacity: 0;
    }
  }
  @keyframes elim-shard {
    from {
      transform: translate(0, 0) rotate(0deg) scale(1);
      opacity: 1;
    }
    to {
      transform: translate(calc(var(--dx) * var(--d)), calc(var(--dy) * var(--d)))
        rotate(var(--rot)) scale(0.2);
      opacity: 0;
    }
  }
  @keyframes elim-spark {
    from {
      transform: translate(-50%, -50%);
      opacity: 1;
    }
    to {
      transform: translate(calc(var(--dx) * var(--d)), calc(var(--dy) * var(--d)));
      opacity: 0;
    }
  }
  /* Reduced motion, the OS signal: the layer draws nothing (the script
     also never mounts a piece), and this keeps a stray one still. */
  @media (prefers-reduced-motion: reduce) {
    .ring,
    .shard,
    .spark {
      display: none;
    }
  }
</style>

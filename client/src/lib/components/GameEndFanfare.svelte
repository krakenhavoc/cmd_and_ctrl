<script lang="ts">
  // #2920: the game-end moment. A dialog over the table that names the
  // winner in their seat colour, with a confetti burst for the viewer's
  // win (or a spectator watching one), a quiet card for a draw or the
  // viewer's own loss, and two ways out: Back to lobby, or Keep looking
  // at the board (Escape does the same). The "game over" banner and the
  // dock's Back to lobby stay underneath; this only adds the moment.
  //
  // `motion` is the caller's verdict from the animation settings
  // (master switch on, not reduced motion). Without it the card appears
  // at once and nothing moves; sound is the existing win/loss cue.
  import { onMount } from "svelte";
  import type { Fanfare } from "../gameEndFanfare";
  import Icon from "./Icon.svelte";

  interface Props {
    fanfare: Fanfare;
    motion: boolean;
    onback: () => void;
    ondismiss: () => void;
  }
  let { fanfare, motion, onback, ondismiss }: Props = $props();

  // A fixed fan of pieces: position, delay and drift are arithmetic on
  // the index, so a render is the same every time and tests can count.
  const PIECES = 28;
  const pieces = Array.from({ length: PIECES }, (_, i) => ({
    left: ((i * 37) % 100) + 0.5,
    delay: (i % 7) * 90,
    dur: 1800 + ((i * 53) % 900),
    drift: ((i * 29) % 80) - 40,
    hue: i % 3,
  }));

  let keep = $state<HTMLButtonElement | null>(null);
  onMount(() => keep?.focus({ preventScroll: true }));

  function onKey(e: KeyboardEvent): void {
    if (e.key === "Escape") {
      e.stopPropagation();
      ondismiss();
    }
  }
</script>

<svelte:window onkeydown={onKey} />

<div
  class="fanfare-scrim"
  data-testid="game-end-fanfare"
  data-tone={fanfare.tone}
  data-motion={motion ? "1" : "0"}
  style={fanfare.color ? `--seat:${fanfare.color}` : ""}
>
  {#if motion && fanfare.celebrate}
    <div class="confetti" aria-hidden="true" data-testid="fanfare-confetti">
      {#each pieces as p, i (i)}
        <span
          class="piece hue{p.hue}"
          style="left:{p.left}%;animation-delay:{p.delay}ms;animation-duration:{p.dur}ms;--drift:{p.drift}px"
        ></span>
      {/each}
    </div>
  {/if}
  <div
    class="card"
    class:pop={motion}
    role="dialog"
    aria-modal="true"
    aria-labelledby="fanfare-headline"
    aria-describedby="fanfare-detail"
  >
    <div class="kicker">
      {#if fanfare.celebrate}
        <Icon name="crown" size={16} />
      {/if}
      {fanfare.kicker}
    </div>
    <h2 id="fanfare-headline" class="headline" class:seat={fanfare.color !== null}>
      {fanfare.headline}
    </h2>
    <p id="fanfare-detail" class="detail">{fanfare.detail}</p>
    <div class="actions">
      <button type="button" class="primary" onclick={onback}>Back to lobby</button>
      <button type="button" bind:this={keep} onclick={ondismiss}>Keep looking at the board</button>
    </div>
  </div>
</div>

<style>
  .fanfare-scrim {
    position: fixed;
    inset: 0;
    z-index: 60;
    display: grid;
    place-items: center;
    background: color-mix(in srgb, var(--bg) 72%, transparent);
    overflow: hidden;
  }
  .card {
    position: relative;
    min-width: min(420px, 90vw);
    max-width: 90vw;
    padding: 28px 36px;
    text-align: center;
    background: var(--surface-raised);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-lg);
  }
  [data-tone="own-win"] .card {
    border-color: var(--seat, var(--gold));
    box-shadow:
      0 0 48px color-mix(in srgb, var(--seat, var(--gold)) 45%, transparent),
      var(--shadow-lg);
  }
  [data-tone="watch"] .card {
    border-color: color-mix(in srgb, var(--seat, var(--gold)) 60%, transparent);
  }
  .kicker {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    font-size: 0.8rem;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--gold);
  }
  [data-tone="loss"] .kicker,
  [data-tone="draw"] .kicker {
    color: inherit;
    opacity: 0.7;
  }
  .headline {
    margin: 8px 0 4px;
    font-size: clamp(1.8rem, 6vw, 3rem);
    line-height: 1.1;
    overflow-wrap: anywhere;
  }
  .headline.seat {
    color: var(--seat);
  }
  [data-tone="own-win"] .headline {
    font-size: clamp(2.2rem, 8vw, 3.8rem);
  }
  .detail {
    margin: 0 0 20px;
    opacity: 0.85;
  }
  .actions {
    display: flex;
    gap: 10px;
    justify-content: center;
    flex-wrap: wrap;
  }
  .confetti {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }
  .piece {
    position: absolute;
    top: -12px;
    width: 8px;
    height: 14px;
    border-radius: 2px;
    background: var(--seat, var(--gold));
    opacity: 0;
    animation: fall linear 1 forwards;
  }
  .piece.hue1 {
    background: var(--gold);
  }
  .piece.hue2 {
    background: var(--text, #fff);
  }
  .pop {
    animation: pop 420ms cubic-bezier(0.2, 1.4, 0.4, 1) both;
  }
  @keyframes pop {
    from {
      transform: scale(0.7);
      opacity: 0;
    }
    to {
      transform: scale(1);
      opacity: 1;
    }
  }
  @keyframes fall {
    0% {
      opacity: 1;
      transform: translate3d(0, 0, 0) rotate(0deg);
    }
    100% {
      opacity: 0;
      transform: translate3d(var(--drift), 105vh, 0) rotate(540deg);
    }
  }
  /* The caller already turns motion off for the setting; this covers a
     device that asks before settings load. */
  @media (prefers-reduced-motion: reduce) {
    .confetti {
      display: none;
    }
    .pop {
      animation: none;
    }
  }
</style>

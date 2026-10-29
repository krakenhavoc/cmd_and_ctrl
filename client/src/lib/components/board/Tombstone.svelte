<script lang="ts">
  // Tombstone — the stone that marks a player who has left the game
  // (#4449, ADR 0209). It stays on that player's panel for the rest of the
  // game and carries the epitaph, their name and their commander.
  //
  // It is drawn from the seat's own `eliminated` flag, so a reload, a
  // reconnect or a seat that was out before this page opened shows it at
  // once, still. Only a seat this component SEES go out arrives: the stone
  // drops in once the avatar's burst (EliminationFx.svelte) has played.
  // Reduced motion, all three ways (DESIGN.md §4): with the master switch
  // off, animations.effects Off, accessibility.reduceMotion or the OS
  // signal, the stone does not drop; it only fades in. That fade is opacity
  // alone, so it keeps its length when app.css cuts every other animation.
  //
  // `compact` is the small stone in a collapsed seat's card; otherwise the
  // stone sits in the middle of the seat's whole panel, over the (now empty)
  // board. It takes no pointer events, so nothing under it is blocked.

  import { untrack } from "svelte";
  import { get } from "svelte/store";
  import type { GameView, PlayerView } from "../../protocol";
  import { settings } from "../../settings";
  import { L } from "../../labels";
  import {
    TOMBSTONE_DELAY_MS,
    TOMBSTONE_DROP_MS,
    TOMBSTONE_FADE_MS,
    commanderLine,
    rememberCommanders,
    tombstoneCommanders,
  } from "../../elimination";
  import { effectMotion, type EffectMotion } from "../../tableEffects";

  interface Props {
    seat: PlayerView;
    view: GameView;
    compact?: boolean;
  }

  const { seat, view, compact = false }: Props = $props();

  const eliminated = $derived(seat.eliminated === true);
  const name = $derived(seat.display_name ?? seat.name);

  // Remember the commander while it can be seen: the frame that eliminates
  // the seat no longer shows it.
  $effect(() => {
    rememberCommanders(view, seat);
  });
  const commander = $derived(eliminated ? commanderLine(tombstoneCommanders(view, seat)) : "");

  function osReducesMotion(): boolean {
    return (
      typeof window !== "undefined" &&
      typeof window.matchMedia === "function" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches
    );
  }

  // Whether this stone arrives (the seat went out while it watched) or was
  // there when it mounted. Arriving is decided once, at the moment the flag
  // turns true, with the motion the settings allow then.
  let seen = untrack(() => eliminated);
  let arrival = $state<{ motion: EffectMotion; delayMs: number } | null>(null);
  $effect(() => {
    if (!eliminated) {
      seen = false;
      arrival = null;
      return;
    }
    if (seen) return;
    seen = true;
    const s = get(settings);
    const motion = effectMotion({
      enabled: s.animations.enabled,
      effects: s.animations.effects,
      reduceMotion: s.accessibility.reduceMotion,
      osReduceMotion: osReducesMotion(),
    });
    // A still stone has no burst to wait for.
    arrival = {
      motion,
      delayMs: motion === "still" ? 0 : Math.round(TOMBSTONE_DELAY_MS * s.animations.speed),
    };
  });
</script>

{#if eliminated}
  <div
    class="tomb"
    class:compact
    class:arrives={arrival !== null}
    class:still={arrival?.motion === "still"}
    data-tombstone={seat.id}
    data-motion={arrival?.motion}
    style:--tomb-delay={arrival ? `${arrival.delayMs}ms` : undefined}
    style:--tomb-drop={`${TOMBSTONE_DROP_MS}ms`}
    style:--tomb-fade={`${TOMBSTONE_FADE_MS}ms`}
    role="group"
    aria-label={L.tombstoneOf(name)}
  >
    <div class="stone">
      <svg class="cross" viewBox="0 0 12 16" aria-hidden="true">
        <path d="M5 0h2v4h4v2H7v10H5V6H1V4h4z" fill="currentColor" />
      </svg>
      <span class="rip">{L.tombstoneRip}</span>
      <span class="who">{name}</span>
      {#if commander}
        <span class="commander">{commander}</span>
      {/if}
    </div>
  </div>
{/if}

<style>
  /* Over the panel's own contents (its rows and playmat), under nothing a
     player has to reach: it takes no pointer events. */
  .tomb {
    position: absolute;
    inset: 0;
    z-index: 20;
    /* Clear of an avatar at the top of a short panel. */
    padding-top: 44px;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: none;
    container-type: inline-size;
  }
  /* In a collapsed seat's card the stone sits in the flow. */
  .tomb.compact {
    position: static;
    inset: auto;
    z-index: auto;
    container-type: normal;
    flex: 1 1 auto;
    min-height: 0;
    padding-top: 0;
  }
  .stone {
    --stone-hi: color-mix(in srgb, var(--fg) 22%, var(--surface-sunken));
    --stone-lo: color-mix(in srgb, var(--fg) 6%, var(--surface-sunken));
    box-sizing: border-box;
    width: min(180px, 60%);
    min-width: 132px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 14px 12px 12px;
    text-align: center;
    background: linear-gradient(180deg, var(--stone-hi) 0%, var(--stone-lo) 70%);
    border: 2px solid color-mix(in srgb, var(--fg) 28%, var(--surface-sunken));
    border-bottom-width: 5px;
    border-radius: 64px 64px 6px 6px;
    box-shadow:
      0 10px 28px rgba(0, 0, 0, 0.7),
      inset 0 2px 3px color-mix(in srgb, var(--fg) 18%, transparent);
    color: var(--fg);
  }
  .compact .stone {
    width: 100%;
    min-width: 0;
    max-width: 128px;
    gap: 0;
    padding: 5px 6px 2px;
    border-bottom-width: 4px;
    border-radius: 28px 28px 4px 4px;
    box-shadow: 0 4px 10px rgba(0, 0, 0, 0.55);
  }
  .cross {
    width: 12px;
    height: 16px;
    color: var(--fg-muted);
  }
  /* A small card has no room for the cross; the epitaph says it. */
  .compact .cross {
    display: none;
  }
  .rip {
    font-size: 1.15rem;
    font-weight: 800;
    letter-spacing: 0.22em;
    /* The letter-spacing trails the last letter; pull the text back to centre. */
    margin-right: -0.22em;
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.7);
  }
  .compact .rip {
    font-size: 0.7rem;
  }
  .who {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.85rem;
    font-weight: 700;
    color: var(--fg);
  }
  .compact .who {
    font-size: 0.68rem;
  }
  .commander {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.72rem;
    font-style: italic;
    color: var(--fg-muted);
  }
  .compact .commander {
    font-size: 0.6rem;
  }

  /* A stone that arrives waits for the burst, then drops and settles. The
     first keyframe holds it hidden through the delay. Without an arrival
     (a reload, a reconnect) there is no animation at all. */
  .tomb.arrives .stone {
    animation: tomb-drop var(--tomb-drop) var(--tomb-delay, 0ms)
      cubic-bezier(0.175, 0.885, 0.32, 1.1) both;
  }
  @keyframes tomb-drop {
    from {
      opacity: 0;
      transform: translateY(-70px) scale(0.85);
    }
    55% {
      opacity: 1;
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  /* Still: no drop and no scale, the stone only fades in. */
  .tomb.arrives.still .stone {
    animation: tomb-fade var(--tomb-fade) 0ms ease-out both;
  }
  @keyframes tomb-fade {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  /* The OS signal and the setting both cut every animation (app.css, and
     the media block below for the OS), so the fade is restated here with
     more weight: opacity alone is no motion, and the stone should not just
     appear. */
  :global(:root[data-reduce-motion="1"]) .tomb.arrives .stone {
    animation: tomb-fade var(--tomb-fade) 0ms ease-out both !important;
  }
  @media (prefers-reduced-motion: reduce) {
    .tomb.arrives .stone {
      animation: tomb-fade var(--tomb-fade) 0ms ease-out both !important;
    }
  }
</style>

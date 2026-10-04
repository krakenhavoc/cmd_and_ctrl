<script lang="ts">
  // PhaseIcon renders a small pictogram for one MTG turn step.
  // Used by PhaseDisplay.svelte to replace the row of identical
  // 8px dots with something that tells you the phase at a glance.
  //
  // All glyphs draw in currentColor on a 16×16 viewBox so the
  // caller can tint them via CSS (the active step inherits the
  // active player's avatar-derived colour; inactive steps read a
  // muted chrome colour). Shapes are deliberately simple —
  // silhouette-first, no strokes where a fill works, because at
  // 12px inside a panel even a two-pixel stroke reads as noise.
  //
  // #2214 redrew the set as one family, in the groups PhaseDisplay
  // draws (CR 500.1's five phases):
  //   beginning   refresh arrow, hourglass, a card with an up arrow
  //   main 1 / 2  ONE glyph, a hand playing a card, numbered I or II
  //               on the card's face
  //   combat      crossed swords, one sword, shield, lightning, burst,
  //               crossed scabbards (the first and last are a pair:
  //               the same X, drawn and then put away)
  //   ending      moon, broom
  // The fill is set once on the <svg>; a stroked shape opts out with
  // fill="none". Keep every arc at least half its chord long
  // (phaseIcon.render.test.ts, #1620).

  import type { StepID } from "../../turn";

  interface Props {
    step: StepID;
  }

  const { step }: Props = $props();
</script>

<svg
  class="phase-icon"
  viewBox="0 0 16 16"
  aria-hidden="true"
  focusable="false"
  fill="currentColor"
  data-step={step}
>
  {#if step === "untap"}
    <!-- Circular arrow: everything turns upright again. -->
    <path
      d="M13 8 A5 5 0 1 1 8 3"
      fill="none"
      stroke="currentColor"
      stroke-width="1.7"
      stroke-linecap="round"
    />
    <polygon points="7,0.6 10.6,3 7,5.4" />
  {:else if step === "upkeep"}
    <!-- Hourglass with its caps: upkeep is the "time passes" beat. -->
    <rect x="3" y="1" width="10" height="1.6" rx="0.6" />
    <rect x="3" y="13.4" width="10" height="1.6" rx="0.6" />
    <path
      d="M4.4 2.6 H11.6 Q11.6 5.6 8.9 8 Q11.6 10.4 11.6 13.4 H4.4 Q4.4 10.4 7.1 8 Q4.4 5.6 4.4 2.6 Z"
    />
  {:else if step === "draw"}
    <!-- A card with an up arrow beside it: a card comes to you. -->
    <rect x="1.8" y="4.6" width="7.4" height="10.6" rx="1.2" />
    <path
      d="M12.6 14.6 V4"
      fill="none"
      stroke="currentColor"
      stroke-width="1.8"
      stroke-linecap="round"
    />
    <polygon points="9.6,5.6 12.6,1 15.6,5.6" />
  {:else if step === "precombat_main" || step === "postcombat_main"}
    <!-- The two main phases share one glyph: a hand playing a card.
         The card is an outline so the hand reads over it, and its
         face carries the phase's number: I before combat, II after. -->
    <rect
      x="6.2"
      y="1.3"
      width="7"
      height="9.6"
      rx="1"
      fill="none"
      stroke="currentColor"
      stroke-width="1.4"
      transform="rotate(14 9.7 6.1)"
    />
    <path
      d="M1.2 15.5 V11.6 Q1.2 9.6 3 8.9 L6.6 7.4 Q8.2 6.9 8.6 8.1 Q9 9.3 7.6 9.9 L6.6 10.3 H11.2 Q12.4 10.3 12.4 11.5 Q12.4 12.4 11.4 12.6 Q12.2 12.9 12.1 13.8 Q12 15.5 10 15.5 Z"
    />
    {#if step === "precombat_main"}
      <g class="numeral" data-numeral="I" transform="rotate(14 9.7 6.1)">
        <rect x="9" y="3.2" width="1.4" height="4.2" />
      </g>
    {:else}
      <g class="numeral" data-numeral="II" transform="rotate(14 9.7 6.1)">
        <rect x="7.9" y="3.2" width="1.3" height="4.2" />
        <rect x="10.2" y="3.2" width="1.3" height="4.2" />
      </g>
    {/if}
  {:else if step === "begin_combat" || step === "end_combat"}
    <!-- The pair. Beginning of combat: two swords crossed, points up,
         weapons drawn. End of combat: the same X with the blades in
         their scabbards and hung hilt-up, put away. -->
    {#each step === "begin_combat" ? [45, -45] : [135, -135] as angle (angle)}
      <g transform="rotate({angle} 8 8)">
        {#if step === "begin_combat"}
          <path d="M8 0.8 L9.1 2.6 V10 H6.9 V2.6 Z" />
        {:else}
          <rect x="6.5" y="1.2" width="3" height="8.8" rx="1.5" />
        {/if}
        <rect x="4.9" y="10" width="6.2" height="1.5" rx="0.6" />
        <rect x="7.25" y="11.5" width="1.5" height="2.4" />
        <circle cx="8" cy="14.5" r="1.15" />
      </g>
    {/each}
  {:else if step === "declare_attackers"}
    <!-- One sword, pointing forward: the attack. -->
    <g transform="rotate(45 8 8) scale(1.06) translate(-0.45 -0.45)">
      <path d="M8 0.8 L9.1 2.6 V10 H6.9 V2.6 Z" />
      <rect x="4.9" y="10" width="6.2" height="1.5" rx="0.6" />
      <rect x="7.25" y="11.5" width="1.5" height="2.4" />
      <circle cx="8" cy="14.5" r="1.15" />
    </g>
  {:else if step === "declare_blockers"}
    <!-- Shield. -->
    <path d="M8 1.2 L14 3.4 V8 Q14 12.4 8 14.8 Q2 12.4 2 8 V3.4 Z" />
  {:else if step === "first_strike_damage"}
    <!-- Lightning: the damage that lands first (CR 510.4). -->
    <path d="M10 0.8 L3.2 9.2 H7.4 L5.8 15.2 L12.8 6.6 H8.6 Z" />
  {:else if step === "combat_damage"}
    <!-- Impact burst: nine uneven points, so it reads as a hit rather
         than a star. -->
    <polygon
      points="8.00,0.80 8.99,5.47 11.99,3.45 10.51,6.75 15.29,6.92 10.86,8.70 13.37,11.30 9.86,10.42 10.53,15.15 8.00,11.10 5.88,14.03 6.14,10.42 1.59,11.90 5.14,8.70 1.89,7.12 5.49,6.75 3.24,2.53 7.01,5.47"
    />
  {:else if step === "end"}
    <!-- Crescent moon — end step, "nightfall". -->
    <!-- The inner arc must reach the chord (12 units): a radius
         under 6 is scaled up to exactly 6, which retraces the outer
         arc and draws nothing (#1620 — the end step's icon was
         invisible). Outer r=7 large arc, inner r=6 half circle. -->
    <path d="M11 2 A7 7 0 1 0 11 14 A6 6 0 0 1 11 2 Z" />
  {:else if step === "cleanup"}
    <!-- Broom: cleanup sweeps the turn away. -->
    <g transform="rotate(-38 8 8)">
      <rect x="7.3" y="-0.6" width="1.4" height="8.6" rx="0.7" />
      <rect x="5.6" y="8" width="4.8" height="1.6" rx="0.5" />
      <path d="M5.4 10.2 H10.6 L12 15.4 Q8 16.4 4 15.4 Z" />
    </g>
  {/if}
</svg>

<style>
  .phase-icon {
    width: 100%;
    height: 100%;
    display: block;
    color: inherit;
    pointer-events: none;
    overflow: visible;
  }
</style>

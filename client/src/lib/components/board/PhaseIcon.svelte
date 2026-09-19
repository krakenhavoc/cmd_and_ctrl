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

  import type { StepID } from "../../turn";

  interface Props {
    step: StepID;
  }

  const { step }: Props = $props();
</script>

<svg class="phase-icon" viewBox="0 0 16 16" aria-hidden="true" focusable="false">
  {#if step === "untap"}
    <!-- Circular refresh arrow. Open at the top-right so the
         arrowhead reads as "return to start of turn". -->
    <path
      d="M13 8 A5 5 0 1 1 8 3"
      fill="none"
      stroke="currentColor"
      stroke-width="1.6"
      stroke-linecap="round"
    />
    <polygon points="7,1 10,3 7,5" fill="currentColor" />
  {:else if step === "upkeep"}
    <!-- Hourglass — upkeep is the "time passes" beat. -->
    <path d="M4.5 2.5 H11.5 L8 8 L11.5 13.5 H4.5 L8 8 Z" fill="currentColor" />
  {:else if step === "draw"}
    <!-- Card with a small down-pointing chevron, reads as "draw a card". -->
    <rect x="4.5" y="2.5" width="7" height="10" rx="1.2" fill="currentColor" opacity="0.85" />
    <path
      d="M6 14 L8 15.5 L10 14"
      fill="none"
      stroke="currentColor"
      stroke-width="1.4"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
  {:else if step === "precombat_main"}
    <!-- Sun — precombat main is "midday: plan your turn". -->
    <circle cx="8" cy="8" r="2.8" fill="currentColor" />
    <g stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
      <line x1="8" y1="1.5" x2="8" y2="3.2" />
      <line x1="8" y1="12.8" x2="8" y2="14.5" />
      <line x1="1.5" y1="8" x2="3.2" y2="8" />
      <line x1="12.8" y1="8" x2="14.5" y2="8" />
      <line x1="3.3" y1="3.3" x2="4.6" y2="4.6" />
      <line x1="11.4" y1="11.4" x2="12.7" y2="12.7" />
      <line x1="3.3" y1="12.7" x2="4.6" y2="11.4" />
      <line x1="11.4" y1="4.6" x2="12.7" y2="3.3" />
    </g>
  {:else if step === "begin_combat"}
    <!-- Single upright sword — combat begins, weapon drawn. -->
    <path d="M8 1 L9 3 V10 H10.5 V11.5 H9 V14 H7 V11.5 H5.5 V10 H7 V3 Z" fill="currentColor" />
  {:else if step === "declare_attackers"}
    <!-- Crossed swords. -->
    <g fill="currentColor">
      <path d="M1.5 3 L3 1.5 L13 11.5 V14 H10.5 Z" />
      <path d="M14.5 3 L13 1.5 L3 11.5 V14 H5.5 Z" />
    </g>
  {:else if step === "declare_blockers"}
    <!-- Shield. -->
    <path d="M8 1.5 L14 3.5 V8 Q14 12.2 8 14.5 Q2 12.2 2 8 V3.5 Z" fill="currentColor" />
  {:else if step === "first_strike_damage"}
    <!-- The impact burst again, drawn small and offset up-left, with a
         speed line behind it: the same event as combat damage, but
         first. Reads as a sibling of the step below, which is what
         CR 510.4's two combat damage steps are. -->
    <path d="M6 2 L7.2 6 L11 7.2 L7.2 8.4 L6 12.4 L4.8 8.4 L1 7.2 L4.8 6 Z" fill="currentColor" />
    <path
      d="M9.5 11 L14 15"
      fill="none"
      stroke="currentColor"
      stroke-width="1.8"
      stroke-linecap="round"
    />
  {:else if step === "combat_damage"}
    <!-- Impact burst — 8-point star. -->
    <path d="M8 1 L9.5 6.5 L15 8 L9.5 9.5 L8 15 L6.5 9.5 L1 8 L6.5 6.5 Z" fill="currentColor" />
  {:else if step === "end_combat"}
    <!-- Chevron — "moving on from combat". -->
    <path
      d="M5 3 L10 8 L5 13"
      fill="none"
      stroke="currentColor"
      stroke-width="2.2"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
  {:else if step === "postcombat_main"}
    <!-- Half-sun — the "afternoon" main phase. Visually echoes
         precombat_main without duplicating it exactly. -->
    <path d="M2.5 11 A5.5 5.5 0 0 1 13.5 11 Z" fill="currentColor" />
    <line
      x1="1.5"
      y1="12.2"
      x2="14.5"
      y2="12.2"
      stroke="currentColor"
      stroke-width="1.5"
      stroke-linecap="round"
    />
    <g stroke="currentColor" stroke-width="1.3" stroke-linecap="round">
      <line x1="2.2" y1="8.5" x2="3.4" y2="8.5" />
      <line x1="12.6" y1="8.5" x2="13.8" y2="8.5" />
      <line x1="4" y1="5.4" x2="5" y2="6.2" />
      <line x1="11" y1="6.2" x2="12" y2="5.4" />
    </g>
  {:else if step === "end"}
    <!-- Crescent moon — end step, "nightfall". -->
    <path d="M11 2 A6 6 0 1 0 11 14 A4.5 4.5 0 1 1 11 2 Z" fill="currentColor" />
  {:else if step === "cleanup"}
    <!-- 4-point sparkle — cleanup "sweeps" the turn. -->
    <path d="M8 1.5 L9 7 L14.5 8 L9 9 L8 14.5 L7 9 L1.5 8 L7 7 Z" fill="currentColor" />
  {/if}
</svg>

<style>
  .phase-icon {
    width: 100%;
    height: 100%;
    display: block;
    color: inherit;
    pointer-events: none;
  }
</style>

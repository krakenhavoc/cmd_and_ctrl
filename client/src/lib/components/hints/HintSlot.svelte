<script lang="ts">
  // HintSlot — where a modal dialog draws its own hints (ADR 0125 §3.6).
  //
  // Settings is a modal that traps focus, so a hint whose place is
  // "settings" renders inside the dialog, through this slot, and its
  // buttons are inside the trap. HintLayer decides what shows and where;
  // the slot only draws it. Renders nothing when no settings hint is up.

  import { activeTip } from "../../hints/runtime";
  import TipCard from "./TipCard.svelte";
</script>

{#if $activeTip && $activeTip.hint.place === "settings"}
  <TipCard tip={$activeTip} />
{/if}

<script lang="ts">
  // A Card under a board's combat damage clock (ADR 0134 §2), for the
  // render test of the hide-while-the-copy-flies rule.
  import { untrack } from "svelte";
  import Card from "../components/board/Card.svelte";
  import { provideCombatCues, type CombatCues } from "../combatCues.svelte";
  import type { CardView } from "../protocol";

  const { cues, card }: { cues: CombatCues; card: CardView } = $props();
  provideCombatCues(untrack(() => cues));
</script>

<Card {card} />

<script lang="ts">
  // BoardAttentionStub stands in for Board.svelte in Game-level render
  // tests that need the attention strip (ADR 0111 PR 3): it renders the
  // `attention` snippet Game.svelte passes, inside a stand-in strip, and
  // a button per card id that asks Game to select it for combat, which
  // is what a click on a creature does on the real board.
  import type { Snippet } from "svelte";
  import { L } from "../labels";

  const {
    attention,
    onSelectCombatCard,
    selectCards = ["b1"],
  }: {
    attention?: Snippet;
    onSelectCombatCard?: (cardID: string) => void;
    selectCards?: string[];
  } = $props();
</script>

<div data-testid="board-stub">
  <div class="strip" data-testid="strip" role="region" aria-label={L.attention}>
    {@render attention?.()}
  </div>
  {#each selectCards as id (id)}
    <button type="button" data-testid={`select-${id}`} onclick={() => onSelectCombatCard?.(id)}
      >select {id}</button
    >
  {/each}
</div>

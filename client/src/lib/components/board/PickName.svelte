<script lang="ts">
  // PickName is the name under one tile of a card-pick sheet (the
  // discard to hand size, the bounded card grid, the scry grid). The
  // art alone does not tell two similar cards apart, so the name sits
  // beneath the card: one line, truncated with an ellipsis, with the
  // full name in the title for a hover or a long press.
  //
  // It draws nothing for a face-down card, whose name the viewer may not
  // know, and the caller's button keeps its own accessible name.

  import type { CardView } from "../../protocol";
  import { displayName } from "../../faces";

  interface Props {
    card: CardView;
  }

  const { card }: Props = $props();

  const name = $derived(card.face_down ? "" : displayName(card));
</script>

{#if name}
  <span class="pick-name" title={name}>{name}</span>
{/if}

<style>
  .pick-name {
    display: block;
    width: 100%;
    margin-top: 4px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 11px;
    line-height: 1.3;
    text-align: center;
    color: var(--fg);
  }
</style>

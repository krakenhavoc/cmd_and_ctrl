<script lang="ts">
  // DraftPreview — the large card beside the card being hovered, focused
  // or held (#5848 §2). One host per surface: it draws only the previews
  // whose cards carry its `scope`, so the pack and the picks panel can
  // each have one and a swap from one to the other is still a swap.
  //
  // It sits beside the card, never on it: to the right, flipped to the
  // left when there is no room, and centred above or below on a phone
  // (lib/draftPreview.svelte.ts placePreview). It takes no pointer
  // events and is hidden from a screen reader: the card's own control
  // already says everything the preview shows.
  //
  // `children` is drawn under the image with the card, for what a
  // surface wants to say about it (the reasons for its rating).
  import type { Snippet } from "svelte";
  import { tick } from "svelte";

  import type { DraftCard } from "../../draft";
  import { placePreview, preview } from "../../draftPreview.svelte";
  import DraftCardArt from "./DraftCardArt.svelte";

  interface Props {
    scope: string;
    children?: Snippet<[DraftCard]>;
  }
  const { scope, children }: Props = $props();

  const card = $derived(preview.scope === scope ? preview.card : null);
  const anchor = $derived(preview.scope === scope ? preview.anchor : null);

  let el = $state<HTMLElement | null>(null);

  // Placed once the card is in the DOM and measured, before it paints; a
  // swap to another card moves it at once.
  $effect(() => {
    const box = el;
    const a = anchor;
    void card;
    if (!box || !a) return;
    const place = (): void => {
      const { x, y } = placePreview(
        a,
        { w: box.offsetWidth, h: box.offsetHeight },
        { vw: window.innerWidth, vh: window.innerHeight },
      );
      box.style.transform = `translate(${x}px, ${y}px)`;
    };
    place();
    // The image and the extra lines can change the height a moment later.
    void tick().then(place);
  });
</script>

{#if card}
  <div
    class="pv"
    bind:this={el}
    aria-hidden="true"
    data-testid="draft-preview"
    data-card-id={card.id}
  >
    <DraftCardArt {card} />
    {#if children}
      <div class="extra" data-testid="draft-preview-extra">{@render children(card)}</div>
    {/if}
  </div>
{/if}

<style>
  .pv {
    position: fixed;
    left: 0;
    top: 0;
    /* 340px is the width the owner chose; a narrow window shrinks it. */
    width: min(340px, calc(100vw - 16px));
    z-index: 60;
    pointer-events: none;
    filter: drop-shadow(0 18px 28px rgba(0, 0, 0, 0.55));
  }
  .extra {
    margin-top: 8px;
  }
</style>

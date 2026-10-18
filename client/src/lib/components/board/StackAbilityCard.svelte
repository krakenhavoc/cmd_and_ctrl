<script lang="ts">
  // StackAbilityCard — a triggered or activated ability on the stack,
  // drawn as a card of C2's own (ADR 0119, amendment 2026-10-16 §B, §E).
  //
  // An ability has no card (CR 405.1), so this is not a Magic card frame
  // and imitates none: it is a steel panel in a card's 63:88 proportions,
  // so it sits in the pile like a spell does.
  //
  //   - the source's art crop across the top, its name over the art's
  //     lower edge in the display voice. A source the viewer may not see
  //     (`previewCard` null, #697) shows the card back's pattern and the
  //     name the model gives;
  //   - the cost badge on the art's top-left corner: a loyalty ability's
  //     signed cost, an activated ability's symbols ({T} included). None
  //     for a triggered ability (lib/stackAbilityCard.ts);
  //   - the kind line in the mono voice: what the rules call it;
  //   - the verbatim text in the ui voice at reading size, scrolling
  //     inside the card when it is long.
  //
  // Presentation only, and only the face: the caller draws the frame
  // around it (the seat band, rings, the Target button, hover, Counter),
  // sizes it, and owns the item's title and data attributes. That keeps
  // it reusable wherever the pile draws an item: its top slot, and the
  // cascade's lower cards.

  import type { StackLaneItem } from "../../stackLane";
  import { abilityCardKind, abilityCostOf, abilityKindLine } from "../../stackAbilityCard";
  import { cardImageURL } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import ManaCost from "./ManaCost.svelte";

  interface Props {
    item: StackLaneItem;
  }

  const { item }: Props = $props();

  const kind = $derived(abilityCardKind(item) ?? "activated");
  const cost = $derived(abilityCostOf(item));
  const art = $derived(item.previewCard ? cardImageURL(item.previewCard, "art_crop") : null);
  const hidden = $derived(item.previewCard === null);
  const sourceName = $derived(
    item.previewCard?.name || item.sourceName || (hidden ? "face-down card" : item.name),
  );
  // Short text reads larger; a long one steps down before it scrolls.
  const length = $derived(
    item.text.length <= 90 ? "short" : item.text.length <= 200 ? "medium" : "long",
  );
  /** What a screen reader reads for the art: what the eye reads on the card. */
  const alt = $derived(`${sourceName}. ${item.text}`);
</script>

<div class="ability-card" data-ability-kind={kind} class:hidden>
  <div class="art">
    {#if art}
      <img class="art-img" src={art} {alt} decoding="async" use:cardArt={art} />
    {:else}
      <img class="art-img back" src="/card-back.jpg" {alt} decoding="async" />
    {/if}
    <span class="source-name">{sourceName}</span>
    {#if cost}
      <!-- Decorative: the text below starts with the same cost. -->
      <span class="cost-badge" data-cost-kind={cost.kind} aria-hidden="true">
        {#if cost.kind === "loyalty"}
          {cost.value}
        {:else}
          <ManaCost cost={cost.cost} size={16} />
        {/if}
      </span>
    {/if}
  </div>
  <div class="kind-line">{abilityKindLine(kind)}</div>
  <div class="ability-text" data-length={length}>{item.text}</div>
</div>

<style>
  .ability-card {
    container-type: inline-size;
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    background: var(--surface);
    color: var(--fg);
  }

  /* ---- the art, the source's name, the badge ---- */
  .art {
    position: relative;
    flex: 0 0 auto;
    aspect-ratio: 16 / 11;
    overflow: hidden;
    background: var(--surface-sunken);
    border-bottom: 1px solid var(--border-strong);
  }
  .art-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  /* The card back's own pattern, its centre: no art to show. */
  .art-img.back {
    object-position: center 45%;
    transform: scale(1.6);
  }
  .source-name {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    padding: 22px 10px 6px 13px;
    font-family: var(--font-display);
    font-weight: 700;
    font-size: clamp(12px, 6cqi, 17px);
    line-height: 1.15;
    letter-spacing: -0.01em;
    color: var(--fg);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    /* A scrim of the card's own surface under the name, so it reads on
       any art and in any skin. */
    background: linear-gradient(
      to top,
      color-mix(in srgb, var(--surface) 94%, transparent) 0%,
      color-mix(in srgb, var(--surface) 78%, transparent) 55%,
      transparent 100%
    );
  }
  .cost-badge {
    position: absolute;
    top: 7px;
    left: 11px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 28px;
    height: 24px;
    padding: 0 6px;
    box-sizing: border-box;
    border-radius: var(--radius-sm);
    background: var(--surface-raised);
    border: 1px solid var(--seat-color, var(--border-strong));
    box-shadow: var(--shadow-sm);
    font-family: var(--font-mono);
    font-weight: 700;
    font-size: 13px;
    font-variant-numeric: tabular-nums;
    line-height: 1;
    color: var(--fg);
  }

  /* ---- the kind line ---- */
  .kind-line {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px 0 13px;
    font-family: var(--font-mono);
    font-weight: 700;
    font-size: 9.5px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-muted);
    white-space: nowrap;
  }
  .kind-line::after {
    content: "";
    flex: 1 1 auto;
    height: 1px;
    background: var(--border-strong);
  }

  /* ---- the text ---- */
  .ability-text {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
    overscroll-behavior: contain;
    scrollbar-width: thin;
    padding: 6px 10px 10px 13px;
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    line-height: 1.4;
    font-weight: 500;
    color: var(--fg);
    overflow-wrap: anywhere;
    /* #3830: a modal ability's "•" lines each start a line. */
    white-space: pre-line;
  }
  .ability-text[data-length="short"] {
    font-size: 0.9375rem;
    line-height: 1.38;
  }
  .ability-text[data-length="long"] {
    font-size: 0.75rem;
    line-height: 1.35;
  }
  /* The pile's shrunk top card (160px) while the viewer picks a target. */
  @container (max-width: 200px) {
    .ability-text,
    .ability-text[data-length] {
      font-size: 0.6875rem;
      line-height: 1.3;
      padding: 4px 6px 6px 10px;
    }
    .kind-line {
      font-size: 8px;
      padding: 5px 6px 0 10px;
    }
    .source-name {
      padding: 14px 6px 4px 10px;
    }
    .cost-badge {
      top: 5px;
      left: 9px;
      height: 20px;
      min-width: 24px;
      font-size: 11px;
    }
  }
</style>

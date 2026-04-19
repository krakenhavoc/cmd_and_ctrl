<script lang="ts">
  // PileButton is the small stacked-corner control for the four
  // pile-style zones each player owns: EXILE, GRAVEYARD, DECK
  // (library), CMD ZONE. The face is a label + count plus an
  // optional thumbnail of the top card (used by GRAVEYARD where
  // the top card is public). DECK uses a stylised back; EXILE and
  // CMD ZONE show counts only at v1.
  //
  // Click is dispatched up to the parent so PileBar can route it —
  // the only wired action at v1 is "draw from your own library";
  // GRAVEYARD / EXILE / CMD ZONE clicks are stubbed to no-op until
  // a pile-browser modal lands in a follow-up.

  import type { CardView, ZoneView } from "../../protocol";
  import Card from "./Card.svelte";

  interface Props {
    label: string;
    zone: ZoneView;
    // faceDown: render the pile as a card back (LIBRARY) regardless
    // of whether the top card is technically known. GRAVEYARD / EXILE
    // are face-up; CMD ZONE is face-up too (commanders are public).
    faceDown?: boolean;
    disabled?: boolean;
    onClick?: () => void;
  }

  const { label, zone, faceDown = false, disabled = false, onClick }: Props = $props();

  const topCard = $derived(zone.cards.length > 0 ? zone.cards[zone.cards.length - 1] : null);

  // Synthetic CardView for the face-down render path. Card.svelte
  // ignores name / scryfall_id / etc. when faceDown=true, so the
  // placeholder is purely structural. Derived because `label` is a
  // reactive prop — wrapping it in $derived keeps the instance_id
  // stable across renders for the same pile.
  const backPlaceholder = $derived<CardView>({
    instance_id: `${label}-back`,
    name: "",
    owner: "",
    controller: "",
  });
</script>

<button
  type="button"
  class="pile"
  class:disabled
  class:has-cards={zone.count > 0}
  {disabled}
  onclick={onClick}
  aria-label={`${label}: ${zone.count} card${zone.count === 1 ? "" : "s"}`}
  title={`${label} · ${zone.count}`}
>
  <span class="thumb">
    {#if faceDown && zone.count > 0}
      <Card card={backPlaceholder} faceDown />
    {:else if topCard && !faceDown}
      <Card card={topCard} />
    {:else}
      <span class="empty" aria-hidden="true"></span>
    {/if}
  </span>
  <span class="meta">
    <span class="label">{label}</span>
    <span class="count">{zone.count}</span>
  </span>
</button>

<style>
  .pile {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-start;
    gap: 2px;
    padding: 4px;
    background: #111a2b;
    border: 1px solid #2e3a55;
    border-radius: 6px;
    color: inherit;
    font: inherit;
    cursor: pointer;
    /* Width / thumb dimensions read from inherited CSS vars so a
       parent panel (e.g. an opponent slot) can shrink the whole bar
       without per-component prop plumbing. */
    width: var(--pile-w, 64px);
    box-sizing: border-box;
  }
  .pile:disabled,
  .pile.disabled {
    cursor: default;
    opacity: 0.6;
  }
  .pile:not(:disabled):hover {
    border-color: #5fb0ff;
    background: #18243a;
  }
  .pile:focus-visible {
    outline: 2px solid #5fb0ff;
    outline-offset: 2px;
  }
  .thumb {
    width: var(--thumb-w, 50px);
    height: var(--thumb-h, 70px);
    display: flex;
    align-items: center;
    justify-content: center;
    /* Re-publish the thumb size as --card-w/--card-h so the embedded
       Card.svelte (which reads those vars) sizes down to the pile's
       footprint without overriding its intrinsic 80×112 default. */
    --card-w: var(--thumb-w, 50px);
    --card-h: var(--thumb-h, 70px);
  }
  .empty {
    width: 100%;
    height: 100%;
    border-radius: 4px;
    border: 1px dashed #2e3a55;
  }
  .meta {
    display: flex;
    flex-direction: column;
    align-items: center;
    line-height: 1.1;
  }
  .label {
    font-size: 8px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #6c7a99;
  }
  .count {
    font-size: 13px;
    font-weight: 700;
    color: #e0e6f5;
  }
</style>

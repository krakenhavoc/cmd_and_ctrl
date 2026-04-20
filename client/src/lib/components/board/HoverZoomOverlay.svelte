<script lang="ts">
  // HoverZoomOverlay is the floating card-detail panel anchored at
  // (0.8, 0.2) of the board area — top-right quadrant. Driven by the
  // module-scope `hoveredCard` store written from Card.svelte.
  //
  // S11 grew the overlay from a bare image preview into a full Oracle
  // text + game-context panel:
  //   - Image (existing — Scryfall normal size, fills the top half)
  //   - Mana cost / type line / P/T or loyalty (header strip)
  //   - Oracle text (scrollable for long cards like Cyclonic Rift)
  //   - Game context: tapped state, controller name, attached counters,
  //     attack / block declarations, goad source, commander tax
  //
  // Card metadata (Oracle text, type line, etc.) comes from /cards/{id}
  // via cardMetaCache (lazy on first hover, in-memory thereafter); the
  // game-context fields read straight off the live CardView from the
  // hoveredCard store.

  import { hoveredCard } from "../../cardTypes";
  import { metaFor, type CardMeta } from "../../cardMetaCache";

  const card = $derived($hoveredCard);
  const imgSrc = $derived(
    card?.scryfall_id ? `/cards/${card.scryfall_id}/image?size=normal` : null,
  );

  // Subscribe to the card's metadata store. Re-derived per hover so a
  // new card swaps in a fresh subscription. The inner $-prefixed deref
  // happens inside $derived.by so Svelte tracks reactivity correctly.
  const metaStore = $derived(card?.scryfall_id ? metaFor(card.scryfall_id) : null);
  let meta = $state<CardMeta | null>(null);
  $effect(() => {
    if (!metaStore) {
      meta = null;
      return;
    }
    const unsub = metaStore.subscribe((v) => {
      meta = v;
    });
    return unsub;
  });

  // Counters → "+1/+1 ×3, -1/-1 ×1" rendering
  const counterChips = $derived(
    card?.counters
      ? Object.entries(card.counters)
          .filter(([, n]) => n > 0)
          .map(([name, n]) => ({ name, n }))
      : [],
  );

  // Display P/T with counter modifiers when present so the panel shows
  // the live combat value, not just the printed line. Matches the
  // server's CurrentPower() math (printed +/- counters, clamped at 0).
  const livePT = $derived.by(() => {
    if (!card || card.power == null || card.toughness == null) return null;
    const plusOnes = card.counters?.["+1/+1"] ?? 0;
    const minusOnes = card.counters?.["-1/-1"] ?? 0;
    const liveP = Math.max(0, card.power + plusOnes - minusOnes);
    const liveT = Math.max(0, card.toughness + plusOnes - minusOnes);
    const printed = `${card.power}/${card.toughness}`;
    const live = `${liveP}/${liveT}`;
    return live === printed ? printed : `${live} (${printed})`;
  });
</script>

{#if card}
  <div class="overlay" aria-hidden="true">
    {#if imgSrc}
      <img src={imgSrc} alt="" />
    {:else}
      <div class="name-fallback">{card.name}</div>
    {/if}
    <div class="info">
      <header class="info-head">
        <span class="title">{meta?.name ?? card.name}</span>
        {#if meta?.mana_cost}
          <span class="cost">{meta.mana_cost}</span>
        {/if}
      </header>
      {#if meta?.type_line}
        <div class="type-line">{meta.type_line}</div>
      {/if}
      {#if meta?.oracle_text}
        <div class="oracle">{meta.oracle_text}</div>
      {:else if meta === null}
        <div class="oracle dim">…</div>
      {/if}
      <footer class="info-foot">
        {#if livePT}
          <span class="pt" title="live power/toughness (printed in parens)">{livePT}</span>
        {:else if meta?.loyalty}
          <span class="loyalty" title="starting loyalty">◈ {meta.loyalty}</span>
        {/if}
        {#if card.tapped}
          <span class="state state-tapped">tapped</span>
        {/if}
        {#if card.attacking_target}
          <span class="state state-attack">attacking</span>
        {/if}
        {#if card.blocking_target}
          <span class="state state-block">blocking</span>
        {/if}
        {#if card.goaded_by}
          <span class="state state-goad">goaded</span>
        {/if}
        {#if card.is_commander}
          <span class="state state-cmd">commander</span>
        {/if}
        {#each counterChips as chip (chip.name)}
          <span class="counter-chip">{chip.name} ×{chip.n}</span>
        {/each}
      </footer>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: absolute;
    /* Anchor centre at (0.8, 0.2) of the board area. translate(-50%,
       -50%) re-centres around the chosen anchor so the visual centre,
       not the top-left, sits at the percentage. */
    left: 80%;
    top: 20%;
    transform: translate(-50%, -50%);
    width: clamp(220px, 26vw, 380px);
    background: #0b1220;
    border: 1px solid #4a5270;
    border-radius: 12px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.6);
    pointer-events: none;
    z-index: 40;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    color: #e0e6f5;
    font-size: 12px;
  }
  img {
    width: 100%;
    aspect-ratio: 63 / 88;
    object-fit: contain;
    background: #0b1220;
  }
  .name-fallback {
    width: 100%;
    aspect-ratio: 63 / 88;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 12px;
    box-sizing: border-box;
    font-size: 14px;
    background: #1a2335;
  }
  .info {
    padding: 8px 10px 10px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-height: 0;
    border-top: 1px solid #1a2335;
  }
  .info-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 8px;
  }
  .title {
    font-weight: 700;
    font-size: 13px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .cost {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 11px;
    color: #c8a86a;
    flex: 0 0 auto;
  }
  .type-line {
    font-size: 11px;
    color: #9ec7ff;
    text-transform: lowercase;
  }
  .oracle {
    /* Long cards (Cyclonic Rift, Counterspell) get a scroll once we
       hit ~7 lines so the overlay doesn't grow taller than the
       viewport. clamp height keeps the panel proportional. */
    max-height: clamp(80px, 18vh, 220px);
    overflow-y: auto;
    font-size: 11px;
    line-height: 1.35;
    color: #e0e6f5;
    white-space: pre-wrap;
    background: #07101e;
    padding: 6px 8px;
    border-radius: 4px;
    /* Suppress the default scrollbar in the floating overlay; the panel
       is non-interactive (pointer-events: none on the wrapper) so a
       scrollbar would just be visual noise. Long oracle text fades
       out at the bottom edge to hint there's more. */
    scrollbar-width: none;
    mask-image: linear-gradient(to bottom, black 85%, transparent 100%);
  }
  .oracle.dim {
    color: #6c7a99;
    font-style: italic;
  }
  .info-foot {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    align-items: center;
    margin-top: 2px;
  }
  .pt,
  .loyalty {
    font-weight: 700;
    color: #ffd07a;
    font-variant-numeric: tabular-nums;
    margin-right: 4px;
  }
  .state {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    padding: 1px 5px;
    border-radius: 3px;
    background: rgba(0, 0, 0, 0.35);
    color: #c8c8c8;
  }
  .state-tapped {
    color: #9ec7ff;
  }
  .state-attack {
    color: #ff7a7a;
  }
  .state-block {
    color: #9ec7ff;
  }
  .state-goad {
    color: #ff7a7a;
    background: rgba(80, 0, 0, 0.4);
  }
  .state-cmd {
    color: #ffd07a;
  }
  .counter-chip {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    padding: 1px 5px;
    border-radius: 3px;
    background: #1a2335;
    color: #c8a86a;
    border: 1px solid #2e3a55;
  }
</style>

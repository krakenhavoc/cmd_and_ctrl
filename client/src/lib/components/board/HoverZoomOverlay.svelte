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
  import type { GameView, PlayerView } from "../../protocol";
  import { playerColor } from "../../avatarColor";
  import { seatColor } from "../../colors";

  interface Props {
    view: GameView;
  }
  const { view }: Props = $props();

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

  // Commander-damage rows — only computed (and rendered) when the
  // hovered card is a commander. Reads `commander_damage[instance_id]`
  // on every OTHER seat so partner / token-copy commanders each show
  // their own instance-scoped totals. The controller is excluded per
  // CR (a player can't deal commander damage to themselves) and to
  // avoid a muddy self-row.
  const playerColors = $state<Record<string, string>>({});
  $effect(() => {
    if (!card?.is_commander) return;
    for (const s of view.seats) {
      const id = s.id;
      playerColors[id] = playerColor(s, (c) => {
        playerColors[id] = c;
      });
    }
  });
  const colorFor = (seat: PlayerView): string => playerColors[seat.id] ?? seatColor(seat.seat);
  const cmdrRows = $derived.by(() => {
    if (!card?.is_commander) return [];
    const instanceID = card.instance_id;
    const controllerID = card.controller;
    return view.seats
      .filter((s) => s.id !== controllerID)
      .map((s) => ({ seat: s, amount: s.commander_damage?.[instanceID] ?? 0 }));
  });
  const cmdrHasDamage = $derived(cmdrRows.some((r) => r.amount > 0));

  // Live P/T comes straight off the wire: since S16 the server sends
  // CurrentPower()/CurrentToughness() — effective P/T with +1/+1 and
  // -1/-1 counter deltas already baked in — so re-adding counters
  // here would double-count them. The printed parenthetical reads the
  // Scryfall metadata (string-typed; handles "*" stats) and only
  // renders once the meta fetch lands and the values actually differ.
  const livePT = $derived.by(() => {
    if (!card || card.power == null || card.toughness == null) return null;
    const live = `${card.power}/${card.toughness}`;
    const printed =
      meta?.power != null && meta?.toughness != null ? `${meta.power}/${meta.toughness}` : null;
    return printed !== null && printed !== live ? `${live} (${printed})` : live;
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
      {#if card.is_commander}
        <section class="cmdr-dmg" aria-label="commander damage dealt">
          <header class="cmdr-head">
            <span class="cmdr-dot" aria-hidden="true"></span>
            <span class="cmdr-label">cmdr dmg dealt</span>
          </header>
          {#if cmdrHasDamage}
            <ul class="cmdr-rows">
              {#each cmdrRows as row (row.seat.id)}
                <li class="cmdr-row" class:zero={row.amount === 0}>
                  <span
                    class="cmdr-name"
                    style:--seat-color={colorFor(row.seat)}
                    title={row.seat.name}
                  >
                    <span class="cmdr-name-dot" aria-hidden="true"></span>
                    <span class="cmdr-name-text">{row.seat.name}</span>
                  </span>
                  <span class="cmdr-amount" class:lethal={row.amount >= 21}>
                    {row.amount}
                  </span>
                </li>
              {/each}
            </ul>
          {:else}
            <div class="cmdr-empty">no commander damage dealt yet</div>
          {/if}
        </section>
      {/if}
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: absolute;
    /* Pin to the top-right corner of the board with a small inset so
       the panel always fits inside the viewport regardless of width. */
    right: 12px;
    top: 12px;
    width: clamp(220px, 26vw, 380px);
    max-height: calc(100vh - 24px);
    background: linear-gradient(180deg, rgba(19, 26, 44, 0.92) 0%, rgba(8, 12, 24, 0.92) 100%);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid rgba(122, 167, 255, 0.22);
    border-radius: var(--radius-lg);
    box-shadow:
      0 12px 40px rgba(0, 0, 0, 0.6),
      0 0 0 1px rgba(0, 0, 0, 0.4),
      inset 0 1px 0 rgba(255, 255, 255, 0.06);
    pointer-events: none;
    /* 300 sits above the modal-backdrop layer (200) so the preview
       isn't blurred by backdrop-filter when the user is picking
       cards in a modal (Discard / Choice / deck-import). */
    z-index: 300;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    color: var(--fg);
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
    font-size: 14px;
    letter-spacing: -0.01em;
    color: var(--fg);
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
  /* Commander-damage section — lives at the bottom of the info
     panel when the hovered card is a commander. Replaces the
     free-floating CommanderDamageTooltip so the information is
     anchored to the card the user is actually inspecting. */
  .cmdr-dmg {
    margin-top: 8px;
    padding-top: 6px;
    border-top: 1px solid #1a2335;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .cmdr-head {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .cmdr-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--gold, #c8a86a);
    display: inline-block;
  }
  .cmdr-label {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--fg-dim, #6c7a99);
    font-weight: 700;
  }
  .cmdr-rows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .cmdr-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 1px 0;
    font-size: 11px;
  }
  .cmdr-row.zero {
    opacity: 0.55;
  }
  .cmdr-name {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    min-width: 0;
    color: var(--fg);
  }
  .cmdr-name-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--seat-color, #888);
    flex: 0 0 auto;
  }
  .cmdr-name-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 14ch;
  }
  .cmdr-amount {
    font-variant-numeric: tabular-nums;
    font-weight: 700;
    color: #c8a86a;
    min-width: 2ch;
    text-align: right;
  }
  .cmdr-amount.lethal {
    color: #ff7a7a;
    text-shadow: 0 0 6px rgba(255, 122, 122, 0.55);
  }
  .cmdr-row.zero .cmdr-amount {
    color: var(--fg-dim, #6c7a99);
    font-weight: 500;
  }
  .cmdr-empty {
    font-size: 10px;
    color: var(--fg-dim, #6c7a99);
    font-style: italic;
    padding: 2px 0;
  }
</style>

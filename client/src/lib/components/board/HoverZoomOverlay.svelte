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
  import { cardImageURL } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import { metaFor, type CardMeta } from "../../cardMetaCache";
  import type { GameView, PlayerView } from "../../protocol";
  import { playerColor } from "../../avatarColor";
  import { seatColor } from "../../colors";
  import { noUntapFooterLines } from "../../noUntap";
  import { chosenValueChips } from "../../chosenValues";

  interface Props {
    view: GameView;
  }
  const { view }: Props = $props();

  const card = $derived($hoveredCard);
  const imgSrc = $derived(cardImageURL(card, "normal"));
  // The OTHER face of a multi-face card, if there is one — the back
  // of a modal DFC being previewed from hand, or the front of one
  // already flipped onto the battlefield. Null for the ~33,000
  // single-faced oracle IDs, which is what hides the panel.
  const otherFace = $derived.by(() => {
    const faces = card?.faces;
    if (!faces || faces.length < 2) return null;
    const active = card?.active_face ?? 0;
    const idx = active === 0 ? 1 : 0;
    return { index: idx, face: faces[idx], src: cardImageURL(card, "normal", idx) };
  });

  // Subscribe to the card's metadata store. Re-subscribed per hovered
  // PRINTING, not per hovered card: two copies of one card share a
  // scryfall_id and so share a store, and re-subscribing to the store
  // we are already on would only make the panel flicker.
  //
  // #740 — the derived holds the ID, and `metaFor` is called from the
  // EFFECT. It used to be the other way round (`$derived(… metaFor(id)
  // …)`), and that is what froze a board: `metaFor` wrote to the store
  // it was about to return, the write ran this subscriber, the
  // subscriber assigned to `meta` below, and a `$state` write inside a
  // running `$derived` is `state_unsafe_mutation`. Before the #720
  // store guards that throw escaped into svelte/store's shared
  // subscriber queue and stalled every store in the app — the board
  // stuck on one frame while the socket kept delivering.
  //
  // `metaFor` is pure again (see cardMetaCache.ts), so this is belt
  // and braces. It is also where the call belongs: it kicks a fetch,
  // and a derived may be evaluated, discarded and re-evaluated at the
  // renderer's convenience.
  const metaID = $derived(card?.scryfall_id ?? null);
  let meta = $state<CardMeta | null>(null);
  $effect(() => {
    if (!metaID) {
      meta = null;
      return;
    }
    const unsub = metaFor(metaID).subscribe((v) => {
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
  //
  // This lookup was written against a server that did not exist yet:
  // until S25 (#77) the server keyed `commander_damage` by the
  // OPPOSING PLAYER's ID, so every row here read 0 no matter how hard
  // a commander had connected. S25 rekeyed the map to commander
  // instance IDs (CR 903.10a), which is what this code always wanted.
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
  const noUntapLines = $derived(card ? noUntapFooterLines(card, view.seats) : []);
  // #781 — the chosen colour / creature type, through the one module
  // that turns "G" into "Green". The panel is where a player comes to
  // ask what a permanent does, and CR 607.2d makes this half of the
  // answer for anything with a "the chosen …" clause.
  const chosen = $derived(card ? chosenValueChips(card) : []);
  // Bars are tinted with the commander's controller colour (the seat
  // dealing the damage) and flip to danger at lethal.
  const controllerColor = $derived(
    seatColor(view.seats.find((s) => s.id === card?.controller)?.seat ?? 0),
  );

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
    <div class="scan">
      {#if imgSrc}
        <!-- keyed so a new card never shows the previous card's scan
             while its own image is still loading. The failure pip
             is display-only here: the panel is pointer-events: none
             and aria-hidden, and it closes as soon as the pointer
             leaves the card it previews. fetchpriority (#33): the
             panel exists to be read the moment it opens; the small
             other-face inset below does not need the hint. -->
        {#key imgSrc}
          <img
            src={imgSrc}
            alt=""
            fetchpriority="high"
            use:cardArt={{ url: imgSrc, interactive: false }}
          />
        {/key}
      {:else}
        <div class="name-fallback">{card.name}</div>
      {/if}
      <!-- ADR 0034: the other printed face. A modal DFC in hand is
           two playable objects, and the half you are NOT looking at
           is exactly the information the hover panel exists to
           supply. Absent for every single-faced card. -->
      {#if otherFace?.src}
        <div class="other-face" title={otherFace.face.name}>
          {#key otherFace.src}
            <img
              src={otherFace.src}
              alt={otherFace.face.name}
              use:cardArt={{ url: otherFace.src, interactive: false }}
            />
          {/key}
        </div>
      {/if}
    </div>
    <div class="info">
      <header class="info-head">
        <span class="title">{meta?.name ?? card.name}</span>
        {#if livePT}
          <span class="pt" title="live power/toughness (printed in parens)">{livePT}</span>
        {:else if meta?.loyalty}
          <span class="pt" title="starting loyalty">◈ {meta.loyalty}</span>
        {/if}
      </header>
      <div class="type-line">
        {#if meta?.type_line}<span>{meta.type_line}</span>{/if}
        {#if meta?.mana_cost}<span class="cost">{meta.mana_cost}</span>{/if}
      </div>
      {#if meta?.oracle_text}
        <div class="oracle">{meta.oracle_text}</div>
      {:else if meta === null}
        <div class="oracle dim">…</div>
      {/if}
      <!-- ADR 0093 Decision 8: the abilities OTHER permanents gave this
           one, each as its granting card prints it, one row per
           grantor. The oracle text above cannot say them — the card
           never printed them — and a granted trigger has no menu row,
           so this is the only place it is visible. -->
      {#if card.granted_abilities && card.granted_abilities.length > 0}
        <ul class="granted" aria-label="granted abilities">
          {#each card.granted_abilities as g, i (i)}
            <li>
              <span class="granted-text">{g.text}</span>
              {#if g.source_name}<span class="granted-from">from {g.source_name}</span>{/if}
            </li>
          {/each}
        </ul>
      {/if}
      <!-- Directly under the oracle text, which is the text it is
           about. Inspecting a card is the one moment a player is
           already asking what it does, so it costs nothing to answer
           the other half of the question here. -->
      {#if card.unimplemented}
        <div class="not-implemented">rules not implemented — resolve this card by hand</div>
      {/if}
      {#if card.tapped || card.attacking_target || card.blocking_target || card.goaded_by || card.is_commander || counterChips.length > 0 || noUntapLines.length > 0 || chosen.length > 0}
        <footer class="info-foot">
          {#if card.is_commander}
            <span class="state state-cmd">commander</span>
          {/if}
          <!-- #781: first in the footer, ahead of tapped / attacking.
               The others describe what is happening to the permanent
               right now; this one is what the rest of its printed text
               MEANS, and a player reading the oracle line above needs
               it to finish the sentence. -->
          {#each chosen as chip (chip.kind)}
            <span class="state state-chosen" title={chip.title}>
              {chip.kind === "color" ? "chosen color" : "chosen type"}: {chip.label}
            </span>
          {/each}
          {#if card.tapped}
            <span class="state">tapped</span>
          {/if}
          {#each noUntapLines as line}
            <span class="state">{line}</span>
          {/each}
          {#if card.attacking_target}
            <span class="state state-attack">attacking</span>
          {/if}
          {#if card.blocking_target}
            <span class="state">blocking</span>
          {/if}
          {#if card.goaded_by}
            <span class="state state-attack">goaded</span>
          {/if}
          {#each counterChips as chip (chip.name)}
            <span class="state">{chip.name} ×{chip.n}</span>
          {/each}
        </footer>
      {/if}
      {#if card.is_commander}
        <section class="cmdr-dmg" aria-label="commander damage dealt">
          <span class="cmdr-label">commander damage dealt</span>
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
                  <span class="bar-track">
                    <i
                      style:width={`${Math.min(100, (row.amount / 21) * 100)}%`}
                      style:background={row.amount >= 21 ? "var(--danger)" : controllerColor}
                    ></i>
                  </span>
                  <span class="cmdr-amount" class:lethal={row.amount >= 21}>{row.amount}</span>
                </li>
              {/each}
            </ul>
            <div class="cmdr-empty">lethal at 21</div>
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
    right: 10px;
    top: 10px;
    width: min(332px, calc(50% - 132px));
    max-height: calc(100% - 20px);
    background: color-mix(in srgb, var(--surface) 96%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid var(--border-strong);
    border-radius: 14px;
    box-shadow: var(--shadow-lg);
    padding: 12px;
    box-sizing: border-box;
    pointer-events: none;
    /* 300 sits above the modal-backdrop layer (200) so the preview
       isn't blurred by backdrop-filter when the user is picking
       cards in a modal (Discard / Choice / deck-import). */
    z-index: 300;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    gap: 10px;
    color: var(--fg);
    font-size: 12px;
  }
  .scan {
    width: min(240px, 100%);
    aspect-ratio: 63 / 88;
    align-self: center;
    border-radius: 12px;
    overflow: hidden;
    background: var(--surface-sunken);
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.6);
    flex: 0 0 auto;
    position: relative;
  }

  /* The other printed face, tucked into the bottom-right corner of
     the scan as a small inset — present enough to read, small
     enough not to compete with the face that is actually up. */
  .other-face {
    position: absolute;
    right: 6px;
    bottom: 6px;
    width: 38%;
    aspect-ratio: 63 / 88;
    border-radius: 6px;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.35);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.7);
  }

  .other-face img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .name-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 12px;
    box-sizing: border-box;
    font-size: 14px;
    color: var(--fg-muted);
  }
  .info {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-height: 0;
  }
  .info-head {
    display: flex;
    align-items: center;
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
  .pt {
    font-family: var(--font-mono);
    font-size: 12px;
    font-weight: 700;
    color: var(--fg);
    flex: 0 0 auto;
    font-variant-numeric: tabular-nums;
  }
  .type-line {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    font-size: 11.5px;
    color: var(--fg-muted);
  }
  .cost {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--fg-dim);
    flex: 0 0 auto;
  }
  .oracle {
    /* Long cards (Cyclonic Rift, Counterspell) get a scroll once we
       hit ~7 lines so the overlay doesn't grow taller than the
       viewport. clamp height keeps the panel proportional. */
    max-height: clamp(80px, 18vh, 220px);
    overflow-y: auto;
    font-size: 11.5px;
    line-height: 1.4;
    color: var(--fg);
    white-space: pre-wrap;
    background: var(--surface-sunken);
    padding: 8px 10px;
    border-radius: 8px;
    margin-top: 4px;
    /* Suppress the default scrollbar in the floating overlay; the panel
       is non-interactive (pointer-events: none on the wrapper) so a
       scrollbar would just be visual noise. Long oracle text fades
       out at the bottom edge to hint there's more. */
    scrollbar-width: none;
    mask-image: linear-gradient(to bottom, black 85%, transparent 100%);
  }
  .oracle.dim {
    color: var(--fg-dim);
    font-style: italic;
  }
  /* ADR 0093: granted abilities, read like oracle text but set apart
     from it — the card did not print them. */
  .granted {
    list-style: none;
    margin: 4px 0 0;
    padding: 6px 10px;
    border-radius: 8px;
    border: 1px dashed var(--border, var(--fg-dim));
    font-size: 11px;
    line-height: 1.35;
    color: var(--fg);
  }
  .granted li + li {
    margin-top: 3px;
  }
  .granted-from {
    margin-left: 6px;
    color: var(--fg-dim);
    font-size: 10px;
  }
  /* Deliberately quiet — dim, small, no colour of its own. This is a
     statement about the engine, not about the card, and it sits
     beside oracle text the player is trying to read. */
  .not-implemented {
    margin-top: 4px;
    font-size: 10.5px;
    line-height: 1.35;
    letter-spacing: 0.02em;
    color: var(--fg-dim);
  }
  .info-foot {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    align-items: center;
    margin-top: 4px;
  }
  .state {
    font-family: var(--font-mono);
    font-size: 9.5px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    padding: 2px 7px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    color: var(--fg-muted);
  }
  .state-attack {
    color: var(--danger);
    border-color: rgba(255, 107, 107, 0.5);
  }
  .state-cmd {
    color: var(--gold-strong);
    border-color: rgba(217, 180, 92, 0.5);
  }
  /* #781: brighter than the ambient state chips beside it. This one
     is not a passing condition — it is part of reading the card. */
  .state-chosen {
    color: var(--fg);
    border-color: var(--border-strong);
    text-transform: none;
    letter-spacing: 0.04em;
  }
  /* Commander-damage section — lives at the bottom of the info
     panel when the hovered card is a commander. Replaces the
     free-floating CommanderDamageTooltip so the information is
     anchored to the card the user is actually inspecting. */
  .cmdr-dmg {
    margin-top: 6px;
    padding-top: 8px;
    border-top: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .cmdr-label {
    font-family: var(--font-mono);
    font-size: 9.5px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--fg-dim);
    font-weight: 700;
  }
  .cmdr-rows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }
  .cmdr-row {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .cmdr-row.zero {
    opacity: 0.5;
  }
  .cmdr-name {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    flex: 0 0 auto;
    width: 30%;
  }
  .cmdr-name-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--seat-color, #888);
    flex: 0 0 auto;
  }
  .cmdr-name-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .bar-track {
    flex: 1;
    height: 4px;
    border-radius: 999px;
    background: var(--surface-hover);
    position: relative;
  }
  .bar-track i {
    position: absolute;
    left: 0;
    top: 0;
    height: 4px;
    border-radius: 999px;
  }
  .cmdr-amount {
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
    font-weight: 700;
    color: var(--fg);
    min-width: 2ch;
    text-align: right;
  }
  .cmdr-amount.lethal {
    color: var(--danger);
  }
  .cmdr-empty {
    font-size: 10.5px;
    color: var(--fg-dim);
    padding: 2px 0 0;
  }
</style>

<script lang="ts">
  import ManaCost from "./ManaCost.svelte";
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
  import {
    isFlippedFlipCard,
    isOneSidedLayout,
    isSidewaysArt,
    LAYOUT_FLIP,
    otherSideHeading,
    otherSideIndex,
    printedFace,
    type PrintedFace,
  } from "../../faces";
  import { L } from "../../labels";
  import { cardArt } from "../../cardArt";
  import { isToken } from "../../tokenGroups";
  import { stickerText } from "../../stickers";
  import { metaFor, type CardMeta } from "../../cardMetaCache";
  import type { GameView, PlayerView } from "../../protocol";
  import { playerColor } from "../../avatarColor";
  import { seatColor } from "../../colors";
  import { noUntapFooterLines } from "../../noUntap";
  import { chosenValueChips } from "../../chosenValues";
  import { damageBadge } from "../../damageBadge";

  interface Props {
    view: GameView;
  }
  const { view }: Props = $props();

  const card = $derived($hoveredCard);
  const imgSrc = $derived(cardImageURL(card, "normal"));
  // ADR 0151 (owner decision 2): a flipped flip card's one image is
  // shown turned 180 degrees, as on the board, so the half that is up
  // reads the right way up.
  const flippedArt = $derived(isFlippedFlipCard(card));
  // #3518: a split card's one image has its text sideways; the scan is
  // landscape and the image turned a quarter-turn so it reads.
  const sideways = $derived(isSidewaysArt(card));
  // The OTHER face of a multi-face card, if there is one — the back
  // of a modal DFC being previewed from hand, or the front of one
  // already flipped onto the battlefield. Null for the ~33,000
  // single-faced oracle IDs, which is what hides the panel.
  const otherFace = $derived.by(() => {
    const faces = card?.faces;
    if (!faces || faces.length < 2) return null;
    // A split, adventurer or prepare card's halves share one image:
    // nothing to inset, and no ?face=1 scan to request (#3518, #4109).
    if (isOneSidedLayout(card)) return null;
    const active = card?.active_face ?? 0;
    const idx = active === 0 ? 1 : 0;
    // A flip card's other half is the same image turned the other way
    // (ADR 0151 section 7): upside down beside its top half, upright
    // beside a flipped permanent.
    const turned = card?.layout === LAYOUT_FLIP && !flippedArt;
    return { index: idx, face: faces[idx], src: cardImageURL(card, "normal", idx), turned };
  });

  // ADR 0145 — meld. A melded permanent shows the two cards it is
  // made of; a meld card shows the permanent it melds into. Both are
  // small insets in the same corner the other face uses, because they
  // answer the same question: what else is this card?
  const meldInsets = $derived.by(() => {
    const parts = card?.melded_from;
    if (parts && parts.length > 0) {
      return parts
        .filter((p) => p.image)
        .map((p) => ({ name: p.name, title: `Melded from ${p.name}`, src: p.image as string }));
    }
    const into = card?.melds_into;
    if (into?.image) {
      return [{ name: into.name, title: `Melds into ${into.name}`, src: into.image }];
    }
    return [];
  });
  const meldedFromNames = $derived(card?.melded_from?.map((p) => p.name).join(" + ") ?? "");

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
  // #2257 — the damage marked, and on an indestructible creature why
  // it doesn't die. The badge on the card says it in a tooltip; the
  // panel is where a player looks when the board surprises them.
  const damage = $derived(card ? damageBadge(card) : null);
  // Bars are tinted with the commander's controller colour (the seat
  // dealing the damage) and flip to danger at lethal.
  const controllerColor = $derived(
    seatColor(view.seats.find((s) => s.id === card?.controller)?.seat ?? 0),
  );

  // Live P/T comes straight off the wire: since S16 the server sends
  // CurrentPower()/CurrentToughness() — effective P/T with every P/T
  // counter kind's delta already baked in (#1664) — so re-adding counters
  // here would double-count them. The printed parenthetical reads the
  // Scryfall metadata (string-typed; handles "*" stats) and only
  // renders once the meta fetch lands and the values actually differ.
  const livePT = $derived.by(() => {
    if (!card || card.power == null || card.toughness == null) return null;
    const live = `${card.power}/${card.toughness}`;
    const printed = upFace
      ? upFace.statsTitle === "printed power/toughness"
        ? (upFace.stats ?? null)
        : null
      : meta?.power != null && meta?.toughness != null
        ? `${meta.power}/${meta.toughness}`
        : null;
    return printed !== null && printed !== live ? `${live} (${printed})` : live;
  });

  // #4106 — the two sides of a two-faced card, as text. Scryfall leaves
  // the top level of a double-faced or flip card with the joined name
  // ("A // B") and no rules text at all, so for these cards the panel
  // reads each face on its own: the face that is up in the usual place,
  // and the other one in a section below it. That section is what makes
  // the back readable without a key, on touch and to a screen reader.
  // A split or adventurer card is one side and stays as it was (#3518);
  // a face-down card shows nothing more (otherSideIndex).
  const otherIdx = $derived(otherSideIndex(card));
  const upFace = $derived.by((): PrintedFace | null => {
    if (otherIdx === null || !card?.faces) return null;
    const i = card.active_face ?? 0;
    const face = card.faces[i];
    return face ? printedFace(face, meta?.card_faces?.[i]) : null;
  });
  // A meld card's other side is the permanent it melds into (ADR 0145):
  // its back is half of that card's face (CR 712.1), so the section
  // names it by what it becomes.
  const meldsInto = $derived(card && !card.face_down ? card.melds_into : undefined);
  const otherSide = $derived.by((): { heading: string; face: PrintedFace } | null => {
    if (!card) return null;
    if (otherIdx !== null && card.faces?.[otherIdx]) {
      return {
        heading: otherSideHeading(card),
        face: printedFace(card.faces[otherIdx], meta?.card_faces?.[otherIdx]),
      };
    }
    if (meldsInto?.name) return { heading: "Melds into", face: printedFace(meldsInto) };
    return null;
  });

  // The name, type line and cost of the face that is up: the face's own
  // for a two-faced card, the card's for everything else.
  const upName = $derived(upFace?.name ?? meta?.name ?? card?.name ?? "");
  const upType = $derived(upFace ? upFace.type_line : meta?.type_line);
  const upCost = $derived(upFace ? upFace.mana_cost : meta?.mana_cost);
  const upText = $derived(upFace ? upFace.oracle_text : meta?.oracle_text);
  const upLoyalty = $derived(
    upFace
      ? upFace.statsTitle === "starting loyalty"
        ? upFace.stats
        : undefined
      : meta?.loyalty
        ? `◈ ${meta.loyalty}`
        : undefined,
  );

  // #4106 — hold Shift to see the other side big. While Shift is down
  // the big image shows the face that is NOT up and the inset the one
  // that is; letting go swaps them back. Only the images swap: the text
  // above already lists both sides. A key typed into a text field (chat,
  // a deck name, an X value) is not a request to flip anything, and the
  // flag drops whenever the window loses the key's release: blur, a tab
  // switch, or a pointer event that reports Shift up.
  let shiftHeld = $state(false);
  const NON_TEXT_INPUTS = new Set([
    "button",
    "checkbox",
    "color",
    "file",
    "image",
    "radio",
    "range",
    "reset",
    "submit",
  ]);
  function typingIn(target: EventTarget | null): boolean {
    if (!(target instanceof HTMLElement)) return false;
    if (target instanceof HTMLTextAreaElement || target instanceof HTMLSelectElement) return true;
    if (target instanceof HTMLInputElement) return !NON_TEXT_INPUTS.has(target.type);
    return (
      target.isContentEditable ||
      target.closest('[contenteditable]:not([contenteditable="false"])') !== null
    );
  }
  function onKeyDown(e: KeyboardEvent): void {
    if (e.key !== "Shift") {
      if (!e.shiftKey) shiftHeld = false;
      return;
    }
    if (e.repeat || typingIn(e.target)) return;
    shiftHeld = true;
  }
  function onKeyUp(e: KeyboardEvent): void {
    if (e.key === "Shift" || !e.shiftKey) shiftHeld = false;
  }
  function releaseShift(): void {
    shiftHeld = false;
  }
  function onPointerMove(e: PointerEvent): void {
    if (shiftHeld && !e.shiftKey) shiftHeld = false;
  }

  // What the flip swaps in: the other face of a two-faced card (a flip
  // card's is the same image turned the other way), or the permanent a
  // meld card melds into. Null for every card with no other side, so
  // Shift does nothing to them.
  const flipSide = $derived.by((): { src: string; name: string; turned: boolean } | null => {
    if (otherFace?.src && otherSide && otherIdx !== null) {
      return { src: otherFace.src, name: otherSide.face.name, turned: otherFace.turned };
    }
    if (meldsInto?.image) return { src: meldsInto.image, name: meldsInto.name, turned: false };
    return null;
  });
  const flipped = $derived(shiftHeld && flipSide !== null && imgSrc !== null);
  const bigSrc = $derived(flipped && flipSide ? flipSide.src : imgSrc);
  const bigAlt = $derived(flipped && flipSide ? flipSide.name : upName);
  const bigTurned = $derived(flipped && flipSide ? flipSide.turned : flippedArt);
  // The face that is up, for the inset while the flip is held.
  const upInset = $derived(
    flipped && imgSrc ? { src: imgSrc, name: upName, turned: flippedArt } : null,
  );
</script>

<svelte:window
  onkeydowncapture={onKeyDown}
  onkeyupcapture={onKeyUp}
  onblur={releaseShift}
  onpointermovecapture={onPointerMove}
/>
<svelte:document onvisibilitychange={releaseShift} />

{#if card}
  <!-- #4106: no longer aria-hidden. The panel is still non-interactive
       (pointer-events: none) and takes no focus, but its text is how a
       screen reader reaches a card's other face, so it stays in the
       accessibility tree and its images are named for the face each
       one shows. -->
  <div class="overlay" class:two-sided={otherSide !== null} data-board-shot-private>
    <div class="scan" class:sideways>
      {#if bigSrc}
        <!-- keyed so a new card never shows the previous card's scan
             while its own image is still loading. The failure pip
             is display-only here: the panel is pointer-events: none,
             and it closes as soon as the pointer leaves the card it
             previews. fetchpriority (#33): the panel exists to be read
             the moment it opens; the small other-face inset below does
             not need the hint. -->
        {#key bigSrc}
          <img
            src={bigSrc}
            alt={bigAlt}
            class:flipped={bigTurned}
            class:sideways
            fetchpriority="high"
            use:cardArt={{ url: bigSrc, interactive: false }}
          />
        {/key}
      {:else}
        <div class="name-fallback">{card.name}</div>
      {/if}
      <!-- ADR 0034: the other printed face. A modal DFC in hand is
           two playable objects, and the half you are NOT looking at
           is exactly the information the hover panel exists to
           supply. Absent for every single-faced card. While Shift is
           held (#4106) the big image and the inset trade places, so
           the inset is then the face that is up. -->
      {#if meldInsets.length > 0}
        <div class="meld-insets">
          {#if upInset && meldsInto}
            <div class="other-face meld-inset" title={upInset.name}>
              {#key upInset.src}
                <img
                  src={upInset.src}
                  alt={upInset.name}
                  use:cardArt={{ url: upInset.src, interactive: false }}
                />
              {/key}
            </div>
          {:else}
            {#each meldInsets as inset (inset.src)}
              <div class="other-face meld-inset" title={inset.title}>
                <img
                  src={inset.src}
                  alt={inset.name}
                  use:cardArt={{ url: inset.src, interactive: false }}
                />
              </div>
            {/each}
          {/if}
          {#if flipSide}<span class="flip-hint">{L.holdShiftToFlip}</span>{/if}
        </div>
      {/if}
      {#if otherFace?.src}
        {@const inset = upInset ?? {
          src: otherFace.src,
          name: otherFace.face.name,
          turned: otherFace.turned,
        }}
        <div class="inset-col">
          <div class="other-face" title={inset.name}>
            {#key inset.src}
              <img
                src={inset.src}
                alt={inset.name}
                class:flipped={inset.turned}
                use:cardArt={{ url: inset.src, interactive: false }}
              />
            {/key}
          </div>
          {#if flipSide}<span class="flip-hint">{L.holdShiftToFlip}</span>{/if}
        </div>
      {/if}
    </div>
    <div class="info">
      <header class="info-head">
        <span class="title">{upName}</span>
        {#if livePT}
          <span class="pt" title="live power/toughness (printed in parens)">{livePT}</span>
        {:else if upLoyalty}
          <span class="pt" title="starting loyalty">{upLoyalty}</span>
        {/if}
      </header>
      <div class="type-line">
        {#if upType}<span>{upType}</span>{/if}
        {#if upCost}<span class="cost"><ManaCost cost={upCost} size={13} /></span>{/if}
      </div>
      <!-- ADR 0160: an exchanged text box. The oracle text below is
           what the card prints, which is no longer what it does. -->
      {#if card.text_box_from}
        <div class="text-box-from">
          Has the text box of {card.text_box_from} — the printed text below doesn't apply.
        </div>
      {/if}
      <!-- ADR 0160's amendment: changed words. The oracle text below
           still prints the old ones. -->
      {#if card.text_changes?.length}
        <div class="text-box-from">
          Text changed: {card.text_changes.join(", ")}
        </div>
      {/if}
      {#if upText}
        <div class="oracle">{upText}</div>
      {:else if meta === null && !upFace}
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
      <!-- #4106: the other side of a two-faced card, always listed, so
           the back is readable without the Shift flip: on touch, and
           to a screen reader. Printed text only — what the face would
           be, not what the permanent is doing now. -->
      {#if otherSide}
        <section class="other-side">
          <h3 class="other-head">{otherSide.heading}</h3>
          <div class="info-head">
            <span class="title other-title">{otherSide.face.name}</span>
            {#if otherSide.face.stats}
              <span class="pt" title={otherSide.face.statsTitle}>{otherSide.face.stats}</span>
            {/if}
          </div>
          {#if otherSide.face.type_line || otherSide.face.mana_cost}
            <div class="type-line">
              {#if otherSide.face.type_line}<span>{otherSide.face.type_line}</span>{/if}
              {#if otherSide.face.mana_cost}<span class="cost"
                  ><ManaCost cost={otherSide.face.mana_cost} size={13} /></span
                >{/if}
            </div>
          {/if}
          {#if otherSide.face.oracle_text}
            <div class="oracle other-oracle">{otherSide.face.oracle_text}</div>
          {/if}
        </section>
      {/if}
      {#if isToken(card) || card.tapped || card.attacking_target || card.blocking_target || card.goaded_by || card.is_commander || counterChips.length > 0 || (card.stickers?.length ?? 0) > 0 || noUntapLines.length > 0 || chosen.length > 0 || damage || meldedFromNames}
        <footer class="info-foot">
          {#if isToken(card)}
            <!-- #3043: a token copy shows the original's art and oracle
                 text, so the panel has to say it is not the card. -->
            <span class="state state-token" title="token (CR 111)"
              >{card.token_copy ? "token copy" : "token"}</span
            >
          {/if}
          {#if card.is_commander}
            <span class="state state-cmd">commander</span>
          {/if}
          {#if meldedFromNames}
            <span class="state" title="one permanent, two cards (CR 712.4a)"
              >melded from {meldedFromNames}</span
            >
          {/if}
          <!-- #781: first in the footer, ahead of tapped / attacking.
               The others describe what is happening to the permanent
               right now; this one is what the rest of its printed text
               MEANS, and a player reading the oracle line above needs
               it to finish the sentence. -->
          {#each chosen as chip (chip.kind)}
            <span class="state state-chosen" title={chip.title}>
              {#if chip.kind === "color"}
                chosen color: {chip.label}
              {:else if chip.kind === "tribe"}
                chosen type: {chip.label}
              {:else}
                <!-- option/name labels already carry their own prefix
                     ("Mode: Temur", "Named: Sol Ring") — see
                     chosenValues.ts. -->
                {chip.label}
              {/if}
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
            <span class="state"
              >blocking{#if (card.blocking_targets?.length ?? 0) > 1}
                ×{card.blocking_targets?.length}{/if}</span
            >
          {/if}
          {#if card.goaded_by}
            <span class="state state-attack">goaded</span>
          {/if}
          {#each counterChips as chip (chip.name)}
            <span class="state">{chip.name} ×{chip.n}</span>
          {/each}
          <!-- ADR 0208: what is stuck on it (CR 123). -->
          {#each card.stickers ?? [] as sticker, i (i)}
            <span class="state" title={sticker.sheet ? `from ${sticker.sheet}` : undefined}
              >sticker: {stickerText(sticker)}</span
            >
          {/each}
          {#if damage}
            <span class="state state-damage" title={damage.title}>{damage.summary}</span>
          {/if}
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
  /* #4213: the zoom scales with the screen, height first. The art is
     240px up to a 800px-tall viewport (a laptop keeps the size it had),
     then 30vh, up to 380px on a 1440p monitor; the panel is the art plus
     its text margin. Global so Game.svelte can keep the panel on screen
     beside a wide dock sheet (--zoom-right). */
  :global(:root) {
    --zoom-art: clamp(240px, 30vh, 380px);
    --zoom-w: calc(var(--zoom-art) + 92px);
  }
  /* A tutorial step pointing at the portraits and the mana pool moves the
     preview to the left, off them (TutorialCoach sets the attribute). */
  :global(:root[data-tutorial-preview="left"]) .overlay {
    right: auto;
    left: 10px;
  }
  .overlay {
    position: absolute;
    /* Pin to the top-right corner of the board with a small inset so
       the panel always fits inside the viewport regardless of width. */
    /* ADR 0111 §4 (PR 6): left of an open dock sheet (Game.svelte sets
       --zoom-right while one is open). */
    right: var(--zoom-right, 10px);
    top: 10px;
    width: min(var(--zoom-w), calc(50% - 132px));
    /* ADR 0111 §4: a long card stops above the action dock, which
       shares this right edge (Game.svelte sets --dock-zoom-clear; 0
       without a dock and on a phone, where the dock is under the board). */
    max-height: calc(100% - 20px - var(--dock-zoom-clear, 0px));
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
    /* Reading text grows a little with the art: 12px on a laptop,
       14px at 1440 tall. The sizes below are em of this. */
    font-size: clamp(12px, calc(6px + 0.56vh), 14px);
  }
  .scan {
    width: min(var(--zoom-art), 100%);
    aspect-ratio: 63 / 88;
    align-self: center;
    border-radius: 12px;
    overflow: hidden;
    background: var(--surface-sunken);
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.6);
    flex: 0 0 auto;
    position: relative;
  }

  /* The other printed face, tucked into the TOP-right corner of the
     scan as a small inset — present enough to read, small enough not
     to compete with the face that is actually up. Top, not bottom
     (#1965): the bottom of a card is its rules-text box, which is the
     thing the panel is opened to read, whereas the top corner only
     covers a title bar and mana cost the info panel repeats below. */
  .scan.sideways {
    width: 100%;
    aspect-ratio: 88 / 63;
    container-type: size;
  }
  .scan img.sideways {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 100cqh;
    height: 100cqw;
    transform: translate(-50%, -50%) rotate(90deg);
  }

  .other-face {
    position: absolute;
    right: 6px;
    top: 6px;
    width: 30%;
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
  /* ADR 0151: a flip card's image turned to show its other half. */
  img.flipped {
    transform: rotate(180deg);
  }

  /* ADR 0145: a melded permanent's two cards (or a meld card's combined
     back face), stacked down the same top-right corner as the other
     face — the column replaces the single inset's absolute placement. */
  .meld-insets {
    position: absolute;
    right: 6px;
    top: 6px;
    width: 30%;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .meld-inset {
    position: static;
    width: 100%;
  }

  /* #4106: the other face's inset with its hold-Shift hint under it,
     in the same corner the inset always had. */
  .inset-col {
    position: absolute;
    right: 6px;
    top: 6px;
    width: 30%;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 4px;
  }
  .inset-col .other-face {
    position: static;
    width: 100%;
  }
  /* A key only a keyboard has, so the hint is for pointer devices only;
     on touch the info panel's other-face section is the way to read the
     back. Wider than the inset, so it hangs left from its right edge. */
  .flip-hint {
    display: none;
    align-self: flex-end;
    white-space: nowrap;
    font-family: var(--font-mono);
    font-size: 9.5px;
    font-weight: 600;
    letter-spacing: 0.04em;
    line-height: 1.4;
    padding: 1px 7px;
    border-radius: 999px;
    color: var(--fg);
    background: color-mix(in srgb, var(--surface) 90%, transparent);
    border: 1px solid var(--border-strong);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
  }
  @media (hover: hover) {
    .flip-hint {
      display: block;
    }
  }

  /* #4106: the other side's text. The same shape as the face that is
     up, one step quieter: a hairline above, a mono label, and the rules
     text in the muted ink. */
  .other-side {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-top: 6px;
    padding-top: 8px;
    border-top: 1px solid var(--border);
  }
  .other-head {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 9.5px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--fg-muted);
  }
  .other-title {
    font-size: 1.0833em;
  }
  .oracle.other-oracle {
    max-height: clamp(60px, 14vh, 180px);
    margin-top: 2px;
    color: var(--fg-muted);
  }
  /* Two faces of text below the scan: on a short screen the scan gives
     up width so both fit (200px at a laptop's 768px), and from about
     1000px tall it is the full --zoom-art again (#4213). When the text
     still doesn't fit, the other side's text is the one that shortens
     (it fades at the cut), never the face that is up. */
  .overlay.two-sided .scan:not(.sideways) {
    width: min(clamp(200px, calc(49vh - 180px), var(--zoom-art)), 100%);
  }
  .overlay.two-sided .info > .oracle {
    flex-shrink: 0;
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
    font-size: 1.1667em;
    letter-spacing: -0.01em;
    color: var(--fg);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .pt {
    font-family: var(--font-mono);
    font-size: 1em;
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
    font-size: 0.9583em;
    color: var(--fg-muted);
  }
  .cost {
    font-family: var(--font-mono);
    font-size: 0.9167em;
    color: var(--fg-dim);
    flex: 0 0 auto;
  }
  .oracle {
    /* Long cards (Cyclonic Rift, Counterspell) get a scroll once we
       hit ~7 lines so the overlay doesn't grow taller than the
       viewport. clamp height keeps the panel proportional. */
    max-height: clamp(80px, 18vh, 260px);
    overflow-y: auto;
    font-size: 0.9583em;
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
  /* ADR 0160: an exchanged text box, set apart like a grant. */
  .text-box-from {
    margin: 4px 0;
    padding: 6px 10px;
    border-radius: 8px;
    border: 1px dashed var(--border, var(--fg-dim));
    font-size: 0.9167em;
    line-height: 1.35;
    color: var(--fg);
  }
  /* ADR 0093: granted abilities, read like oracle text but set apart
     from it — the card did not print them. */
  .granted {
    list-style: none;
    margin: 4px 0 0;
    padding: 6px 10px;
    border-radius: 8px;
    border: 1px dashed var(--border, var(--fg-dim));
    font-size: 0.9167em;
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
    font-size: 0.875em;
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
  .state-attack,
  .state-damage {
    color: var(--danger);
    border-color: color-mix(in srgb, var(--danger) 50%, transparent);
  }
  /* #3043: matches the teal TOKEN badge on the board tile. */
  .state-token {
    color: #bff3ea;
    border-color: rgba(110, 220, 200, 0.6);
  }
  .state-cmd {
    color: var(--accent-strong);
    border-color: var(--accent-line);
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
    font-size: 1em;
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

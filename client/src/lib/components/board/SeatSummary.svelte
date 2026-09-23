<script lang="ts">
  // SeatSummary is an opponent's board rendered as a read-out rather
  // than as cards. It is what `settings.display.opponentDetail =
  // "summary"` (the default) mounts in place of PlayerPanel.
  //
  // Rows, top to bottom:
  //
  //   1. identity   avatar, name, life, active/priority marks
  //   2. mana       one pip per colour with a count, then a dashed
  //                 "?" pip for anything whose colour is unknown
  //   3. counts     untapped creatures / total / untapped power,
  //                 hand size, library and graveyard counts
  //   4. creatures  one pip per body: P/T plus combat keywords
  //   5. structural one tile per planeswalker, commander, or
  //                 permanent wearing an attachment
  //
  // WHY. See the header of seatSummary.ts: card size is a share of
  // panel height with a clamp() floor, so a small panel stops
  // shrinking and starts clipping (#956). A representation that
  // degrades by CHANGING rather than SCALING has no floor to hit.
  //
  // This component holds no derivation logic — every number comes from
  // buildSeatSummary, which is pure and unit-tested. Keep it that way:
  // a rule that decides what the player sees belongs in a test, not in
  // a template.
  //
  // THE ARIA CONTRACT IS LOAD-BEARING. tests-e2e/tests/board-layout.spec.ts
  // and game.spec.ts find a seat's board by `role="region"` +
  // aria-label `"<name> board"`. A summary is still that seat's board,
  // so it keeps both. The expand control is labelled separately so a
  // screen reader user is told the full board is reachable.

  import type { ActionPayload, ActionType, CardView, GameView, PlayerView } from "../../protocol";
  import { buildSeatSummary, manaLabel, MANA_ORDER } from "../../seatSummary";
  import { COLOR_META } from "../../manaPick";
  import { KEYWORD_ICONS } from "../../keywordIcons";
  import { openZoneBrowser } from "../../zoneBrowser";
  import PlayerIdentity from "./PlayerIdentity.svelte";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seat: PlayerView;
    view: GameView;
    viewerID: string | null;
    /** Battlefield slice already filtered to cards this seat controls. */
    controlledCards: CardView[];
    isActive: boolean;
    hasPriority: boolean;
    isMonarch: boolean;
    isInitiative: boolean;
    sendAction: ActionSender;
    combatMode: "idle" | "attack" | "block";
    selectedCombatCardID: string | null;
    onDeclareAttack: (targetPlayerID: string) => void;
    onDeclareBlock: (attackerCardID: string) => void;
    onTargetPlayer?: (targetPlayerID: string) => void;
    onTargetCard?: (card: CardView) => boolean;
    /** Pin this seat open. Board owns the pin; the summary just asks. */
    onExpand?: () => void;
    // #1307: threaded straight through to PlayerIdentity — see its
    // prop doc.
    considering?: boolean;
  }

  const {
    seat,
    view,
    viewerID,
    controlledCards,
    isActive,
    hasPriority,
    isMonarch,
    isInitiative,
    sendAction,
    combatMode,
    selectedCombatCardID,
    onDeclareAttack,
    onDeclareBlock,
    onTargetPlayer,
    onTargetCard,
    onExpand,
    considering = false,
  }: Props = $props();

  const summary = $derived(buildSeatSummary(seat, controlledCards, view.battlefield?.cards ?? []));

  // The commander's art crop, for the avatar fallback when the seat
  // has no Discord portrait. Same derivation PlayerPanel uses; kept
  // here rather than lifted because the two components are mounted
  // exclusively of each other.
  const commanderScryfallID = $derived.by((): string | null => {
    const inCommand = seat.command?.cards?.find((c) => c.scryfall_id);
    if (inCommand?.scryfall_id) return inCommand.scryfall_id;
    for (const z of [view.battlefield, view.stack, view.exile, seat.graveyard]) {
      const hit = z?.cards?.find(
        (c) => c.is_commander && (c.owner === seat.id || c.controller === seat.id) && c.scryfall_id,
      );
      if (hit?.scryfall_id) return hit.scryfall_id;
    }
    return null;
  });

  // The seat's slice of the shared exile zone. Derived here rather
  // than passed in so Board needs no new prop — the same reason
  // commanderScryfallID is derived from `view` above.
  const exileCount = $derived((view.exile?.cards ?? []).filter((c) => c.owner === seat.id).length);

  const attackTargetable = $derived(
    !seat.eliminated && combatMode === "attack" && !!selectedCombatCardID,
  );

  // The colours with something behind them, in WUBRG order. An empty
  // colour is left out rather than shown as a zero: five greyed pips
  // read as "five colours" at a glance, which is the opposite of true.
  const manaPips = $derived(
    MANA_ORDER.filter((c) => (summary.mana.byColor[c] ?? 0) > 0).map((c) => ({
      color: c,
      count: summary.mana.byColor[c],
      meta: COLOR_META[c] ?? { label: c, fill: "#ccc" },
    })),
  );

  // Keywords worth a pip. The full badge row lives on the card in an
  // expanded panel; at summary size only the ones that change whether
  // a creature can be attacked into are worth the pixels.
  const COMBAT_KEYWORDS = [
    "flying",
    "reach",
    "deathtouch",
    "first strike",
    "double strike",
    "menace",
    "trample",
    "vigilance",
    "defender",
    "indestructible",
  ];

  function pipKeywords(c: CardView): string[] {
    const on = c.abilities ?? [];
    return COMBAT_KEYWORDS.filter((k) => on.includes(k));
  }

  function pipLabel(c: CardView): string {
    const pt = `${c.power ?? 0}/${c.toughness ?? 0}`;
    const kw = pipKeywords(c);
    const bits = [c.known_by_you === false ? "face-down creature" : c.name || "creature", pt];
    if (c.tapped) bits.push("tapped");
    if (c.summoning_sick) bits.push("summoning sick");
    if (c.attacking_target) bits.push("attacking");
    if (kw.length > 0) bits.push(kw.join(", "));
    return bits.join(" — ");
  }

  // Graveyard and exile are PUBLIC information — PileBar has opened
  // them to every viewer since S18.5 — so the summary keeps them
  // openable rather than reducing them to a number. A count you
  // cannot click is a worse answer to "what did they bin" than a pile
  // you can, and it would make the summary lossy in a way the rest of
  // this panel is careful not to be.
  //
  // Library is deliberately NOT a button. PileBar keeps it owner-only
  // (it is a draw affordance, not a browser), and this panel only ever
  // draws an opponent.
  function openGraveyard(): void {
    openZoneBrowser({ zoneKind: "graveyard", ownerID: seat.id, ownerName: seat.name });
  }

  function openExile(): void {
    openZoneBrowser({ zoneKind: "exile", ownerID: seat.id, ownerName: seat.name });
  }

  // PileButton's exact aria-label. Not a coincidence and not worth
  // "improving": tests-e2e/tests/zone-browser.spec.ts finds an
  // opponent's graveyard by `/^grave: /`, and two spellings of one
  // affordance is how a contract rots.
  function pileLabel(label: string, n: number): string {
    return `${label}: ${n} card${n === 1 ? "" : "s"}`;
  }

  // Click routing mirrors PlayerPanel's: a live targeting prompt wins,
  // then block-mode on an incoming attacker. A pip is NOT a card tile,
  // so there is no tap-toggle and no card menu — anything beyond these
  // two intents expands the panel instead, which is the whole point of
  // the expansion model. Board decides whether the expansion happens;
  // this only asks.
  function handlePipClick(c: CardView): void {
    if (onTargetCard?.(c)) return;
    if (combatMode === "block" && c.attacking_target === viewerID) {
      onDeclareBlock(c.instance_id);
      return;
    }
    onExpand?.();
  }
</script>

<div
  class="summary"
  class:eliminated={summary.eliminated}
  class:active={isActive}
  role="region"
  aria-label={`${seat.name} board`}
>
  <div class="top">
    <PlayerIdentity
      {seat}
      {commanderScryfallID}
      isSelf={false}
      {isActive}
      {hasPriority}
      {attackTargetable}
      {isMonarch}
      {isInitiative}
      {sendAction}
      {onDeclareAttack}
      {onTargetPlayer}
      {considering}
    />
    <button
      class="expand"
      type="button"
      aria-label={`Show ${seat.name}'s full board`}
      onclick={() => onExpand?.()}
    >
      ⤢
    </button>
  </div>

  {#if summary.eliminated}
    <!-- An eliminated seat keeps its identity and life and reports
         nothing else. A read-out full of zeroes reads as a bug rather
         than as a player who is out. -->
    <p class="out">out</p>
  {:else}
    <!-- Mana first. It is the "can they respond?" read, and it is the
         line this whole design exists to surface: today a player gets
         it by counting untapped lands by eye and guessing at colours.

         It is a FLOOR, not a total — see manaAvailable's contract.
         The dashed "?" pip is what says so visually; manaLabel names
         the source count so the spoken form carries the same caveat. -->
    <div class="row mana" role="group" aria-label={manaLabel(summary.mana)}>
      {#each manaPips as pip (pip.color)}
        <span class="pip mana-pip" style:--pip-fill={pip.meta.fill} title={pip.meta.label}>
          {pip.count}
        </span>
      {/each}
      {#if summary.mana.flexible > 0}
        <span class="pip mana-pip flexible" title="Colour not determined — any-colour or hybrid">
          {summary.mana.flexible}?
        </span>
      {/if}
      {#if summary.mana.sources === 0}
        <span class="quiet">no mana</span>
      {/if}
    </div>

    <!-- Combat math: what can block, and how hard it hits. Untapped is
         the number you read on someone else's turn. -->
    <div class="row counts">
      <span class="stat">
        <strong>{summary.creatures.untapped}</strong> untapped
      </span>
      <span class="stat quiet">
        of {summary.creatures.total} · {summary.creatures.untappedPower} power
      </span>
      <span class="stat quiet">{summary.handCount} in hand</span>
      <span class="stat quiet" title="library">{summary.libraryCount} deck</span>
      <button
        class="pile"
        type="button"
        aria-label={pileLabel("grave", summary.graveyardCount)}
        title={`grave · ${summary.graveyardCount}`}
        onclick={openGraveyard}
      >
        gy {summary.graveyardCount}
      </button>
      <button
        class="pile"
        type="button"
        aria-label={pileLabel("exile", exileCount)}
        title={`exile · ${exileCount}`}
        onclick={openExile}
      >
        ex {exileCount}
      </button>
    </div>

    {#if summary.creatureCards.length > 0}
      <div class="row pips" role="group" aria-label={`${seat.name}'s creatures`}>
        {#each summary.creatureCards as c (c.instance_id)}
          <!-- Each pip stays a real, hit-testable element with the
               card's instance id on it: a pip is a legal target and a
               legal block, and CombatArrows resolves an endpoint by
               measuring the element that carries the card. If the
               arrow layer keys off a different attribute than this,
               match it here rather than adding a second contract. -->
          <button
            class="pip creature-pip"
            type="button"
            data-card-id={c.instance_id}
            class:tapped={c.tapped}
            class:sick={c.summoning_sick}
            class:attacking={!!c.attacking_target}
            title={pipLabel(c)}
            aria-label={pipLabel(c)}
            onclick={() => handlePipClick(c)}
          >
            <span class="pt">{c.power ?? 0}/{c.toughness ?? 0}</span>
            {#each pipKeywords(c) as k (k)}
              {@const icon = KEYWORD_ICONS[k]}
              {#if icon}
                <!-- Same table and the same suppression as KeywordBadgeRow:
                     KEYWORD_ICONS is literal SVG written in this repo, never
                     anything that came off the wire. The `{#if}` replaces the
                     old `?? ""` fallback, so an unmapped keyword renders
                     nothing at all rather than an empty span the pip still
                     pays gap for. -->
                <!-- eslint-disable-next-line svelte/no-at-html-tags -->
                <span class="kw" aria-hidden="true">{@html icon}</span>
              {/if}
            {/each}
          </button>
        {/each}
      </div>
    {/if}

    {#if summary.structural.length > 0}
      <div class="row structural">
        {#each summary.structural as s (s.card.instance_id)}
          <button
            class="pip struct-pip"
            type="button"
            data-card-id={s.card.instance_id}
            data-kind={s.kind}
            title={s.card.name}
            onclick={() => handlePipClick(s.card)}
          >
            <span class="name">{s.card.name}</span>
            {#if s.counter !== undefined}<span class="counter">{s.counter}</span>{/if}
          </button>
        {/each}
      </div>
    {/if}
  {/if}
</div>

<style>
  /* No container queries and no cqh-derived sizes: that is the point.
     Every size here is fixed, so the summary reads identically in a
     quadrant, in a third of a top row, and at any window height. */
  .summary {
    display: flex;
    flex-direction: column;
    gap: 6px;
    height: 100%;
    box-sizing: border-box;
    padding: 8px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    overflow: hidden;
  }
  .summary.active {
    border-color: rgba(217, 180, 92, 0.28);
  }
  .summary.eliminated {
    opacity: 0.55;
  }
  .top {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 8px;
  }
  .expand {
    flex: 0 0 auto;
    background: none;
    border: 1px solid var(--border);
    border-radius: 6px;
    color: var(--fg-dim);
    cursor: pointer;
    font-size: 13px;
    line-height: 1;
    padding: 4px 6px;
  }
  .expand:hover {
    color: var(--fg);
  }
  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px;
    min-width: 0;
  }
  .pip {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    border-radius: 5px;
    border: 1px solid var(--border);
    padding: 2px 5px;
    font-size: 11px;
    font-weight: 600;
    line-height: 1.3;
    background: rgba(255, 255, 255, 0.03);
    color: var(--fg);
  }
  .mana-pip {
    background: var(--pip-fill);
    color: #10141c;
    border-color: transparent;
    min-width: 18px;
    justify-content: center;
  }
  /* A colour this module could not pin down. Rendered as a question,
     not as a colour: see the under-report rule in seatSummary.ts. */
  .mana-pip.flexible {
    background: repeating-linear-gradient(45deg, var(--fg-dim) 0 3px, transparent 3px 6px);
    color: var(--fg);
    border: 1px dashed var(--border);
  }
  .creature-pip {
    cursor: pointer;
  }
  .creature-pip.tapped {
    opacity: 0.5;
  }
  .creature-pip.sick {
    border-style: dashed;
  }
  .creature-pip.attacking {
    border-color: #ff9a85;
    box-shadow: 0 0 0 1px rgba(255, 154, 133, 0.35);
  }
  .creature-pip .pt {
    font-variant-numeric: tabular-nums;
  }
  .kw {
    display: inline-flex;
    width: 11px;
    height: 11px;
    color: var(--fg-dim);
  }
  .kw :global(svg) {
    width: 100%;
    height: 100%;
  }
  .struct-pip {
    cursor: pointer;
    max-width: 100%;
  }
  .struct-pip .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 96px;
  }
  .struct-pip .counter {
    font-variant-numeric: tabular-nums;
    color: var(--fg-dim);
  }
  .stat strong {
    font-variant-numeric: tabular-nums;
  }
  /* The two openable piles. Styled as quiet text rather than as
     PileButton's tile, because at summary size a tile would compete
     with the creature pips for the eye — but it is a real button with
     PileButton's aria-label, so screen readers and the e2e suite see
     the same affordance the full panel offers. */
  .pile {
    background: none;
    border: none;
    padding: 0;
    color: var(--fg-dim);
    font: inherit;
    font-size: 11px;
    cursor: pointer;
    text-decoration: underline;
    text-decoration-style: dotted;
    text-underline-offset: 2px;
    font-variant-numeric: tabular-nums;
  }
  .pile:hover {
    color: var(--fg);
  }
  .quiet {
    color: var(--fg-dim);
    font-size: 11px;
  }
  .out {
    margin: 0;
    color: var(--fg-dim);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    font-weight: 600;
  }
</style>

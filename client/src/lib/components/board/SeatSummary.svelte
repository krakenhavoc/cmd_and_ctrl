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
  //
  // `avatarCard` (ADR 0077 §6 as amended, #4065 and #4102) is the
  // reflow's avatar card: a turn chip, the identity — name, avatar, life
  // — the expand control, and four read-out lines (in hand, mana,
  // creatures, piles), in one column of cards at the left end of the
  // opponents' row. Its
  // avatar peeks and pins the expanded overlay as every avatar does
  // (ADR 0120). It is still the seat's region, and it carries
  // `data-seat-stand-in` so an arrow to or from a card it is not drawing
  // starts or ends on it (CombatArrows, boardAnchor.findStandIn).

  import type { ActionPayload, ActionType, CardView, GameView, PlayerView } from "../../protocol";
  import { defendingPlayerOf } from "../../attackTargets";
  import { NO_LEGAL_ACTIONS, attackTargetOpen, type LegalActions } from "../../legalActions";
  import { buildSeatSummary, manaLabel, MANA_ORDER } from "../../seatSummary";
  import { COLOR_META } from "../../manaPick";
  import { KEYWORD_ICONS } from "../../keywordIcons";
  import { openZoneBrowser } from "../../zoneBrowser";
  import { heldPlayableCount } from "../../exileGrants";
  import PlayerIdentity from "./PlayerIdentity.svelte";
  import CommandStrip from "./CommandStrip.svelte";
  import CompanionSlot from "./CompanionSlot.svelte";
  import { seatColor } from "../../colors";
  import { L } from "../../labels";

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
    /**
     * Open this seat's board in the expanded overlay, pinned (ADR 0120
     * §6). Board owns the overlay; the summary just asks. `from` is the
     * control that asked, which gets focus back when the overlay closes.
     */
    onExpand?: (from: EventTarget | null) => void;
    // #1307: threaded straight through to PlayerIdentity — see its
    // prop doc.
    considering?: boolean;
    // ADR 0105 sub-PR 5: the frame's FULL legal-action lookup, for the
    // same defender gate PlayerPanel's identity uses (attackTargetOpen).
    // Absent means no information, and every opponent stays clickable.
    legalGate?: LegalActions;
    /**
     * Draw the reflow's avatar card (ADR 0077 §6 as amended, #4065, and
     * 2026-10-17, #4102): the identity, the expand control and a compact
     * read-out of hand, mana, creatures and piles.
     */
    avatarCard?: boolean;
    /**
     * The avatar card's turn chip, who plays right before this seat
     * ("Next up", "After you", "After <name>"; expansion.ts's
     * reflowAvatarOrder). Null draws none.
     */
    turnChip?: string | null;
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
    legalGate = NO_LEGAL_ACTIONS,
    avatarCard = false,
    turnChip = null,
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
  // #2559: how many of those this seat may play itself (Memory Vessel,
  // Rocco) — shown for another seat; the viewer's own are a button.
  const exileHeld = $derived(
    seat.id === viewerID
      ? 0
      : heldPlayableCount(
          (view.exile?.cards ?? []).filter((c) => c.owner === seat.id),
          seat.id,
        ),
  );

  // ADR 0105 sub-PR 5: as PlayerPanel's, read off the server's list.
  const attackTargetable = $derived(
    !seat.eliminated &&
      combatMode === "attack" &&
      !!selectedCombatCardID &&
      attackTargetOpen(
        legalGate,
        view.battlefield?.cards?.find((c) => c.instance_id === selectedCombatCardID),
        selectedCombatCardID,
        seat.id,
      ),
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
  // two intents opens the seat's board in the expanded overlay, pinned
  // (ADR 0120 §6), where its cards are cards. Board owns the overlay;
  // this only asks.
  function handlePipClick(c: CardView): void {
    if (onTargetCard?.(c)) return;
    if (combatMode === "block" && defendingPlayerOf(c) === viewerID) {
      onDeclareBlock(c.instance_id);
      return;
    }
    onExpand?.(null);
  }
</script>

{#snippet identity()}
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
{/snippet}

{#snippet expandControl()}
  <button
    class="expand"
    type="button"
    aria-label={`Show ${seat.name}'s full board`}
    onclick={(e) => onExpand?.(e.currentTarget)}
  >
    ⤢
  </button>
{/snippet}

<!-- The mana read and the two openable piles, shared by the summary and
     the avatar card so the two cannot drift apart. -->
{#snippet manaRead()}
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
{/snippet}

{#snippet graveButton()}
  <button
    class="pile"
    type="button"
    aria-label={pileLabel("grave", summary.graveyardCount)}
    title={`grave · ${summary.graveyardCount}`}
    onclick={openGraveyard}
  >
    gy {summary.graveyardCount}
  </button>
{/snippet}

{#snippet exileButton()}
  <button
    class="pile"
    type="button"
    aria-label={exileHeld > 0
      ? `${pileLabel("exile", exileCount)}, ${exileHeld} theirs to play`
      : pileLabel("exile", exileCount)}
    title={exileHeld > 0
      ? `exile · ${exileCount} · ${exileHeld} theirs to play`
      : `exile · ${exileCount}`}
    onclick={openExile}
  >
    ex {exileCount}{#if exileHeld > 0}<span class="held"> · {exileHeld} theirs</span>{/if}
  </button>
{/snippet}

{#if avatarCard}
  <!-- ADR 0077 as amended 2026-10-17 (#4102) and 2026-10-18 (#4118): one
       card in the reflow's column. Its turn chip and the expand control on
       top, the identity, then the read-out as four stat tiles (in hand,
       untapped creatures, their power, mana) over a footer of the piles.
       The card is a size container: a short one steps the avatar down to
       56px and lays the tiles in one row (the container queries below). -->
  <div
    class="summary seat-panel avatar-card"
    style:--seat-color={seatColor(seat.seat)}
    class:eliminated={summary.eliminated}
    class:active={isActive}
    class:seat-turn={isActive}
    role="region"
    aria-label={L.seatBoard(seat.name)}
    data-seat-stand-in={seat.id}
  >
    <div class="card-head">
      {#if turnChip && !summary.eliminated}
        <span class="turn-chip" class:next={turnChip === L.reflowNextUp} title={turnChip}
          >{turnChip}</span
        >
      {/if}
      {@render expandControl()}
    </div>
    <div class="card-identity">
      {@render identity()}
    </div>
    {#if summary.eliminated}
      <p class="out">out</p>
    {:else}
      <div class="readout" data-dice-avoid>
        <!-- Each tile is a number over its label. The spoken form is the
             sentence the label shortens ("4 in hand", "1 of 2 creatures
             untapped"), so a screen reader hears no bare "1/2". -->
        <div class="tiles">
          <p class="tile hand">
            <strong class="num" aria-hidden="true">{summary.handCount}</strong>
            <span class="lbl" aria-hidden="true">hand</span>
            <span class="sr-only">{summary.handCount} in hand</span>
          </p>
          <p class="tile creatures" title="untapped creatures of all its creatures">
            <strong class="num" aria-hidden="true"
              >{summary.creatures.untapped}<span class="of">/{summary.creatures.total}</span
              ></strong
            >
            <span class="lbl" aria-hidden="true">ready</span>
            <span class="sr-only"
              >{summary.creatures.untapped} of {summary.creatures.total} creatures untapped</span
            >
          </p>
          <p class="tile power" title="power of its untapped creatures">
            <strong class="num" aria-hidden="true">{summary.creatures.untappedPower}</strong>
            <span class="lbl" aria-hidden="true">power</span>
            <span class="sr-only">{summary.creatures.untappedPower} untapped power</span>
          </p>
          <div class="tile mana" role="group" aria-label={manaLabel(summary.mana)}>
            {#if summary.mana.sources === 0}
              <strong class="num none">0</strong>
            {:else}
              <span class="pips">{@render manaRead()}</span>
            {/if}
            <span class="lbl" aria-hidden="true">mana</span>
          </div>
        </div>
        <div class="zones">
          <span class="zone deck" title="library"
            ><span class="lbl">deck</span> <strong>{summary.libraryCount}</strong></span
          >
          <button
            class="zone pile-cell"
            type="button"
            aria-label={pileLabel("grave", summary.graveyardCount)}
            title={`grave · ${summary.graveyardCount}`}
            onclick={openGraveyard}
            ><span class="lbl">gy</span> <strong>{summary.graveyardCount}</strong></button
          >
          <button
            class="zone pile-cell"
            type="button"
            aria-label={exileHeld > 0
              ? `${pileLabel("exile", exileCount)}, ${exileHeld} theirs to play`
              : pileLabel("exile", exileCount)}
            title={exileHeld > 0
              ? `exile · ${exileCount} · ${exileHeld} theirs to play`
              : `exile · ${exileCount}`}
            onclick={openExile}
            ><span class="lbl">ex</span> <strong>{exileCount}</strong>{#if exileHeld > 0}<span
                class="held">+{exileHeld}</span
              >{/if}</button
          >
        </div>
      </div>
    {/if}
  </div>
{:else}
  <div
    class="summary seat-panel"
    style:--seat-color={seatColor(seat.seat)}
    class:eliminated={summary.eliminated}
    class:active={isActive}
    class:seat-turn={isActive}
    role="region"
    aria-label={L.seatBoard(seat.name)}
  >
    <div class="top">
      {@render identity()}
      <!-- #2349: the seat's commander, face up with its tax, as it sits
         beside the hand on a full board. -->
      <CommandStrip {seat} compact />
      <!-- ADR 0172: the seat's revealed companion, public, view only. -->
      <CompanionSlot {seat} compact />
      {@render expandControl()}
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
      <div class="row mana" data-dice-avoid role="group" aria-label={manaLabel(summary.mana)}>
        {@render manaRead()}
      </div>

      <!-- Combat math: what can block, and how hard it hits. Untapped is
         the number you read on someone else's turn. -->
      <div class="row counts" data-dice-avoid>
        <span class="stat">
          <strong>{summary.creatures.untapped}</strong> untapped
        </span>
        <span class="stat quiet">
          of {summary.creatures.total} · {summary.creatures.untappedPower} power
        </span>
        <span class="stat quiet">{summary.handCount} in hand</span>
        <span class="stat quiet" title="library">{summary.libraryCount} deck</span>
        {@render graveButton()}
        {@render exileButton()}
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
{/if}

<style>
  /* No container queries and no cqh-derived sizes in the summary: that
     is the point. Every size here is fixed, so the summary reads
     identically in a quadrant, in a third of a top row, and at any
     window height. The avatar card below is the one exception, and it
     still changes rather than scales: at a set height it swaps a size. */
  .summary {
    display: flex;
    flex-direction: column;
    gap: 6px;
    height: 100%;
    box-sizing: border-box;
    padding: 8px;
    border-radius: 14px;
    overflow: hidden;
  }
  .summary.eliminated {
    opacity: 0.55;
  }
  /* The reflow's avatar card (ADR 0077 as amended 2026-10-17, #4102, and
     2026-10-18, #4118): a size container in the 176px column
     (--reflow-avatar-w). Seat-tinted glass: the seat colour washes down
     from the top, edges the card and rings the avatar, so whose card it
     is reads before the name. Labels and counts are the mono voice; what
     a player reads to act is --fg-muted or --fg, never --fg-dim
     (DESIGN.md §2). */
  .avatar-card {
    container: avatar-card / size;
    gap: 6px;
    padding: 6px 8px 8px;
    background:
      radial-gradient(
        120% 55% at 50% 0%,
        color-mix(in srgb, var(--seat-color, var(--fg-dim)) 16%, transparent),
        transparent 75%
      ),
      linear-gradient(
        180deg,
        color-mix(in srgb, var(--surface-raised) 60%, var(--surface)),
        color-mix(in srgb, var(--surface) 75%, var(--bg))
      );
  }
  /* The turn's gold edge (.seat-turn) still wins over the seat's. */
  .avatar-card:not(.seat-turn) {
    border-color: color-mix(in srgb, var(--seat-color, var(--fg-dim)) 35%, var(--border));
  }
  .card-head {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 6px;
    min-height: 20px;
    flex: 0 0 auto;
  }
  .turn-chip {
    margin-right: auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding: 2px 8px;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--seat-color, var(--fg-dim)) 40%, transparent);
    background: color-mix(in srgb, var(--seat-color, var(--fg-dim)) 8%, transparent);
    color: var(--fg-muted);
    font-family: var(--font-mono);
    font-size: 9.5px;
    font-weight: 700;
    letter-spacing: 0.12em;
    line-height: 1.3;
    text-transform: uppercase;
  }
  /* The seat that plays next wears its seat colour, filled. */
  .turn-chip.next {
    color: #10141c;
    background: var(--seat-color, var(--fg));
    border-color: transparent;
  }
  .avatar-card .expand {
    display: grid;
    place-items: center;
    width: 22px;
    height: 22px;
    padding: 0;
    border: none;
    border-radius: 6px;
    color: var(--fg-muted);
    transition: background 120ms var(--ease);
  }
  .avatar-card .expand:hover {
    color: var(--fg);
    background: color-mix(in srgb, var(--overlay-ink) 8%, transparent);
  }
  .card-identity {
    --avatar-size-base: 84px;
    display: flex;
    justify-content: center;
    min-width: 0;
    flex: 0 0 auto;
    /* The identity and the read-out sit together, centred in the room
       under the head. */
    margin-top: auto;
  }
  /* The identity takes the card's width, so its name and tags (BOT,
     considering…, offline) share one wrapping line, as a seated identity
     does (ADR 0204), instead of a line each as in a rail. */
  .card-identity :global(.identity) {
    width: 100%;
  }
  .card-identity :global(.name-line) {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    align-items: center;
    gap: 2px 5px;
    max-width: 100%;
  }
  .card-identity :global(.name-line > .tag) {
    margin-top: 0;
    margin-bottom: 0;
    padding: 1px 5px;
    font-size: 8px;
    letter-spacing: 0.08em;
  }
  .card-identity :global(.core-row) {
    gap: 0;
  }
  /* A solid ring in the seat colour with a soft halo, on bots too: the
     BOT tag already says it is a bot, and the dashes read as noise. The
     colour and halo stand down for PlayerIdentity's state rings (turn,
     priority, a target), which keep their own. */
  .avatar-card .card-identity :global(.identity .avatar-wrap) {
    border-style: solid;
  }
  .card-identity
    :global(
      .identity:not(.active, .priority, .targetable, .cast-targetable, .cast-picked) .avatar-wrap
    ) {
    border-color: color-mix(in srgb, var(--seat-color, #888) 80%, var(--surface));
    box-shadow:
      0 0 0 4px color-mix(in srgb, var(--seat-color, #888) 14%, transparent),
      0 8px 20px rgba(0, 0, 0, 0.45);
  }
  .readout {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
    margin-bottom: auto;
    font-family: var(--font-mono);
    color: var(--fg-muted);
    font-variant-numeric: tabular-nums;
  }
  /* The tiles and the footer are segmented: cells on a 1px gap over a
     border-coloured ground, so the rules between them are the gap. */
  .tiles,
  .zones {
    display: grid;
    gap: 1px;
    border-radius: 9px;
    overflow: hidden;
    background: color-mix(in srgb, var(--seat-color, var(--fg-dim)) 18%, var(--border));
    border: 1px solid color-mix(in srgb, var(--seat-color, var(--fg-dim)) 18%, var(--border));
  }
  .tiles {
    grid-template-columns: 1fr 1fr;
  }
  .tile {
    margin: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1px;
    min-width: 0;
    padding: 6px 4px 5px;
    background: color-mix(in srgb, var(--surface) 82%, var(--bg));
  }
  .tile .num {
    color: var(--fg);
    font-size: 18px;
    font-weight: 700;
    line-height: 1.1;
    letter-spacing: -0.02em;
  }
  .tile .num .of {
    color: var(--fg-muted);
    font-size: 0.7em;
    font-weight: 600;
  }
  .tile .num.none {
    color: var(--fg-muted);
  }
  .lbl {
    font-size: 8.5px;
    font-weight: 600;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .tile .pips {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 2px;
    min-height: 20px;
    align-items: center;
  }
  .readout .mana-pip {
    font-family: var(--font-mono);
    font-size: 10.5px;
    padding: 1px 4px;
    min-width: 16px;
    border-radius: 999px;
  }
  .zones {
    grid-template-columns: 1.3fr 1fr 1fr;
  }
  .zone {
    display: flex;
    align-items: baseline;
    justify-content: center;
    gap: 4px;
    min-width: 0;
    padding: 4px 2px;
    background: color-mix(in srgb, var(--surface) 82%, var(--bg));
    font-size: 11px;
    white-space: nowrap;
  }
  .zone strong {
    color: var(--fg);
    font-weight: 700;
  }
  .pile-cell {
    border: none;
    border-radius: 0;
    font-family: inherit;
    color: inherit;
    cursor: pointer;
    transition: background 120ms var(--ease);
  }
  .pile-cell:hover,
  .pile-cell:focus-visible {
    background: color-mix(in srgb, var(--seat-color, var(--fg-dim)) 14%, var(--surface));
  }
  .pile-cell .held {
    color: var(--accent);
    font-size: 9.5px;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
    border: 0;
  }
  @container avatar-card (max-height: 239px) {
    .avatar-card {
      gap: 4px;
    }
    .card-identity {
      --avatar-size-base: 56px;
    }
    .card-identity
      :global(
        .identity:not(.active, .priority, .targetable, .cast-targetable, .cast-picked) .avatar-wrap
      ) {
      box-shadow:
        0 0 0 3px color-mix(in srgb, var(--seat-color, #888) 14%, transparent),
        0 6px 14px rgba(0, 0, 0, 0.45);
    }
    /* One row of four tiles instead of two rows of two: the same four
       facts, half the height. Mana gets the widest cell for its pips. */
    .tiles {
      grid-template-columns: repeat(3, minmax(0, 1fr)) minmax(0, 1.5fr);
    }
    .tile {
      padding: 3px 2px;
      gap: 0;
    }
    .tile .num {
      font-size: 14px;
    }
    .lbl {
      font-size: 7.5px;
      letter-spacing: 0.08em;
    }
    .tile .pips {
      min-height: 17px;
      gap: 1px;
    }
    .readout .mana-pip {
      font-size: 9.5px;
      padding: 0 3px;
      min-width: 13px;
    }
    .zone {
      padding: 2px;
      font-size: 10px;
    }
  }
  /* The shortest step (two cards on a 1024×768 board are about 170px):
     the same tiles, packed tighter. */
  @container avatar-card (max-height: 199px) {
    .avatar-card {
      gap: 2px;
      padding: 4px 6px 5px;
    }
    .card-head {
      min-height: 16px;
    }
    .turn-chip {
      padding: 1px 6px;
      font-size: 8.5px;
    }
    .avatar-card .expand {
      width: 18px;
      height: 18px;
      font-size: 12px;
    }
    .card-identity :global(.identity) {
      gap: 1px;
    }
    .readout {
      gap: 3px;
    }
    .tile .num {
      font-size: 13px;
    }
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
    background: color-mix(in srgb, var(--overlay-ink) 3%, transparent);
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
  /* #2559: the seat's own playable exile, beside the count. */
  .held {
    color: var(--fg-muted);
  }
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

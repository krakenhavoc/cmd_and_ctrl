<script lang="ts">
  // DeckStats — the body of one deck's advanced stats page (ADR 0217 §8):
  // the stat tiles, then Mana, Cards, Composition, Play and Commander.
  // #/stats/deck/<id> and a shared deck page draw the same component; a
  // shared one passes `shared`, which leaves out card images (they need a
  // session). The page around it (route, header, version picker, filters)
  // is routes/StatsDeck.svelte. Every chart is a ChartFigure.
  import ChartFigure from "./ChartFigure.svelte";
  import BarChart from "./BarChart.svelte";
  import CurvePanels from "./CurvePanels.svelte";
  import CardGrid from "./CardGrid.svelte";
  import StatTile from "./StatTile.svelte";
  import ManaTroubles from "./ManaTroubles.svelte";
  import PlacesFigure from "./PlacesFigure.svelte";
  import WinKindsFigure from "./WinKindsFigure.svelte";
  import AggressionFigure from "./AggressionFigure.svelte";
  import { avgPlaceText, recordText } from "../../myStats";
  import type { ChartBar } from "../../statsCharts";
  import {
    COLOUR_TEXT,
    commanderCaption,
    compositionCaption,
    curveText,
    dayText,
    landCurveCaption,
    medianLengthText,
    NOT_ENOUGH,
    oneDecimal,
    pctText,
    rateText,
    ROLE_TEXT,
    TYPE_TEXT,
    type AdvancedDeck,
  } from "../../advancedStats";

  interface Props {
    deck: AdvancedDeck;
    shared?: boolean;
  }
  const { deck: d, shared = false }: Props = $props();

  const IDENTITY: Record<string, string> = {
    W: "var(--identity-w)",
    U: "var(--identity-u)",
    B: "var(--identity-b)",
    R: "var(--identity-r)",
    G: "var(--identity-g)",
  };

  const facts = $derived(d.deck.facts);
  const played = $derived(d.record.games > 0);
  const others = $derived(d.other_decks);

  /** Counts as bars, each with a tick for the person's average over their other decks. */
  function compositionBars(
    counts: { key: string; cards: number }[],
    averages: Map<string, number>,
    names: Record<string, string>,
    tone: string,
  ): ChartBar[] {
    return counts
      .filter((c) => c.cards > 0 || (averages.get(c.key) ?? 0) > 0)
      .map((c) => {
        const avg = averages.get(c.key);
        return {
          label: names[c.key] ?? c.key,
          value: c.cards,
          text: avg == null ? `${c.cards}` : `${c.cards} (others ${oneDecimal(avg)})`,
          tone,
          tick: avg,
        };
      });
  }

  const typeBars = $derived<ChartBar[]>(
    facts
      ? compositionBars(
          facts.types.map((t) => ({ key: t.type, cards: t.cards })),
          new Map((others?.types ?? []).map((t) => [t.type, t.average])),
          TYPE_TEXT,
          "var(--chart-1)",
        )
      : [],
  );
  const roleBars = $derived<ChartBar[]>(
    facts?.roles_available && facts.roles
      ? compositionBars(
          facts.roles.map((r) => ({ key: r.role, cards: r.cards })),
          new Map((others?.roles ?? []).map((r) => [r.role, r.average])),
          ROLE_TEXT,
          "var(--chart-3)",
        )
      : [],
  );
  const hasRoles = $derived(!!facts?.roles_available);

  const colourBars = $derived<ChartBar[]>(
    (facts?.colour_share ?? [])
      .filter((c) => c.symbols > 0)
      .map((c) => ({
        label: COLOUR_TEXT[c.colour] ?? c.colour,
        mark: c.colour,
        value: c.share ?? 0,
        text: pctText(c.share),
        tone: IDENTITY[c.colour],
      })),
  );
  const colourCaption = $derived.by(() => {
    const top = [...colourBars].sort((a, b) => b.value - a.value)[0];
    return top
      ? `${top.label} is the largest share of this deck's mana symbols: ${pctText(top.value)}.`
      : "This deck has no coloured mana symbols.";
  });

  const commanderBars = $derived<ChartBar[]>([
    {
      label: "Cast by turn 5",
      value: d.commander.by_turn_5.rate ?? 0,
      text: d.commander.by_turn_5.rate == null ? NOT_ENOUGH : rateText(d.commander.by_turn_5),
      muted: d.commander.by_turn_5.rate == null,
      tone: "var(--chart-1)",
    },
    {
      label: "Cast later, or never",
      value: d.commander.later.rate ?? 0,
      text: d.commander.later.rate == null ? NOT_ENOUGH : rateText(d.commander.later),
      muted: d.commander.later.rate == null,
      tone: "var(--chart-6)",
    },
  ]);

  const partial = $derived(d.coverage.games - d.coverage.frames);
</script>

<div class="tiles" data-testid="deck-tiles">
  <StatTile label="Games" value={`${d.record.games}`} sub={recordText(d.record) + " (W-L)"} />
  <StatTile
    label="Win rate"
    value={d.win_rate.rate == null ? "-" : pctText(d.win_rate.rate)}
    sub={rateText(d.win_rate)}
  />
  <StatTile label="Average place" value={avgPlaceText(d.record.avg_place)} />
  <StatTile
    label="Average mana value"
    value={facts?.avg_mana_value == null ? "-" : facts.avg_mana_value.toFixed(2)}
    sub={facts ? curveText(facts) || undefined : undefined}
  />
  <StatTile label="Median length" value={medianLengthText(d.length)} />
</div>
{#if played && partial > 0 && d.coverage.games > 0}
  <p class="coverage">
    {partial} of {d.coverage.games} games have no replay to read, so they count in the record but not
    in Play and Mana.
  </p>
{/if}

<div class="sections">
  <section class="card wide" aria-labelledby="deck-sec-mana" data-testid="deck-mana">
    <h2 id="deck-sec-mana">Mana</h2>
    <div class="pair">
      <div class="col">
        {#if facts}
          <ChartFigure
            id="cap-deck-curves"
            title="Mana curve and land drops"
            caption={landCurveCaption(d.mana.land_curve)}
            columns={[
              "Mana value or land drop",
              "Non-land cards",
              "Land drop hit",
              "Lands controlled",
            ]}
            rows={facts.mana_curve.map((cards, i) => {
              const p = i === 0 ? undefined : d.mana.land_curve[i - 1];
              const name = i >= facts.mana_curve.length - 1 ? `${i}+` : `${i}`;
              return [
                name,
                cards,
                p?.hit.rate == null ? "-" : pctText(p.hit.rate),
                p?.controlled.rate == null ? "-" : pctText(p.controlled.rate),
              ];
            })}
            note="Both panels share one axis: a 5-drop sits above the land drop you need to cast it."
          >
            {#snippet children(id)}
              <CurvePanels labelledby={id} curve={facts.mana_curve} points={d.mana.land_curve} />
            {/snippet}
          </ChartFigure>
        {:else}
          <p class="empty">
            This deck's list can't be read, so there is no mana curve. Card data may not be loaded
            on this server.
          </p>
        {/if}
      </div>
      <div class="col">
        {#if played}
          <ManaTroubles mana={d.mana} />
        {:else}
          <p class="empty">Play 3 games with this deck to see how often you miss a land drop.</p>
        {/if}
      </div>
    </div>
  </section>

  <section class="card wide" aria-labelledby="deck-sec-cards" data-testid="deck-cards">
    <h2 id="deck-sec-cards">Cards</h2>
    <h3 class="sub">Dead cards</h3>
    {#if d.dead_cards.length === 0}
      <p class="empty">
        {played
          ? "No card sat in your hand uncast in this filter."
          : "Play 3 games with this deck to see which cards sit in your hand."}
      </p>
    {:else}
      <p class="cmeta">
        Spells that were in your hand in a game and never cast. “Could afford” means you could have
        paid for it on some turn in at least half of those games.
      </p>
      <CardGrid
        names={shared}
        cards={d.dead_cards.map((c) => ({
          name: c.name,
          scryfallId: c.scryfall_id,
          meta: `Dead in ${c.dead.n} of ${c.dead.of} ${c.dead.of === 1 ? "game" : "games"} in hand${c.dead.rate == null ? "" : ` (${pctText(c.dead.rate)})`}`,
          tag: c.could_afford ? "Could afford" : "Couldn't afford",
        }))}
      />
    {/if}
    <h3 class="sub">Most cast</h3>
    {#if d.most_cast.length === 0}
      <p class="empty">No spells cast in this filter.</p>
    {:else}
      <CardGrid
        names={shared}
        cards={d.most_cast.slice(0, 12).map((c) => ({
          name: c.name,
          scryfallId: c.scryfall_id,
          meta: `${c.casts} ${c.casts === 1 ? "cast" : "casts"} in ${c.games} ${c.games === 1 ? "game" : "games"}`,
        }))}
      />
    {/if}
  </section>

  <section class="card" aria-labelledby="deck-sec-comp" data-testid="deck-composition">
    <h2 id="deck-sec-comp">Composition</h2>
    {#if !facts}
      <p class="empty">This deck's list can't be read, so there is nothing to count.</p>
    {:else}
      <ChartFigure
        id="cap-deck-types"
        title="Card types"
        caption={compositionCaption(typeBars)}
        columns={["Type", "In this deck", "Your other decks, on average"]}
        rows={typeBars.map((b) => [b.label, b.value, b.tick == null ? "-" : oneDecimal(b.tick)])}
        note={others
          ? `The tick marks the average over your ${others.decks} other saved ${others.decks === 1 ? "deck" : "decks"}.`
          : "Save another deck to compare this one against your average."}
      >
        {#snippet children(id)}
          <BarChart labelledby={id} bars={typeBars} />
        {/snippet}
      </ChartFigure>

      <ChartFigure
        id="cap-deck-roles"
        title="Card roles"
        caption={hasRoles
          ? compositionCaption(roleBars)
          : "Card roles aren't available on this server right now, so only card types are counted."}
        columns={["Role", "In this deck", "Your other decks, on average"]}
        rows={roleBars.map((b) => [b.label, b.value, b.tick == null ? "-" : oneDecimal(b.tick)])}
      >
        {#snippet children(id)}
          {#if roleBars.length > 0}
            <BarChart labelledby={id} bars={roleBars} />
          {:else}
            <p class="empty">{hasRoles ? "No card has a role." : "No roles to show."}</p>
          {/if}
        {/snippet}
      </ChartFigure>
      {#if hasRoles}
        <p class="cmeta credit">
          Card roles from
          <a href="https://tagger.scryfall.com" target="_blank" rel="noopener noreferrer"
            >Scryfall Tagger</a
          >{#if d.deck.roles_at}, updated {dayText(d.deck.roles_at)}{/if}. A card counts under every
          role it has.
        </p>
      {/if}

      <ChartFigure
        id="cap-deck-colours"
        title="Colour of mana symbols"
        caption={colourCaption}
        columns={["Colour", "Symbols", "Share"]}
        rows={(facts.colour_share ?? [])
          .filter((c) => c.symbols > 0)
          .map((c) => [COLOUR_TEXT[c.colour] ?? c.colour, c.symbols, pctText(c.share)])}
        note="Counted from the non-land cards' costs; a hybrid symbol counts half to each colour."
      >
        {#snippet children(id)}
          {#if colourBars.length > 0}
            <BarChart labelledby={id} bars={colourBars} max={1} />
          {:else}
            <p class="empty">No coloured symbols.</p>
          {/if}
        {/snippet}
      </ChartFigure>
      {#if facts.unresolved > 0}
        <p class="cmeta">
          {facts.unresolved}
          {facts.unresolved === 1 ? "card" : "cards"} in the list {facts.unresolved === 1
            ? "isn't"
            : "aren't"} in our card data, so {facts.unresolved === 1 ? "it isn't" : "they aren't"} counted
          here.
        </p>
      {/if}
    {/if}
  </section>

  <section class="card" aria-labelledby="deck-sec-play" data-testid="deck-play">
    <h2 id="deck-sec-play">Play</h2>
    {#if played}
      <WinKindsFigure winKinds={d.win_kinds} comboRules={d.combo_rules} />
      <PlacesFigure places={d.places} />
      <AggressionFigure aggression={d.aggression} />
    {:else}
      <p class="empty">Play 3 games with this deck to see how it wins and how it attacks.</p>
    {/if}
  </section>

  <section class="card wide" aria-labelledby="deck-sec-cmd" data-testid="deck-commander">
    <h2 id="deck-sec-cmd">Commander</h2>
    {#if played}
      <div class="pair">
        <div class="cmd-tiles">
          <StatTile
            label="Casts a game"
            value={oneDecimal(d.commander.casts_per_game)}
            testid="deck-cmd-casts"
          />
          <StatTile
            label="First cast"
            value={d.commander.first_cast_turn.median == null
              ? "-"
              : `turn ${Math.round(d.commander.first_cast_turn.median)}`}
            sub="median game turn"
          />
        </div>
        <div class="col">
          <ChartFigure
            id="cap-deck-commander"
            title="Win rate by when the commander came down"
            caption={commanderCaption(d.commander)}
            columns={["Commander", "Win rate", "Wins", "Games"]}
            rows={[
              [
                "Cast by turn 5",
                d.commander.by_turn_5.rate == null ? "-" : pctText(d.commander.by_turn_5.rate),
                d.commander.by_turn_5.n,
                d.commander.by_turn_5.of,
              ],
              [
                "Cast later, or never",
                d.commander.later.rate == null ? "-" : pctText(d.commander.later.rate),
                d.commander.later.n,
                d.commander.later.of,
              ],
            ]}
          >
            {#snippet children(id)}
              <BarChart labelledby={id} bars={commanderBars} max={1} />
            {/snippet}
          </ChartFigure>
        </div>
      </div>
    {:else}
      <p class="empty">Play 3 games with this deck to see how often you cast your commander.</p>
    {/if}
  </section>
</div>

<style>
  .tiles {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
    gap: 12px;
  }
  .coverage {
    margin: 10px 0 0;
    color: var(--fg-muted);
    font-size: 0.85rem;
  }
  .sections {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 16px;
    margin-top: 16px;
    align-items: start;
  }
  .sections > .card {
    display: grid;
    gap: 22px;
    min-width: 0;
  }
  .sections > .wide {
    grid-column: 1 / -1;
  }
  h2 {
    margin: 0;
  }
  .sub {
    margin: 0;
    font: 600 0.72rem/1.2 var(--font-mono);
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .empty {
    margin: 0;
    color: var(--fg-muted);
    font-size: 0.9rem;
  }
  .cmeta {
    color: var(--fg-muted);
    font-size: 0.8rem;
    margin: 0;
  }
  .credit a {
    color: inherit;
  }
  .cmd-tiles {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: 10px;
    align-content: start;
  }
  /* Two columns inside a full-width section, one below the breakpoint. */
  .pair {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 22px 28px;
    align-items: start;
  }
  .col {
    display: grid;
    gap: 22px;
    min-width: 0;
  }
  @media (max-width: 819px) {
    .sections,
    .pair {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>

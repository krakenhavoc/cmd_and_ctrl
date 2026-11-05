<script lang="ts">
  // AdvancedOverview — the body of the advanced stats overview (ADR 0217
  // §8): the stat tiles, then Results, Play, Mana, Cards, Colours and
  // Decks. #/stats and a shared view (#/stats/shared/<token>) draw the same
  // component; a shared view passes `shared`, which leaves out what only
  // the owner can open. Every chart is a ChartFigure: a one-sentence
  // caption, the chart named by it, and the same numbers as a table.
  import ChartFigure from "./ChartFigure.svelte";
  import BarChart from "./BarChart.svelte";
  import StackedBar from "./StackedBar.svelte";
  import LineChart from "./LineChart.svelte";
  import Histogram from "./Histogram.svelte";
  import Sparkline from "./Sparkline.svelte";
  import StatTile from "./StatTile.svelte";
  import TrendStrip from "./TrendStrip.svelte";
  import { scryfallImageURL } from "../../cardImage";
  import { avgPlaceText, placeText, recordText, winRateText } from "../../myStats";
  import type { ChartBar } from "../../statsCharts";
  import { evennessWord } from "../../statsCharts";
  import {
    aggressionCaption,
    COLOUR_TEXT,
    COMBO_RULE_TEXT,
    coloursCaption,
    kindText,
    landNumberText,
    lengthCaption,
    medianLengthText,
    missedCaption,
    NOT_ENOUGH,
    oneDecimal,
    outKindsCaption,
    pctText,
    placesCaption,
    rateText,
    sortedKinds,
    spreadCaption,
    turnOrderCaption,
    winKindsCaption,
    type AdvancedOverview,
    type ManaGame,
  } from "../../advancedStats";

  interface Props {
    overview: AdvancedOverview;
    /** A shared view: no card images (they need a session) and no owner-only links. */
    shared?: boolean;
  }
  const { overview: o, shared = false }: Props = $props();

  const PLACE_TONES = [
    "var(--chart-1)",
    "var(--chart-2)",
    "var(--chart-3)",
    "var(--chart-4)",
    "var(--chart-5)",
    "var(--chart-6)",
  ];
  const IDENTITY: Record<string, string> = {
    W: "var(--identity-w)",
    U: "var(--identity-u)",
    B: "var(--identity-b)",
    R: "var(--identity-r)",
    G: "var(--identity-g)",
  };

  const winKinds = $derived(sortedKinds(o.win_kinds));
  const outKinds = $derived(sortedKinds(o.out_kinds));
  const comboRules = $derived(sortedKinds(o.combo_rules));
  const kindBars = (kinds: { kind: string; games: number }[], tone: string): ChartBar[] =>
    kinds.map((k) => ({ label: kindText(k.kind), value: k.games, text: `${k.games}`, tone }));

  const turnBars = $derived<ChartBar[]>(
    o.turn_order.map((r) => ({
      label: `Acting ${placeText(r.position)}`,
      value: r.wins.rate ?? 0,
      text:
        r.wins.rate == null
          ? `${r.wins.of} ${r.wins.of === 1 ? "game" : "games"}`
          : rateText(r.wins),
      muted: r.wins.rate == null,
      tone: "var(--chart-1)",
    })),
  );

  const lengthCats = $derived(o.length.by_turns.map((b) => `${b.value}`));
  const overTurnCats = $derived(
    o.aggression.over_turns.map((t) => (t.own_turn >= 12 ? "12+" : `${t.own_turn}`)),
  );

  const missed = $derived(o.mana.missed_by_land);
  const missedCats = $derived(missed.map((m) => (m.land >= 6 ? "6+" : `${m.land}`)));

  const colourBars = $derived<ChartBar[]>(
    o.colours.map((c) => ({
      label: COLOUR_TEXT[c.colour] ?? c.colour,
      mark: c.colour,
      value: c.games.rate ?? 0,
      text:
        c.games.rate == null
          ? NOT_ENOUGH
          : `${pctText(c.games.rate)} (${c.games.n} of ${c.games.of})`,
      tone: IDENTITY[c.colour],
    })),
  );

  const troubles: { key: keyof ManaGame; title: string; tone: string; def: string }[] = [
    {
      key: "screw",
      title: "Screwed games",
      tone: "var(--chart-4)",
      def: "2 or more missed land drops with no land in hand, in your first 6 turns.",
    },
    {
      key: "flood",
      title: "Flooded games",
      tone: "var(--chart-2)",
      def: "3 or more turns with more lands than spells in hand, while still hitting drops.",
    },
    {
      key: "colour_screw",
      title: "Colour-screwed games",
      tone: "var(--chart-5)",
      def: "2 or more turns with a spell you could pay for in amount but not colour, at the start of your main phase.",
    },
  ];
  const troubleRate = (k: keyof ManaGame) =>
    k === "screw" ? o.mana.screw : k === "flood" ? o.mana.flood : o.mana.colour_screw;

  const partial = $derived(o.coverage.games - o.coverage.frames);
</script>

<div class="tiles" data-testid="advanced-tiles">
  <StatTile label="Games" value={`${o.record.games}`} sub={recordText(o.record) + " (W-L)"} />
  <StatTile
    label="Win rate"
    value={o.win_rate.rate == null ? "-" : pctText(o.win_rate.rate)}
    sub={rateText(o.win_rate)}
  />
  <StatTile label="Average place" value={avgPlaceText(o.record.avg_place)} />
  <StatTile
    label="Median length"
    value={medianLengthText(o.length)}
    sub={o.length.minutes.median == null
      ? undefined
      : `about ${Math.round(o.length.minutes.median)} minutes`}
  />
  <StatTile
    label="Most-played deck"
    words
    value={o.most_played_deck?.name ?? "-"}
    sub={o.most_played_deck
      ? `${o.most_played_deck.games} ${o.most_played_deck.games === 1 ? "game" : "games"}`
      : undefined}
  />
</div>
{#if partial > 0 && o.coverage.games > 0}
  <p class="coverage">
    {partial} of {o.coverage.games} games have no replay to read, so they count in the record but not
    in Play and Mana.
  </p>
{/if}

<div class="sections">
  <section class="card" aria-labelledby="sec-results" data-testid="stats-results">
    <h2 id="sec-results">Results</h2>
    <ChartFigure
      id="cap-places"
      title="Finishing place"
      caption={placesCaption(o.places)}
      columns={["Table size", "Games", "1st", "2nd", "3rd", "4th"]}
      rows={o.places.map((p) => [
        `${p.players} players`,
        p.games,
        ...[0, 1, 2, 3].map((i) => (i < p.players ? (p.places[i] ?? 0) : "-")),
      ])}
    >
      {#snippet children(id)}
        <div class="stack-rows" role="img" aria-labelledby={id}>
          {#each o.places as p (p.players)}
            <StackedBar
              label={`${p.players} players, ${p.games} ${p.games === 1 ? "game" : "games"}`}
              parts={p.places.slice(0, p.players)}
              names={p.places.slice(0, p.players).map((_, i) => placeText(i + 1))}
              tones={PLACE_TONES}
            />
          {:else}
            <p class="empty">No finished games in this filter.</p>
          {/each}
        </div>
      {/snippet}
    </ChartFigure>

    <ChartFigure
      id="cap-win-kinds"
      title="How you win"
      caption={winKindsCaption(o.win_kinds)}
      columns={["Kind", "Wins"]}
      rows={winKinds.map((k) => [kindText(k.kind), k.games])}
      note={comboRules.length > 0
        ? `Combo wins came by ${comboRules.map((r) => `${COMBO_RULE_TEXT[r.kind] ?? r.kind} (${r.games})`).join(", ")}.`
        : undefined}
    >
      {#snippet children(id)}
        {#if winKinds.length > 0}
          <BarChart labelledby={id} bars={kindBars(winKinds, "var(--chart-1)")} />
        {:else}
          <p class="empty">No wins yet in this filter.</p>
        {/if}
      {/snippet}
    </ChartFigure>

    <ChartFigure
      id="cap-out-kinds"
      title="How you go out"
      caption={outKindsCaption(o.out_kinds)}
      columns={["Kind", "Games"]}
      rows={outKinds.map((k) => [kindText(k.kind), k.games])}
    >
      {#snippet children(id)}
        {#if outKinds.length > 0}
          <BarChart labelledby={id} bars={kindBars(outKinds, "var(--chart-4)")} />
        {:else}
          <p class="empty">You haven't gone out in this filter.</p>
        {/if}
      {/snippet}
    </ChartFigure>

    <ChartFigure
      id="cap-turn-order"
      title="Turn order"
      caption={turnOrderCaption(o.turn_order)}
      columns={["Position", "Win rate", "Wins", "Games", "Avg. place"]}
      rows={o.turn_order.map((r) => [
        placeText(r.position),
        r.wins.rate == null ? "-" : pctText(r.wins.rate),
        r.wins.n,
        r.wins.of,
        avgPlaceText(r.avg_place),
      ])}
      note="Four-seat games only."
    >
      {#snippet children(id)}
        <BarChart labelledby={id} bars={turnBars} max={1} />
      {/snippet}
    </ChartFigure>

    <ChartFigure
      id="cap-length"
      title="Game length"
      caption={lengthCaption(o.length)}
      columns={["Turns", "Games"]}
      rows={o.length.by_turns.map((b) => [b.value, b.games])}
    >
      {#snippet children(id)}
        {#if o.length.by_turns.length > 0}
          <Histogram
            labelledby={id}
            categories={lengthCats}
            values={o.length.by_turns.map((b) => b.games)}
            tone="var(--chart-3)"
            xLabel="Turns"
            marker={o.length.turns.median == null
              ? undefined
              : {
                  at: o.length.turns.median,
                  label: `median ${Math.round(o.length.turns.median)}`,
                  categories: o.length.by_turns.map((b) => b.value),
                }}
          />
        {:else}
          <p class="empty">No game lengths yet.</p>
        {/if}
      {/snippet}
    </ChartFigure>
  </section>

  <section class="card" aria-labelledby="sec-play" data-testid="stats-play">
    <h2 id="sec-play">Play</h2>
    <ChartFigure
      id="cap-aggression"
      title="Attacks over turns"
      caption={aggressionCaption(o.aggression)}
      columns={["Your turn", "Turns", "Attackers a turn", "Damage a turn"]}
      rows={o.aggression.over_turns.map((t, i) => [
        overTurnCats[i],
        t.turns,
        oneDecimal(t.attacks_per_turn),
        oneDecimal(t.damage_per_turn),
      ])}
    >
      {#snippet children(id)}
        {#if o.aggression.over_turns.length > 0}
          <LineChart
            labelledby={id}
            categories={overTurnCats}
            series={[
              {
                name: "Attacking creatures a turn",
                values: o.aggression.over_turns.map((t) => t.attacks_per_turn),
                tone: "var(--chart-1)",
              },
            ]}
            xLabel="Your turn number"
          />
          <p class="panel-title">Damage to players a turn</p>
          <LineChart
            labelledby={id}
            categories={overTurnCats}
            series={[
              {
                name: "Damage to players a turn",
                values: o.aggression.over_turns.map((t) => t.damage_per_turn),
                tone: "var(--chart-2)",
              },
            ]}
          />
        {:else}
          <p class="empty">No turns recorded yet.</p>
        {/if}
      {/snippet}
    </ChartFigure>

    <ChartFigure
      id="cap-spread"
      title="Spread across opponents"
      caption={spreadCaption(o.aggression)}
      columns={["Measure", "Value"]}
      rows={o.aggression.most_hit_share == null
        ? []
        : [
            ["Share at the most-hit opponent", pctText(o.aggression.most_hit_share)],
            ["Evenness (0 all at one, 1 even)", oneDecimal(o.aggression.evenness)],
            ["Games", o.aggression.spread_games],
          ]}
      note="Opponents are never named."
    >
      {#snippet children(id)}
        {#if o.aggression.most_hit_share != null && o.aggression.spread_games >= 3}
          <div role="img" aria-labelledby={id}>
            <StackedBar
              label={`${o.aggression.spread_games} games at three or more seats`}
              parts={[o.aggression.most_hit_share, 1 - o.aggression.most_hit_share]}
              names={["Most-hit opponent", "The rest"]}
              tones={["var(--chart-2)", "var(--chart-6)"]}
            />
            <p class="evenness">
              Evenness {oneDecimal(o.aggression.evenness)}: {evennessWord(o.aggression.evenness)}
            </p>
          </div>
        {:else}
          <p class="empty">{NOT_ENOUGH}.</p>
        {/if}
      {/snippet}
    </ChartFigure>
  </section>

  <section class="card" aria-labelledby="sec-mana" data-testid="stats-mana">
    <h2 id="sec-mana">Mana</h2>
    <ChartFigure
      id="cap-missed"
      title="Missed land drops"
      caption={missedCaption(o.mana)}
      columns={["Land drop", "Missed"]}
      rows={missed.map((m) => [landNumberText(m.land), m.missed])}
      note="A missed drop is one of your turns with no land played. The early ones hurt most."
    >
      {#snippet children(id)}
        {#if missed.length > 0}
          <Histogram
            labelledby={id}
            categories={missedCats}
            values={missed.map((m) => m.missed)}
            muted={missed.map((m) => m.land >= 6)}
            tone="var(--chart-4)"
            xLabel="Land drop you were going for"
          />
        {:else}
          <p class="empty">No land drops recorded yet.</p>
        {/if}
      {/snippet}
    </ChartFigure>

    <div class="trouble-tiles">
      {#each troubles as t (t.key)}
        {@const r = troubleRate(t.key)}
        {@const marks = o.mana.trend.map((g) => g[t.key])}
        <StatTile
          label={t.title}
          value={r.rate == null ? "-" : pctText(r.rate)}
          sub={rateText(r)}
          testid={`trouble-${t.key}`}
        >
          <TrendStrip
            {marks}
            tone={t.tone}
            label={`Last ${marks.length} games: ${marks.filter(Boolean).length} ${t.title.toLowerCase()}`}
          />
          <span class="def">{t.def}</span>
        </StatTile>
      {/each}
    </div>
  </section>

  <section class="card" aria-labelledby="sec-cards" data-testid="stats-cards">
    <h2 id="sec-cards">Cards</h2>
    <h3 class="sub">Most cast</h3>
    {#if o.most_cast.length === 0}
      <p class="empty">No spells cast in this filter.</p>
    {:else}
      <ul class="card-grid" class:names-only={shared}>
        {#each o.most_cast.slice(0, 12) as c (c.name)}
          <li>
            {#if !shared && c.scryfall_id}
              <img
                src={scryfallImageURL(c.scryfall_id, "normal")}
                alt={c.name}
                loading="lazy"
                width="146"
                height="204"
              />
            {/if}
            <span class="cname">{c.name}</span>
            <span class="cmeta"
              >{c.casts}
              {c.casts === 1 ? "cast" : "casts"} in {c.games}
              {c.games === 1 ? "game" : "games"}</span
            >
          </li>
        {/each}
      </ul>
    {/if}
    <h3 class="sub">In the most of your decks</h3>
    {#if !o.library.card_index}
      <p class="empty">Card data isn't loaded on this server.</p>
    {:else if o.library.cards_in_most_decks.length === 0}
      <p class="empty">No card is in more than one of your saved decks yet.</p>
    {:else}
      <ul class="in-decks">
        {#each o.library.cards_in_most_decks as c (c.name)}
          <li><span class="cname">{c.name}</span><span class="cmeta">{c.decks} decks</span></li>
        {/each}
      </ul>
    {/if}
    {#if o.library.avg_mana_value != null}
      <p class="cmeta">
        Average mana value of your decks, weighted by games: {o.library.avg_mana_value.toFixed(2)}.
      </p>
    {/if}
  </section>

  <section class="card" aria-labelledby="sec-colours" data-testid="stats-colours">
    <h2 id="sec-colours">Colours</h2>
    <ChartFigure
      id="cap-colours"
      title="Colours in your games"
      caption={coloursCaption(o.colours)}
      columns={["Colour", "Share", "Games", "Of"]}
      rows={o.colours.map((c) => [
        COLOUR_TEXT[c.colour] ?? c.colour,
        c.games.rate == null ? "-" : pctText(c.games.rate),
        c.games.n,
        c.games.of,
      ])}
      note="Each game counts for every colour in its commanders' identity, so the bars don't add up to 100%."
    >
      {#snippet children(id)}
        <BarChart labelledby={id} bars={colourBars} max={1} />
      {/snippet}
    </ChartFigure>
  </section>

  <section class="card wide" aria-labelledby="sec-decks" data-testid="stats-decks">
    <h2 id="sec-decks">Decks</h2>
    {#if o.decks.length === 0}
      <p class="empty">No decks in this filter.</p>
    {:else}
      <div class="scroll">
        <table class="decks">
          <thead>
            <tr>
              <th scope="col">Deck</th>
              <th scope="col" class="num">Games</th>
              <th scope="col" class="num">Record</th>
              <th scope="col" class="num">Win rate</th>
              <th scope="col" class="num">Avg. place</th>
              <th scope="col">Last results</th>
            </tr>
          </thead>
          <tbody>
            {#each o.decks as d (d.deck_id ?? `name:${d.name}`)}
              <tr>
                <th scope="row" class="name">{d.name}</th>
                <td class="num">{d.record.games}</td>
                <td class="num">{recordText(d.record)}</td>
                <td class="num">{winRateText(d.record)}</td>
                <td class="num">{avgPlaceText(d.record.avg_place)}</td>
                <td class="spark-cell">
                  <Sparkline
                    results={d.last}
                    label={`Last ${d.last.length} results, oldest first: ${d.last.filter((r) => r === "win").length} wins`}
                  />
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
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
  .stack-rows {
    display: grid;
    gap: 14px;
  }
  .panel-title {
    margin: 12px 0 2px;
    font-size: 0.82rem;
    color: var(--fg-muted);
  }
  .evenness {
    margin: 8px 0 0;
    font-size: 0.85rem;
  }
  .trouble-tiles {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 10px;
  }
  .trouble-tiles :global(.tile) {
    background: var(--surface-raised);
  }
  .def {
    margin-top: 4px;
    color: var(--fg-muted);
    font-size: 0.76rem;
    line-height: 1.35;
  }
  .card-grid {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(118px, 1fr));
    gap: 12px;
  }
  .card-grid.names-only {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  }
  .card-grid li {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
  }
  .card-grid img {
    width: 100%;
    height: auto;
    aspect-ratio: 488 / 680;
    border-radius: 5%;
    background: var(--surface-sunken);
    margin-bottom: 4px;
  }
  .cname {
    font-weight: 600;
    font-size: 0.88rem;
    overflow-wrap: anywhere;
  }
  .cmeta {
    color: var(--fg-muted);
    font-size: 0.8rem;
    margin: 0;
  }
  .in-decks {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 6px 16px;
  }
  .in-decks li {
    display: flex;
    justify-content: space-between;
    gap: 8px;
    border-bottom: 1px solid var(--border);
    padding: 5px 0;
  }
  .scroll {
    overflow-x: auto;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.92rem;
  }
  th,
  td {
    padding: 8px 10px;
    text-align: left;
    border-bottom: 1px solid var(--border);
  }
  thead th {
    font: 600 0.7rem/1 var(--font-mono);
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  tbody tr:last-child > * {
    border-bottom: 0;
  }
  th.name {
    font-weight: 600;
    min-width: 12ch;
  }
  .num {
    text-align: right;
    white-space: nowrap;
    font-variant-numeric: tabular-nums;
  }
  .spark-cell {
    min-width: 110px;
  }
  @media (max-width: 819px) {
    .sections {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>

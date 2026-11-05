<script lang="ts">
  // ManaTroubles — missed land drops by land number, then the screw,
  // flood and colour-screw tiles with their small trend (ADR 0217 §4 and
  // §8). The overview and a deck page draw the same pair.
  import ChartFigure from "./ChartFigure.svelte";
  import Histogram from "./Histogram.svelte";
  import StatTile from "./StatTile.svelte";
  import TrendStrip from "./TrendStrip.svelte";
  import {
    landNumberText,
    missedCaption,
    pctText,
    rateText,
    type AdvancedSummary,
    type ManaGame,
  } from "../../advancedStats";

  interface Props {
    mana: AdvancedSummary["mana"];
  }
  const { mana: m }: Props = $props();

  const missed = $derived(m.missed_by_land);
  const missedCats = $derived(missed.map((x) => (x.land >= 6 ? "6+" : `${x.land}`)));

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
    k === "screw" ? m.screw : k === "flood" ? m.flood : m.colour_screw;
</script>

<ChartFigure
  id="cap-missed"
  title="Missed land drops"
  caption={missedCaption(m)}
  columns={["Land drop", "Missed"]}
  rows={missed.map((x) => [landNumberText(x.land), x.missed])}
  note="A missed drop is one of your turns with no land played. The early ones hurt most."
>
  {#snippet children(id)}
    {#if missed.length > 0}
      <Histogram
        labelledby={id}
        categories={missedCats}
        values={missed.map((x) => x.missed)}
        muted={missed.map((x) => x.land >= 6)}
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
    {@const marks = m.trend.map((g) => g[t.key])}
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

<style>
  .empty {
    margin: 0;
    color: var(--fg-muted);
    font-size: 0.9rem;
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
</style>

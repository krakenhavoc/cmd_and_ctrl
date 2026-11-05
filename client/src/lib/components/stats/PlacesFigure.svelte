<script lang="ts">
  // PlacesFigure — finishing place, one stacked bar per table size (ADR
  // 0217 §3 and §8). The overview and a deck page draw the same figure.
  import ChartFigure from "./ChartFigure.svelte";
  import StackedBar from "./StackedBar.svelte";
  import { placeText } from "../../myStats";
  import { placesCaption, type PlaceRow } from "../../advancedStats";

  interface Props {
    places: PlaceRow[];
  }
  const { places }: Props = $props();

  const PLACE_TONES = [
    "var(--chart-1)",
    "var(--chart-2)",
    "var(--chart-3)",
    "var(--chart-4)",
    "var(--chart-5)",
    "var(--chart-6)",
  ];
</script>

<ChartFigure
  id="cap-places"
  title="Finishing place"
  caption={placesCaption(places)}
  columns={["Table size", "Games", "1st", "2nd", "3rd", "4th"]}
  rows={places.map((p) => [
    `${p.players} players`,
    p.games,
    ...[0, 1, 2, 3].map((i) => (i < p.players ? (p.places[i] ?? 0) : "-")),
  ])}
>
  {#snippet children(id)}
    <div class="stack-rows" role="img" aria-labelledby={id}>
      {#each places as p (p.players)}
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

<style>
  .stack-rows {
    display: grid;
    gap: 14px;
  }
  .empty {
    margin: 0;
    color: var(--fg-muted);
    font-size: 0.9rem;
  }
</style>

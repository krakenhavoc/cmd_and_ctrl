<script lang="ts">
  // AggressionFigure — attacks, and damage to players, by own-turn number
  // (ADR 0217 §3 and §8). The overview and a deck page draw the same figure.
  import ChartFigure from "./ChartFigure.svelte";
  import LineChart from "./LineChart.svelte";
  import { aggressionCaption, oneDecimal, type AdvancedSummary } from "../../advancedStats";

  interface Props {
    aggression: AdvancedSummary["aggression"];
  }
  const { aggression: a }: Props = $props();

  const cats = $derived(a.over_turns.map((t) => (t.own_turn >= 12 ? "12+" : `${t.own_turn}`)));
</script>

<ChartFigure
  id="cap-aggression"
  title="Attacks over turns"
  caption={aggressionCaption(a)}
  columns={["Your turn", "Turns", "Attackers a turn", "Damage a turn"]}
  rows={a.over_turns.map((t, i) => [
    cats[i],
    t.turns,
    oneDecimal(t.attacks_per_turn),
    oneDecimal(t.damage_per_turn),
  ])}
>
  {#snippet children(id)}
    {#if a.over_turns.length > 0}
      <LineChart
        labelledby={id}
        categories={cats}
        series={[
          {
            name: "Attacking creatures a turn",
            values: a.over_turns.map((t) => t.attacks_per_turn),
            tone: "var(--chart-1)",
          },
        ]}
        xLabel="Your turn number"
      />
      <p class="panel-title">Damage to players a turn</p>
      <LineChart
        labelledby={id}
        categories={cats}
        series={[
          {
            name: "Damage to players a turn",
            values: a.over_turns.map((t) => t.damage_per_turn),
            tone: "var(--chart-2)",
          },
        ]}
      />
    {:else}
      <p class="empty">No turns recorded yet.</p>
    {/if}
  {/snippet}
</ChartFigure>

<style>
  .panel-title {
    margin: 12px 0 2px;
    font-size: 0.82rem;
    color: var(--fg-muted);
  }
  .empty {
    margin: 0;
    color: var(--fg-muted);
    font-size: 0.9rem;
  }
</style>

<script lang="ts">
  // WinKindsFigure — how the person wins, sorted bars with counts, and
  // the rule behind any combo win (ADR 0217 §3 and §8).
  import ChartFigure from "./ChartFigure.svelte";
  import BarChart from "./BarChart.svelte";
  import type { ChartBar } from "../../statsCharts";
  import {
    COMBO_RULE_TEXT,
    kindText,
    sortedKinds,
    winKindsCaption,
    type KindCount,
  } from "../../advancedStats";

  interface Props {
    winKinds: KindCount[];
    comboRules: KindCount[];
  }
  const { winKinds, comboRules }: Props = $props();

  const kinds = $derived(sortedKinds(winKinds));
  const rules = $derived(sortedKinds(comboRules));
  const bars = $derived<ChartBar[]>(
    kinds.map((k) => ({
      label: kindText(k.kind),
      value: k.games,
      text: `${k.games}`,
      tone: "var(--chart-1)",
    })),
  );
</script>

<ChartFigure
  id="cap-win-kinds"
  title="How you win"
  caption={winKindsCaption(winKinds)}
  columns={["Kind", "Wins"]}
  rows={kinds.map((k) => [kindText(k.kind), k.games])}
  note={rules.length > 0
    ? `Combo wins came by ${rules.map((r) => `${COMBO_RULE_TEXT[r.kind] ?? r.kind} (${r.games})`).join(", ")}.`
    : undefined}
>
  {#snippet children(id)}
    {#if kinds.length > 0}
      <BarChart labelledby={id} {bars} />
    {:else}
      <p class="empty">No wins yet in this filter.</p>
    {/if}
  {/snippet}
</ChartFigure>

<style>
  .empty {
    margin: 0;
    color: var(--fg-muted);
    font-size: 0.9rem;
  }
</style>

<script lang="ts">
  // MatchScore — the score of a best-of-three in the command bar (ADR
  // 0222 §3). Public: every seat and spectator sees it. Nothing is
  // drawn at a Commander table or a best-of-one, which have no score.
  import { L } from "../../labels";
  import { gameOfText, scoreLine } from "../../match";
  import type { MatchView, PlayerView } from "../../protocol";

  const { match, seats }: { match: MatchView; seats: PlayerView[] } = $props();
</script>

<span class="match-score" role="group" aria-label={L.matchScore} data-testid="match-score">
  <span class="game">{gameOfText(match)}</span>
  <span class="score">{scoreLine(match, seats)}</span>
</span>

<style>
  .match-score {
    display: inline-flex;
    align-items: baseline;
    gap: 8px;
    padding: 3px 10px;
    border: 1px solid var(--border-strong);
    border-radius: 999px;
    background: var(--surface-raised, var(--surface));
    font-family: var(--font-mono);
    font-size: 12px;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }
  .game {
    color: var(--fg-muted);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }
  .score {
    color: var(--fg);
  }
</style>

<script lang="ts">
  // PromisesRow renders a small ledger of "I owe you" / "you owe me"
  // promise tokens between the viewer and each opponent. Sits below
  // PlayerHeader on each opponent panel. The viewer's panel doesn't
  // render this row — the totals are mirrored on every opponent
  // header from the viewer's perspective, which is enough.
  //
  // Click semantics (viewer-driven): + button increments the
  // viewer→opponent count; - decrements (clamped at 0). The other
  // direction (opponent→viewer) is read-only here; the opponent
  // adjusts that side on their own header.

  import type { ActionPayload, GameView } from "../../protocol";

  type ActionSender = (type: string, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    view: GameView;
    viewerID: string | null;
    opponentID: string;
    sendAction: ActionSender;
  }

  const { view, viewerID, opponentID, sendAction }: Props = $props();

  const owedToOpponent = $derived(
    viewerID ? (view.promises?.[`${viewerID}->${opponentID}`] ?? 0) : 0,
  );
  const owedToViewer = $derived(
    viewerID ? (view.promises?.[`${opponentID}->${viewerID}`] ?? 0) : 0,
  );

  function bump(delta: number): void {
    if (!viewerID) return;
    const next = Math.max(0, owedToOpponent + delta);
    sendAction("set_promise", {
      from: viewerID,
      to: opponentID,
      count: next,
    });
  }
</script>

{#if viewerID && viewerID !== opponentID}
  <div class="promises" aria-label="promise tokens">
    <span class="seg owed">
      <span class="label" title="they owe you">←</span>
      <span class="count">{owedToViewer}</span>
    </span>
    <span class="seg owe">
      <button
        type="button"
        title="-1 you owe"
        aria-label="decrement promise"
        onclick={() => bump(-1)}>−</button
      >
      <span class="label" title="you owe them">→</span>
      <span class="count">{owedToOpponent}</span>
      <button
        type="button"
        title="+1 you owe"
        aria-label="increment promise"
        onclick={() => bump(1)}>+</button
      >
    </span>
  </div>
{/if}

<style>
  .promises {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 10px;
    color: #6c7a99;
    padding: 1px 10px 0;
  }
  .seg {
    display: inline-flex;
    align-items: center;
    gap: 3px;
  }
  .label {
    font-size: 11px;
    color: #6c7a99;
  }
  .count {
    font-weight: 700;
    color: #c8a86a;
    min-width: 12px;
    text-align: center;
    font-variant-numeric: tabular-nums;
  }
  .seg.owe button {
    width: 14px;
    height: 14px;
    padding: 0;
    border: 1px solid #2e3a55;
    border-radius: 3px;
    background: transparent;
    color: #c8c8c8;
    font: inherit;
    font-size: 10px;
    cursor: pointer;
    line-height: 1;
  }
  .seg.owe button:hover {
    background: #1a2335;
    color: #ffd07a;
    border-color: #ffd07a;
  }
</style>

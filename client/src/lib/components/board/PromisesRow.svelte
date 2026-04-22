<script lang="ts">
  // PromisesRow renders a small ledger of "I owe you" / "you owe me"
  // promise tokens between the viewer and each opponent. Sits below
  // PlayerIdentity on each opponent panel. The viewer's panel doesn't
  // render this row — the totals are mirrored on every opponent
  // identity from the viewer's perspective, which is enough.
  //
  // Click semantics (viewer-driven): + button increments the
  // viewer→opponent count; - decrements (clamped at 0). The other
  // direction (opponent→viewer) is read-only here; the opponent
  // adjusts that side on their own identity.

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
    gap: 10px;
    font-size: 10px;
    color: var(--fg-dim);
    padding: 2px 12px 0;
  }
  .seg {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    border-radius: 999px;
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid rgba(255, 255, 255, 0.05);
  }
  .label {
    font-size: 11px;
    color: var(--fg-dim);
    line-height: 1;
  }
  .count {
    font-weight: 800;
    color: #c8a86a;
    min-width: 12px;
    text-align: center;
    font-variant-numeric: tabular-nums;
  }
  .seg.owe button {
    width: 14px;
    height: 14px;
    padding: 0;
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 50%;
    background: transparent;
    color: var(--fg-dim);
    font: inherit;
    font-size: 10px;
    cursor: pointer;
    line-height: 1;
    box-shadow: none;
    transition:
      border-color 120ms var(--ease),
      color 120ms var(--ease),
      background 120ms var(--ease);
  }
  .seg.owe button:hover {
    background: rgba(255, 208, 122, 0.12);
    color: var(--gold);
    border-color: var(--gold);
  }
</style>

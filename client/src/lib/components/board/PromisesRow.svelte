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

  import type { ActionPayload, ActionType, GameView } from "../../protocol";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

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
  {#if owedToViewer === 0 && owedToOpponent === 0}
    <!-- #2483: nothing owed either way, so nothing on the table. The one
         control shows while this board is hovered or focused, and
         records a promise you make them. -->
    <div class="promises idle" aria-label="promise tokens">
      <button
        type="button"
        class="add"
        title="Record a promise you made this player (a deal, such as not attacking them)"
        aria-label="increment promise"
        onclick={() => bump(1)}>+ Promise</button
      >
    </div>
  {:else}
    <div
      class="promises"
      aria-label="promise tokens"
      title="Promise tokens: deals between you and this player"
    >
      {#if owedToViewer > 0}
        <span class="seg owed" title="Promises this player has made you; they change it">
          <span class="label">Owes you</span>
          <span class="count">{owedToViewer}</span>
        </span>
      {/if}
      <span class="seg owe">
        <span class="label">You owe</span>
        <button
          type="button"
          title="One fewer promise you owe them"
          aria-label="decrement promise"
          onclick={() => bump(-1)}>−</button
        >
        <span class="count">{owedToOpponent}</span>
        <button
          type="button"
          title="One more promise you owe them"
          aria-label="increment promise"
          onclick={() => bump(1)}>+</button
        >
      </span>
    </div>
  {/if}
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
    border: 1px solid color-mix(in srgb, var(--overlay-ink) 5%, transparent);
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
    border: 1px solid var(--overlay-strong);
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
  /* Idle: invisible but still in the tab order, so a keyboard user who
     reaches it sees it (:focus-within on the board) like a pointer
     user hovering the board does. */
  .promises.idle .add {
    opacity: 0;
    padding: 2px 8px;
    border: 1px dashed var(--overlay-strong);
    border-radius: 999px;
    background: transparent;
    color: var(--fg-dim);
    font: inherit;
    font-size: 10px;
    cursor: pointer;
    box-shadow: none;
    transition: opacity 120ms var(--ease);
  }
  :global(.seat-panel:is(:hover, :focus-within)) .promises.idle .add {
    opacity: 1;
  }
  .promises.idle .add:hover {
    color: var(--accent);
    border-color: var(--accent);
  }
  .seg.owe button:hover {
    background: rgba(255, 208, 122, 0.12);
    color: var(--accent);
    border-color: var(--accent);
  }
</style>

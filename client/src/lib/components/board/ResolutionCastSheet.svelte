<script lang="ts">
  // ResolutionCastSheet is "copy it. You may cast the copy" (#2947,
  // CR 707.12, CR 608.2g): an effect made copies of cards and you may
  // cast any of them while it is still resolving. Nobody gets priority
  // until you are done, so the sheet stays up until you have cast the
  // ones you want, or said you cast no more.
  //
  // The server lists every way of casting each copy as this seat's legal
  // moves, each naming the prompt (`resolution_cast`): the copy with each
  // legal set of targets, modes and X, at its offer's price. The sheet
  // shows the copies and those moves, and sends the one picked exactly as
  // the server wrote it. "Don't cast" is always there and always legal.

  import type { ActionType, GameView, LegalMoveView } from "../../protocol";
  import Card from "./Card.svelte";
  import { confirmAction } from "../../dock";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    snap: GameView;
    viewerID: string | null;
    sendAction: (type: ActionType, params?: unknown, player?: string) => void;
  }

  const { snap, viewerID, sendAction }: Props = $props();

  const active = $derived(
    viewerID
      ? (snap.pending_choices?.find(
          (c) => c.kind === "resolution_cast" && c.chooser === viewerID,
        ) ?? null)
      : null,
  );
  const copies = $derived(active?.options ?? []);
  const source = $derived(active?.reason || "The effect");

  // The moves for this prompt, the server's labels without the prompt's
  // own name in front of each.
  const moves = $derived(
    active ? (snap.legal_moves ?? []).filter((m) => m.resolution_cast === active.id) : [],
  );
  const casts = $derived(moves.filter((m) => m.type !== "resolve_choice"));
  const decline = $derived(moves.find((m) => m.type === "resolve_choice") ?? null);

  function moveLabel(m: LegalMoveView): string {
    const prefix = `${source}: `;
    return m.label.startsWith(prefix) ? m.label.slice(prefix.length) : m.label;
  }

  function send(m: LegalMoveView): void {
    sendAction(m.type as ActionType, m.params, m.player);
  }

  function sendDecline(): void {
    // The always-legal answer, written here when the server's move has
    // not arrived yet: the prompt's id and no cards.
    if (decline) send(decline);
    else if (active) sendAction("resolve_choice", { choice_id: active.id, card_ids: [] });
  }

  const primary = $derived(
    casts.length === 1
      ? confirmAction(moveLabel(casts[0]), () => send(casts[0]))
      : casts.length === 0
        ? confirmAction("Don't cast", sendDecline, { enter: false })
        : null,
  );
  const secondary = $derived(
    casts.length > 0 ? [{ id: "decline", label: "Don't cast", onPress: sendDecline }] : [],
  );
</script>

{#if active}
  <DockSheet
    rank="choice"
    label={`${source}: you may cast ${copies.length === 1 ? `the copy of ${copies[0].name}` : "the copies"}`}
    src="CR 707.12"
    width={560}
    sheetKey={`resolution-cast:${active.id}:${copies.length}`}
    {primary}
    {secondary}
  >
    <p class="prompt-hint">
      {source} is still resolving. Cast {copies.length === 1 ? "the copy" : "any of the copies"} now,
      or don't. A copy you don't cast stops existing.
    </p>
    <div class="body">
      {#if copies.length > 0}
        <div class="cards">
          {#each copies as card (card.instance_id)}
            <div class="card"><Card {card} /></div>
          {/each}
        </div>
      {/if}
      {#if casts.length > 1}
        <ul class="casts" aria-label="Ways to cast them">
          {#each casts as m, i (i)}
            <li>
              <button type="button" class="cast" onclick={() => send(m)}>
                {moveLabel(m)}
              </button>
            </li>
          {/each}
        </ul>
      {:else if casts.length === 0 && decline}
        <p class="prompt-hint">
          There is no way to cast {copies.length === 1 ? "it" : "them"} now.
        </p>
      {/if}
    </div>
  </DockSheet>
{/if}

<style>
  .body {
    display: flex;
    gap: 12px;
    align-items: flex-start;
  }
  .cards {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    flex: 0 0 auto;
    max-width: 50%;
  }
  .card {
    width: 120px;
  }
  .casts {
    display: flex;
    flex-direction: column;
    gap: 6px;
    flex: 1;
    max-height: 320px;
    overflow-y: auto;
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .cast {
    width: 100%;
    text-align: left;
  }
</style>

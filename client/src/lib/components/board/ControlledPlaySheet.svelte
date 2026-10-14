<script lang="ts">
  // ControlledPlaySheet is Word of Command's forced play (ADR 0166 §4,
  // CR 723.2): you control another player, and they must play the card
  // you chose from their hand, with you making every choice.
  //
  // The server lists every way of making the play as this seat's legal
  // moves, each naming the prompt (`controlled_play`): the card cast as
  // that player with each legal set of targets, modes and X, paid from
  // their lands, or played if it is a land. The sheet shows the card and
  // those moves, and sends the one picked exactly as the server wrote
  // it, so nothing here re-derives a target or a cost for a card that
  // is not in the viewer's own hand. With no move to make, the only
  // answer is "It can't be played", which the server accepts only when
  // that is true. There is no cancel: the play is owed.

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
          (c) => c.kind === "controlled_play" && c.chooser === viewerID,
        ) ?? null)
      : null,
  );
  const card = $derived(active?.options?.[0] ?? null);
  const playerName = $derived(
    snap.seats.find((s) => s.id === active?.from_player)?.name ?? "that player",
  );
  const source = $derived(active?.reason || "Word of Command");

  // The moves for this prompt, the server's labels without the prompt's
  // own name in front of each.
  const moves = $derived(
    active ? (snap.legal_moves ?? []).filter((m) => m.controlled_play === active.id) : [],
  );
  const plays = $derived(moves.filter((m) => m.type !== "resolve_choice"));
  const cantPlay = $derived(moves.find((m) => m.type === "resolve_choice") ?? null);

  function moveLabel(m: LegalMoveView): string {
    const prefix = `${source}: `;
    return m.label.startsWith(prefix) ? m.label.slice(prefix.length) : m.label;
  }

  // "You do get to choose which lands get tapped" (2009-10-01 ruling).
  // Their untapped lands; with none picked the game chooses, otherwise
  // the rest are locked out of the auto-tapper (`locked_sources`).
  const lands = $derived(
    active
      ? snap.battlefield.cards.filter(
          (c) =>
            c.controller === active.from_player &&
            !c.tapped &&
            (c.type_line ?? "").includes("Land"),
        )
      : [],
  );
  let picked = $state<Set<string>>(new Set());
  let pickedFor = "";
  $effect(() => {
    // A new prompt starts with the game choosing.
    const id = active?.id ?? "";
    if (id !== pickedFor) {
      pickedFor = id;
      picked = new Set();
    }
  });
  function toggleLand(id: string): void {
    const next = new Set(picked);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    picked = next;
  }

  function send(m: LegalMoveView): void {
    let params = m.params;
    if (m.type === "cast_spell" && picked.size > 0) {
      const locked = lands.filter((c) => !picked.has(c.instance_id)).map((c) => c.instance_id);
      params = { ...m.params, locked_sources: locked };
    }
    sendAction(m.type as ActionType, params, m.player);
  }

  const primary = $derived(
    plays.length === 1
      ? confirmAction(moveLabel(plays[0]), () => send(plays[0]))
      : plays.length === 0 && cantPlay
        ? confirmAction("It can't be played", () => send(cantPlay), { enter: false })
        : null,
  );
</script>

{#if active}
  <DockSheet
    rank="choice"
    label={`${source}: play ${card?.name ?? "the card"} as ${playerName}`}
    src="CR 723"
    width={560}
    sheetKey={`controlled-play:${active.id}`}
    {primary}
  >
    <p class="prompt-hint">
      You control <strong>{playerName}</strong> until {source} finishes resolving. They play this card
      if they can, with you making every choice. Only their lands can pay for it.
    </p>
    <div class="body">
      {#if card}
        <div class="card"><Card {card} /></div>
      {/if}
      {#if plays.length > 1}
        <ul class="plays" aria-label="Ways to play it">
          {#each plays as m, i (i)}
            <li>
              <button type="button" class="play" onclick={() => send(m)}>
                {moveLabel(m)}
              </button>
            </li>
          {/each}
        </ul>
      {:else if plays.length === 0 && !cantPlay}
        <p class="prompt-hint">Working out how they can play it…</p>
      {/if}
    </div>
    {#if plays.length > 0 && lands.length > 0}
      <fieldset class="lands">
        <legend>
          Lands to tap: {picked.size === 0 ? "the game chooses" : `${picked.size} picked`}
        </legend>
        {#each lands as land (land.instance_id)}
          <button
            type="button"
            class="land"
            class:on={picked.has(land.instance_id)}
            aria-pressed={picked.has(land.instance_id)}
            onclick={() => toggleLand(land.instance_id)}
          >
            {land.name}
          </button>
        {/each}
      </fieldset>
    {/if}
  </DockSheet>
{/if}

<style>
  .body {
    display: flex;
    gap: 12px;
    align-items: flex-start;
  }
  .card {
    flex: 0 0 120px;
  }
  .plays {
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
  .play {
    width: 100%;
    text-align: left;
  }
  .lands {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin: 10px 0 0;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 6px 8px 8px;
  }
  .land.on {
    border-color: var(--accent);
    background: var(--accent-soft);
  }
</style>

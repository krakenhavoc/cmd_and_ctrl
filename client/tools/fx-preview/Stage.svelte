<script lang="ts">
  // Stage — the fx preview harness's fake four-seat board (#4868).
  //
  // It draws `cards` with the data attributes the real effect layers
  // measure (boardAnchor.ts, stackLinger.ts pileSelector, castFlight.ts):
  // a tile per permanent ([data-instance-id], data-tapped, its art as an
  // <img>, a .badge.pt and a [data-counter] pip), each seat's avatar
  // ([data-seat-id]), its pile bar ([data-pile][data-pile-owner]), its
  // command zone card and its hand ([data-hand-owner]), and the stack's
  // items ([data-stack-item-id]). Over it, the same layers Board.svelte
  // mounts for the table's cues, given `view`: TableEffects, EliminationFx,
  // ExtortScales and DiceLayer. A new `run` remounts them, so each replay
  // starts from a first frame that primes, as a reconnect does.

  import type { GameView } from "../../src/lib/protocol";
  import TableEffects from "../../src/lib/components/board/TableEffects.svelte";
  import EliminationFx from "../../src/lib/components/board/EliminationFx.svelte";
  import ExtortScales from "../../src/lib/components/board/ExtortScales.svelte";
  import DiceLayer from "../../src/lib/components/board/DiceLayer.svelte";
  import { SEATS, artOf, cardsIn, type FxCard, type Seat } from "./table";

  interface Props {
    cards: readonly FxCard[];
    view: GameView;
    run: number;
  }

  const { cards, view, run }: Props = $props();

  let boardEl: HTMLElement | null = $state(null);

  const field = (seat: Seat) => cardsIn(cards, "battlefield", seat).filter((c) => !c.attachedTo);
  const under = (host: string) => cards.filter((c) => c.attachedTo === host);
  const top = (zone: "graveyard" | "exile", seat: Seat) => cardsIn(cards, zone, seat).at(-1);
  const out = (seat: Seat) => view.seats[seat]?.eliminated === true;
</script>

{#snippet tile(c: FxCard, tucked = false)}
  <div
    class="tile"
    class:tucked
    class:tapped={c.tapped}
    data-instance-id={c.id}
    data-tapped={c.tapped ? "true" : "false"}
  >
    <img src={artOf(c)} alt={c.name ?? c.id} />
    {#if c.counter}
      <span class="pip" data-counter={c.counter}>1</span>
    {/if}
    {#if c.pt}
      <span class="badge pt">{c.pt}</span>
    {/if}
  </div>
{/snippet}

<div class="board" bind:this={boardEl}>
  {#each SEATS as s, i (s.id)}
    {@const seat = i as Seat}
    {@const hand = cardsIn(cards, "hand", seat)}
    {@const command = cardsIn(cards, "command", seat)}
    <section
      class="seat-panel panel"
      class:out={out(seat)}
      style:grid-area={`p${i}`}
      style:--seat-color={s.color}
    >
      <header>
        <div class="avatar" data-seat-id={s.id}>{s.name[0]}</div>
        <div class="who">
          <strong>{s.name}</strong>
          <span>40 life</span>
        </div>
        <div class="piles">
          <button type="button" class="pile" data-pile="library" data-pile-owner={s.id}>
            <img class="thumb" src="/card-back-small.jpg" alt="" />
            <span>60</span>
          </button>
          {#each ["graveyard", "exile"] as const as zone (zone)}
            {@const c = top(zone, seat)}
            <button type="button" class="pile" data-pile={zone} data-pile-owner={s.id}>
              {#if c}
                <img class="thumb" src={artOf(c)} alt="" />
              {:else}
                <span class="thumb empty"></span>
              {/if}
              <span>{cardsIn(cards, zone, seat).length}</span>
            </button>
          {/each}
        </div>
        <div class="command" title="Command zone">
          {#each command as c (c.id)}
            <div class="cmd" data-instance-id={c.id}>
              <img src={artOf(c)} alt={c.name ?? c.id} />
            </div>
          {:else}
            <span class="thumb empty"></span>
          {/each}
        </div>
      </header>
      <div class="field">
        {#each field(seat) as c (c.id)}
          <div class="slot" class:holds={under(c.id).length > 0}>
            {#each under(c.id) as a (a.id)}
              {@render tile(a, true)}
            {/each}
            {@render tile(c)}
          </div>
        {/each}
      </div>
      <div class="hand" data-hand-owner={s.id}>
        {#each hand as c (c.id)}
          {#if seat === 0}
            <div class="in-hand" data-instance-id={c.id}>
              <img src={artOf(c)} alt={c.name ?? c.id} />
            </div>
          {:else}
            <img class="in-hand" src="/card-back-small.jpg" alt="" />
          {/if}
        {/each}
      </div>
    </section>
  {/each}

  <div class="stack">
    {#each cardsIn(cards, "stack") as c (c.id)}
      <div class="item" data-stack-item-id={c.item ?? c.id}>
        <img src={artOf(c, "normal")} alt={c.name ?? c.id} />
      </div>
    {/each}
  </div>

  {#key run}
    <TableEffects {view} {boardEl} />
    <EliminationFx {view} {boardEl} />
    <ExtortScales {view} {boardEl} viewerID={SEATS[0].id} />
    <DiceLayer {view} {boardEl} />
  {/key}
</div>

<style>
  .board {
    position: relative;
    display: grid;
    grid-template-areas:
      "p1 p2"
      "p0 p3";
    grid-template-columns: 1fr 1fr;
    grid-template-rows: 1fr 1fr;
    gap: 8px;
    padding: 6px;
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    overflow: hidden;
    background: var(--surface-sunken);
    border-radius: var(--radius-lg);
  }
  .seat-panel {
    display: grid;
    grid-template-rows: auto 1fr auto;
    gap: 8px;
    padding: 10px;
    min-height: 0;
    border-radius: var(--radius-lg);
  }
  .seat-panel.out {
    filter: grayscale(1) brightness(0.6);
  }
  header {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .avatar {
    width: 40px;
    height: 40px;
    border-radius: 999px;
    display: grid;
    place-items: center;
    font: 600 18px var(--font-display);
    color: var(--bg);
    background: var(--seat-color);
    flex: none;
  }
  .who {
    display: grid;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .who strong {
    color: var(--fg);
    font-size: 14px;
  }
  .piles,
  .command {
    display: flex;
    gap: 6px;
    margin-left: auto;
  }
  .command {
    margin-left: 0;
  }
  .pile {
    display: grid;
    justify-items: center;
    gap: 2px;
    padding: 0;
    border: 0;
    background: none;
    color: var(--fg-muted);
    font: 11px var(--font-mono);
  }
  .thumb,
  .cmd {
    display: block;
    width: 36px;
    height: 50px;
    border-radius: var(--radius-sm);
    object-fit: cover;
    overflow: hidden;
  }
  .thumb.empty {
    box-sizing: border-box;
    border: 1px dashed var(--border-strong);
  }
  .cmd img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .field {
    display: flex;
    flex-wrap: wrap;
    align-content: flex-start;
    gap: 10px;
    min-height: 0;
  }
  .slot {
    position: relative;
    isolation: isolate;
  }
  .slot.holds {
    margin-left: 22px;
  }
  .tile {
    position: relative;
    width: 72px;
    height: 100px;
    border-radius: var(--radius-sm);
    box-shadow: var(--shadow-sm);
  }
  .tile.tapped {
    transform: rotate(90deg);
  }
  .tile.tucked {
    position: absolute;
    left: -22px;
    top: 8px;
    z-index: -1;
  }
  .tile img {
    width: 100%;
    height: 100%;
    border-radius: var(--radius-sm);
    object-fit: cover;
    display: block;
  }
  .badge.pt,
  .pip {
    position: absolute;
    padding: 1px 5px;
    border-radius: var(--radius-sm);
    font: 600 11px var(--font-mono);
    background: var(--surface-raised);
    color: var(--fg);
    border: 1px solid var(--border-strong);
  }
  .badge.pt {
    right: 3px;
    bottom: 3px;
  }
  .pip {
    left: 3px;
    top: 3px;
  }
  .hand {
    display: flex;
    gap: 4px;
    min-height: 50px;
  }
  .in-hand {
    width: 36px;
    height: 50px;
    border-radius: var(--radius-sm);
    overflow: hidden;
    object-fit: cover;
  }
  .in-hand img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .stack {
    position: absolute;
    left: 50%;
    top: 50%;
    translate: -50% -50%;
    display: flex;
    gap: 8px;
    z-index: 2;
  }
  .item {
    width: 120px;
    height: 168px;
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    overflow: hidden;
  }
  .item img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
</style>

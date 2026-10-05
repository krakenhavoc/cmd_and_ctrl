<script lang="ts">
  // OpeningRollBanner — the attention strip's `opening roll` banner
  // (ADR 0121 §6). While the opening roll is open it takes the place of
  // the `opening hand decisions` roll call: one chip per seat in the
  // seat's colour, with its d20, "rolling…" while it still owes one, or
  // "—" when it is not in this round. In a reroll the round's tied
  // leaders are outlined, and once one leader is left that seat is
  // marked as the one who chooses who goes first.
  //
  // Everyone sees it, spectators included: the dice are public the
  // moment they land. The dice layer (DiceLayer.svelte) draws the same
  // dice at each seat; the strip raises no cue of its own for them.
  //
  // A chip shows its result only once that die has settled in the dice
  // layer (the game screen's dice queue says when), so the banner never
  // gives a number away mid-tumble; until then it still reads
  // "rolling…", and the chooser is not marked. Mounted with no queue (a
  // render test), results show at once.
  import type { GameView } from "../../protocol";
  import { seatColor } from "../../colors";
  import { L } from "../../labels";
  import {
    latestOpeningDice,
    openingChipText,
    openingRollModel,
    type OpeningChip,
  } from "../../openingRoll";
  import { useDiceQueue, type DiceQueue } from "../../diceQueue.svelte";
  import Icon from "../Icon.svelte";

  interface Props {
    view: GameView;
    // The game screen's dice schedule. Defaults to the one Game.svelte
    // provides; null shows every result at once.
    dice?: DiceQueue | null;
  }
  const { view, dice: given }: Props = $props();

  const provided = useDiceQueue();
  const dice = $derived(given === undefined ? provided : given);
  const model = $derived(openingRollModel(view));
  const latestSeq = $derived(
    new Map(latestOpeningDice(view.log).map((l) => [l.seat, l.seq] as const)),
  );
  const chips = $derived.by((): OpeningChip[] => {
    if (!model) return [];
    void dice?.tick;
    const landed = (chip: OpeningChip): boolean => {
      if (chip.state !== "rolled" || !dice) return true;
      const seq = latestSeq.get(chip.seat);
      return seq === undefined || dice.released(seq);
    };
    const pending = model.chips.some((c) => !landed(c));
    return model.chips.map(
      (c): OpeningChip =>
        landed(c)
          ? { ...c, chooser: c.chooser && !pending }
          : { ...c, state: "rolling", result: undefined, chooser: false },
    );
  });
</script>

{#if model}
  <div class="opening-roll-banner" role="group" aria-label={L.openingRoll}>
    <span class="label">Opening roll</span>
    {#each chips as chip (chip.seat)}
      <span
        class="chip"
        class:rolled={chip.state === "rolled"}
        class:rolling={chip.state === "rolling"}
        class:out={chip.state === "out"}
        class:tied={chip.tied}
        class:chooser={chip.chooser}
        style="--seat-color: {seatColor(chip.seat)}"
        data-seat={chip.seat}
      >
        <span class="seat-dot" style="background:{seatColor(chip.seat)}"></span>
        <b>{chip.name}</b>
        <span class="result">{openingChipText(chip)}</span>
        {#if chip.chooser}
          <span class="chooses"><Icon name="crown" size={11} /> chooses</span>
        {/if}
      </span>
    {/each}
  </div>
{/if}

<style>
  /* The strip's card shell (Game.svelte's .att), so it reads as one of
     the strip's rows. */
  .opening-roll-banner {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    padding: 9px 10px 9px 14px;
    background: color-mix(in srgb, var(--surface) 94%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.35;
    box-sizing: border-box;
  }
  .label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    font-weight: 700;
    color: var(--accent-strong);
    white-space: nowrap;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-size: 12px;
    color: var(--fg-muted);
    padding: 3px 9px;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--seat-color) 40%, var(--border));
    background: color-mix(in srgb, var(--seat-color) 8%, transparent);
  }
  .chip b {
    color: var(--fg);
    font-weight: 600;
  }
  .seat-dot {
    display: inline-block;
    width: 7px;
    height: 7px;
    border-radius: 50%;
  }
  .result {
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
    min-width: 2ch;
    text-align: right;
  }
  .chip.rolled .result {
    color: var(--fg);
    font-weight: 700;
    font-size: 13px;
  }
  .chip.rolling .result {
    color: var(--accent-strong);
  }
  .chip.out {
    opacity: 0.55;
  }
  /* A tied leader rolling again: outlined in its own colour. */
  .chip.tied {
    border-color: var(--seat-color);
    box-shadow: 0 0 0 1px var(--seat-color);
  }
  .chip.chooser {
    border-color: var(--seat-color);
    background: color-mix(in srgb, var(--seat-color) 22%, transparent);
    box-shadow: 0 0 0 1px var(--seat-color);
  }
  .chooses {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    font-weight: 700;
    color: var(--accent-strong);
  }
</style>

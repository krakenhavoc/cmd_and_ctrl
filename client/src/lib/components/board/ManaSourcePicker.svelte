<script lang="ts">
  // ManaSourcePicker — the popover a left-click on a mana source with
  // more than one mana ability opens, anchored at that card (#1438).
  //
  // One option per ability, in the server's order, each drawn as the
  // mana it makes plus whatever else it does ("deals 1 damage to
  // you"). Picking one hands the ability index to Board, which routes
  // it exactly as the right-click menu's row would — straight to
  // `activate_mana_ability`, or first through the sacrifice / discard
  // / counter picker when the cost needs an answer. Escape, a click
  // outside, or clicking the card again closes it with nothing sent,
  // so nothing is tapped.
  //
  // position: fixed and mounted by Board, never inside the Card: the
  // card's tap rotation is a CSS transform, and a transformed ancestor
  // would pin a fixed box to the card instead of the viewport.
  import { tick } from "svelte";
  import type { CardView, GameView } from "../../protocol";
  import { findCard, locateCard } from "../../contextMenu.logic";
  import { manaAbilityOptions, placePopover, type ManaPickOption } from "../../manaSource";
  import type { ManaSourcePickerOpen } from "../../manaSourcePicker";
  import ModalLayer from "../ModalLayer.svelte";
  import ManaSymbolPicker from "./ManaSymbolPicker.svelte";

  interface Props {
    view: GameView;
    open: ManaSourcePickerOpen;
    onPick: (card: CardView, abilityIndex: number) => void;
    onClose: () => void;
  }

  const { view, open, onPick, onClose }: Props = $props();

  // Re-read from the live snapshot so a flag the server flips while
  // the picker is open (an exhaust spent elsewhere) greys its option.
  const card = $derived(findCard(view, open.cardID));
  const onBattlefield = $derived(locateCard(view, open.cardID)?.zone === "battlefield");
  const options = $derived(card ? manaAbilityOptions(card) : []);

  // The source left the battlefield (or lost its mana abilities) while
  // the picker was open: there is nothing left to pick.
  $effect(() => {
    if (!card || !onBattlefield || options.length === 0) onClose();
  });

  let el: HTMLDivElement | null = $state(null);
  let left = $state(-9999);
  let top = $state(-9999);
  let side = $state<"above" | "below">("above");

  async function place(): Promise<void> {
    await tick();
    const node = el;
    if (!node) return;
    const p = placePopover(
      open.anchor,
      node.offsetWidth,
      node.offsetHeight,
      window.innerWidth,
      window.innerHeight,
    );
    left = p.left;
    top = p.top;
    side = p.side;
    node.querySelector<HTMLButtonElement>("button:not([disabled])")?.focus({ preventScroll: true });
  }

  $effect(() => {
    // Re-place when the anchor moves (a second card clicked) or the
    // option count changes the box's size.
    void open.anchor;
    void options.length;
    if (el) place();
  });

  // Outside click closes. Capture phase, so it runs before the click
  // reaches a card: clicking a DIFFERENT land closes this picker and
  // then opens that one, and clicking THIS card again is left to its
  // own handler, which toggles the picker shut rather than reopening.
  $effect(() => {
    function onDown(ev: MouseEvent): void {
      const target = ev.target as Node | null;
      if (!target || !el) return;
      if (el.contains(target)) return;
      const source = (target as Element).closest?.("[data-instance-id]");
      if (source?.getAttribute("data-instance-id") === open.cardID) return;
      onClose();
    }
    window.addEventListener("click", onDown, true);
    return () => window.removeEventListener("click", onDown, true);
  });

  function pick(o: ManaPickOption): void {
    if (!card || o.abilityIndex === undefined) return;
    onClose();
    onPick(card, o.abilityIndex);
  }
</script>

{#if card}
  <ModalLayer />
  <div
    bind:this={el}
    class="mana-source-picker"
    class:below={side === "below"}
    style:left={`${left}px`}
    style:top={`${top}px`}
    role="dialog"
    aria-label={`Tap ${card.name} for mana`}
  >
    <div class="head">
      <span class="title">Tap <strong>{card.name}</strong> for</span>
      <button type="button" class="close" aria-label="cancel" title="Cancel (Esc)" onclick={onClose}
        >×</button
      >
    </div>
    <ManaSymbolPicker
      {options}
      onPick={pick}
      onCancel={onClose}
      label={`mana abilities of ${card.name}`}
    />
  </div>
{/if}

<style>
  .mana-source-picker {
    position: fixed;
    z-index: 70;
    box-sizing: border-box;
    max-width: calc(100vw - 32px);
    padding: 8px 10px 10px;
    background: rgba(12, 16, 30, 0.97);
    color: var(--gold, #e0c890);
    border: 1px solid rgba(200, 168, 106, 0.6);
    border-radius: 12px;
    box-shadow: 0 14px 32px rgba(0, 0, 0, 0.65);
  }
  .head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 8px;
    font-size: 12px;
  }
  .title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .close {
    flex: 0 0 auto;
    width: 44px;
    height: 32px;
    margin: -6px -6px -6px 0;
    background: transparent;
    color: inherit;
    border: none;
    border-radius: 6px;
    font-size: 18px;
    line-height: 1;
    cursor: pointer;
  }
  .close:hover,
  .close:focus-visible {
    background: rgba(200, 168, 106, 0.15);
    outline: none;
  }
</style>

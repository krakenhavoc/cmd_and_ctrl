<script lang="ts">
  // ManaSourcePicker — the popover a left-click on a mana source with
  // more than one mana ability opens, anchored at that card (#1438).
  //
  // One option per final result, in the server's order, each drawn as
  // the mana it makes plus whatever else it does ("deals 1 damage to
  // you"). #1443: an ability whose output is a choice of colours is
  // already expanded into one option per colour (manaAbilityOptions,
  // from the server's `color_options`), so a painland is {C}, {R} and
  // {W} here and Birds of Paradise its five colours — and nothing is
  // tapped until one is picked. Picking one hands the ability index,
  // and the colours it names, to Board, which routes
  // it exactly as the right-click menu's row would — straight to
  // `activate_mana_ability`, or first through the sacrifice / discard
  // / counter picker when the cost needs an answer. Escape, a click
  // outside, or clicking the card again closes it with nothing sent,
  // so nothing is tapped.
  //
  // ADR 0117 §4: an ability with two or more picking slots (Vivi
  // Ornitier, Relic of Sauron, a filter land) is one option here, and
  // picking it turns this picker into the per-colour stepper in place
  // (ManaSplitStepper), named "Split N mana from <card>". Opened on one
  // ability (`open.abilityIndex`: a lone such ability clicked, or the
  // popover's mana row with a colour choice), the picker lists only
  // that ability, and goes straight to its stepper when it has one.
  //
  // position: fixed and mounted by Board, never inside the Card: the
  // card's tap rotation is a CSS transform, and a transformed ancestor
  // would pin a fixed box to the card instead of the viewport.
  import { tick } from "svelte";
  import type { CardView, GameView } from "../../protocol";
  import { findCard, locateCard } from "../../contextMenu.logic";
  import { manaAbilityOptions, placePopover, type ManaPickOption } from "../../manaSource";
  import type { ManaSourcePickerOpen } from "../../manaSourcePicker";
  import { splitMemoryKey, stepperSlots } from "../../manaStepper";
  import ModalLayer from "../ModalLayer.svelte";
  import ManaSplitStepper from "./ManaSplitStepper.svelte";
  import ManaSymbolPicker from "./ManaSymbolPicker.svelte";

  interface Props {
    view: GameView;
    open: ManaSourcePickerOpen;
    onPick: (card: CardView, abilityIndex: number, colors?: string[]) => void;
    onClose: () => void;
  }

  const { view, open, onPick, onClose }: Props = $props();

  // Re-read from the live snapshot so a flag the server flips while
  // the picker is open (an exhaust spent elsewhere) greys its option.
  const card = $derived(findCard(view, open.cardID));
  const onBattlefield = $derived(locateCard(view, open.cardID)?.zone === "battlefield");
  // ADR 0117 §2: greyed by the popover's own predicate, which reads the
  // paying player's life for a "Pay N life" cost: the controller's.
  const payerLife = $derived(
    card ? view.seats.find((s) => s.id === (card.controller || card.owner))?.life : undefined,
  );
  const options = $derived(
    card
      ? manaAbilityOptions(card, { payerLife }).filter(
          (o) => open.abilityIndex === undefined || o.abilityIndex === open.abilityIndex,
        )
      : [],
  );
  const optionCount = $derived(options.length);

  // ADR 0117 §4: the ability whose stepper is showing. Chosen from the
  // list (a split option picked), or the one ability the picker was
  // opened on when that ability is a live split.
  // Held with the `open` it was chosen in, so a picker re-opened on
  // another card (the same component, a new `open`) starts on its list.
  let chosen = $state.raw<{ open: ManaSourcePickerOpen; index: number } | null>(null);
  const chosenSplit = $derived(chosen && chosen.open === open ? chosen.index : null);
  const openedSplit = $derived(
    open.abilityIndex !== undefined && options.length === 1 && options[0].split
      ? open.abilityIndex
      : null,
  );
  const stepIndex = $derived(chosenSplit ?? openedSplit);
  // Re-read from every frame: Vivi's power is the number of lists.
  const stepLists = $derived.by(() => {
    if (stepIndex === null || !card) return null;
    const a = (card.mana_abilities ?? []).find((m) => m.index === stepIndex);
    return a ? stepperSlots(a) : null;
  });
  const stepTotal = $derived(stepLists?.length ?? 0);
  const stepping = $derived(stepLists !== null);
  const stepDisabled = $derived(options.find((o) => o.abilityIndex === stepIndex)?.disabled ?? "");

  // The source left the battlefield (or lost its mana abilities) while
  // the picker was open: there is nothing left to pick.
  $effect(() => {
    if (!card || !onBattlefield || options.length === 0) onClose();
  });

  let el: HTMLDivElement | null = $state(null);
  let left = $state(-9999);
  let top = $state(-9999);
  let side = $state<"above" | "below">("above");

  // Focus moves in only when what is shown changes (opened, or turned
  // into the stepper), never on an ordinary frame: a snapshot landing
  // while the player steps must not pull focus off the row.
  let focusedFor = "";

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
    const shown = `${open.cardID}:${stepping ? `split-${stepIndex}` : "options"}`;
    if (shown === focusedFor) return;
    focusedFor = shown;
    const first = stepping
      ? (node.querySelector<HTMLButtonElement>("button[data-add-mana]:not([disabled])") ??
        node.querySelector<HTMLButtonElement>(".split button:not([disabled])"))
      : node.querySelector<HTMLButtonElement>("button:not([disabled])");
    first?.focus({ preventScroll: true });
  }

  $effect(() => {
    // Re-place when the anchor moves (a second card clicked), the
    // option count changes the box's size, or the stepper opens or
    // changes its total.
    void open.anchor;
    void optionCount;
    void stepping;
    void stepTotal;
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
    // Read the card BEFORE closing. Board mounts this picker off the
    // store that onClose empties, so after it `open` is null and the
    // derived `card` would throw re-reading `open.cardID` — the pick
    // was lost in the real Board, which #1438's tests (mounting the
    // picker alone, with a fixed `open`) could not see.
    const source = card;
    if (!source || o.abilityIndex === undefined) return;
    // ADR 0117 §4: a split option opens its stepper in place.
    if (o.split) {
      chosen = { open, index: o.abilityIndex };
      return;
    }
    onClose();
    onPick(source, o.abilityIndex, o.colors);
  }

  function confirmSplit(colors: string[]): void {
    const source = card;
    const index = stepIndex;
    if (!source || index === null) return;
    onClose();
    onPick(source, index, colors);
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
    aria-label={stepping
      ? `Split ${stepTotal} mana from ${card.name}`
      : `Tap ${card.name} for mana`}
  >
    <div class="head">
      {#if stepping}
        <span class="title"
          >Split <strong>{stepTotal}</strong> mana from <strong>{card.name}</strong></span
        >
      {:else}
        <span class="title">Tap <strong>{card.name}</strong> for</span>
      {/if}
      <button type="button" class="close" aria-label="cancel" title="Cancel (Esc)" onclick={onClose}
        >×</button
      >
    </div>
    {#if stepping && stepLists && stepIndex !== null}
      {@const memoryKey = splitMemoryKey(view.id, card.instance_id, stepIndex)}
      <!-- Keyed, so another card's stepper starts from its own state. -->
      {#key memoryKey}
        <ManaSplitStepper
          lists={stepLists}
          {memoryKey}
          disabled={stepDisabled}
          onConfirm={confirmSplit}
          onCancel={onClose}
        />
      {/key}
    {:else}
      <ManaSymbolPicker
        {options}
        onPick={pick}
        onCancel={onClose}
        label={`mana abilities of ${card.name}`}
      />
    {/if}
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

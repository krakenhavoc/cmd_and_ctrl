<script lang="ts">
  // CardContextMenu — the #170 admin override menu. Right-clicking any
  // card (in any zone) while settings.gameplay.adminOverrides is on
  // opens this panel at the cursor with every manual override the
  // server already supports for that card in that zone.
  //
  // Why it exists: most of the catalog is not automated yet, and the
  // automated part sometimes gets a card wrong. Without a manual
  // escape hatch the table simply stalls. Every option here maps to
  // an action the engine already implements, so "do it by hand" is
  // always available.
  //
  // Positioning follows ManaAbilityMenu's precedent (anchored popover,
  // Escape + click-outside dismiss) but uses position: fixed against
  // the cursor rather than an anchor element — the menu has to be able
  // to open over the zone-browser modal, which is a separate stacking
  // context from any one card.
  //
  // Submenus are drill-down rather than flyout: clicking a row with
  // children replaces the panel body and adds a "back" row. One
  // fixed-position box means no second layer of overflow maths, and
  // it works the same on a narrow window.

  import { tick } from "svelte";
  import type { ActionPayload, ActionType, CardView, GameView } from "../../protocol";
  import type { CardMenuOpen } from "../../contextMenu";
  import {
    ZONE_LABELS,
    buildMenuSections,
    findCard,
    locateCard,
    type MenuActivate,
    type MenuAction,
    type MenuItem,
  } from "../../contextMenu.logic";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    view: GameView;
    viewerID: string | null;
    isAdmin: boolean;
    open: CardMenuOpen;
    sendAction: ActionSender;
    // Ability rows hand back to Board, which owns the sacrifice
    // picker and the targeting flow an activation may need.
    onActivate: (card: CardView, activate: MenuActivate) => void;
    onClose: () => void;
  }

  const { view, viewerID, isAdmin, open, sendAction, onActivate, onClose }: Props = $props();

  // Re-resolve the card from the live snapshot so counter totals,
  // tapped state and marked damage stay current while the menu is
  // open. Falls back to the CardView captured at right-click time if
  // the card has left the game (it can't, but the menu shouldn't
  // crash if it does).
  const card = $derived(findCard(view, open.card.instance_id) ?? open.card);
  const where = $derived(locateCard(view, open.card.instance_id));
  const zoneLabel = $derived(where ? ZONE_LABELS[where.zone] : "gone");
  const sections = $derived(buildMenuSections(view, card, viewerID, isAdmin));

  // trail is the drill-down path, held as item IDs rather than item
  // objects so a snapshot arriving while a submenu is open refreshes
  // its disabled flags instead of showing the capture from the click.
  let trail = $state<string[]>([]);
  const drilled = $derived.by(() => {
    let items: MenuItem[] = sections.flatMap((s) => s.items);
    let label = "";
    for (const id of trail) {
      const found = items.find((i) => i.id === id);
      if (!found?.items) return { label, items: [] as MenuItem[] };
      label = found.label;
      items = found.items;
    }
    return { label, items };
  });

  // promptItem is the item whose inline form is showing, if any.
  let promptItem = $state<MenuItem | null>(null);
  let counterName = $state("");
  let counterDelta = $state(1);
  let damageDelta = $state(1);

  let el: HTMLDivElement | null = $state(null);
  let left = $state(open.x);
  let top = $state(open.y);

  // place clamps the panel inside the viewport once it has a measured
  // size, then moves focus to the first thing in it. Called on mount
  // and after every content swap, because drilling into a submenu
  // changes both the panel's height and what should hold focus.
  async function place(): Promise<void> {
    await tick();
    const node = el;
    if (!node) return;
    const maxX = Math.max(8, window.innerWidth - node.offsetWidth - 8);
    const maxY = Math.max(8, window.innerHeight - node.offsetHeight - 8);
    left = Math.max(8, Math.min(open.x, maxX));
    top = Math.max(8, Math.min(open.y, maxY));
    node.querySelector<HTMLElement>("input, button:not([disabled])")?.focus();
  }

  $effect(() => {
    if (el) place();
  });

  function fire(action: MenuAction): void {
    sendAction(action.type, action.params, action.player);
  }

  function choose(item: MenuItem): void {
    if (item.disabled) return;
    if (item.items) {
      trail = [...trail, item.id];
      place();
      return;
    }
    if (item.prompt) {
      promptItem = item;
      counterName = "";
      counterDelta = 1;
      damageDelta = 1;
      place();
      return;
    }
    if (item.activate) {
      onActivate(card, item.activate);
      onClose();
      return;
    }
    if (item.action) {
      fire(item.action);
      // Incremental rows (counters, damage) stay open so a second
      // click lands without re-opening the menu.
      if (item.repeat) return;
    }
    onClose();
  }

  function goBack(): void {
    trail = trail.slice(0, -1);
    place();
  }

  function closePrompt(): void {
    promptItem = null;
    place();
  }

  function applyPrompt(): void {
    const item = promptItem;
    if (!item) return;
    if (item.prompt === "custom_counter") {
      const name = counterName.trim();
      if (!name) return;
      const params = { instance_id: card.instance_id, name, delta: counterDelta };
      sendAction("add_counter", params);
    } else {
      sendAction("mark_damage", { instance_id: card.instance_id, delta: damageDelta });
    }
    onClose();
  }

  function onKey(ev: KeyboardEvent): void {
    if (ev.key !== "Escape") return;
    ev.preventDefault();
    ev.stopPropagation();
    onClose();
  }

  // Click-outside dismiss. pointerdown (not click) so the menu closes
  // on the press that starts a right-click elsewhere, before that
  // card's own contextmenu handler reopens it on the new card.
  function onPointerDown(ev: PointerEvent): void {
    const node = el;
    if (!node) return;
    if (ev.target instanceof Node && node.contains(ev.target)) return;
    onClose();
  }

  // Suppress the browser's own context menu while the pointer is over
  // ours; everywhere else it behaves normally.
  function onContext(ev: MouseEvent): void {
    const node = el;
    if (!node) return;
    if (ev.target instanceof Node && node.contains(ev.target)) ev.preventDefault();
  }
</script>

<svelte:window onkeydown={onKey} onpointerdown={onPointerDown} oncontextmenu={onContext} />

{#snippet row(item: MenuItem)}
  <button
    type="button"
    class="ctx-item"
    class:danger={item.danger}
    role="menuitem"
    disabled={item.disabled}
    title={item.hint || item.label}
    onclick={(ev) => {
      ev.stopPropagation();
      choose(item);
    }}
  >
    <span class="ctx-text">{item.label}</span>
    {#if item.items}
      <span class="ctx-more" aria-hidden="true">›</span>
    {:else if item.prompt}
      <span class="ctx-more" aria-hidden="true">…</span>
    {/if}
  </button>
{/snippet}

<div
  bind:this={el}
  class="ctx-menu"
  role="menu"
  aria-label="card overrides"
  style="left: {left}px; top: {top}px"
>
  <div class="ctx-head">
    <span class="ctx-name">{card.name || "face-down card"}</span>
    <span class="ctx-zone">{zoneLabel}</span>
  </div>
  {#if promptItem}
    <div class="ctx-form">
      {#if promptItem.prompt === "custom_counter"}
        <label class="ctx-field">
          <span>counter</span>
          <input type="text" bind:value={counterName} placeholder="quest" />
        </label>
        <label class="ctx-field">
          <span>delta</span>
          <input type="number" step="1" bind:value={counterDelta} />
        </label>
      {:else}
        <label class="ctx-field">
          <span>damage</span>
          <input type="number" step="1" bind:value={damageDelta} />
        </label>
      {/if}
      <div class="ctx-form-row">
        <button type="button" class="ctx-item" onclick={closePrompt}>cancel</button>
        <button type="button" class="ctx-item apply" onclick={applyPrompt}>apply</button>
      </div>
    </div>
  {:else if trail.length > 0}
    <button type="button" class="ctx-item back" onclick={goBack}>‹ back</button>
    <div class="ctx-label">{drilled.label}</div>
    {#each drilled.items as item (item.id)}
      {@render row(item)}
    {/each}
  {:else if sections.length === 0}
    <p class="ctx-empty">
      No overrides here — you can only override cards you control (admins can override any).
    </p>
  {:else}
    {#each sections as sec (sec.id)}
      {#if sec.label}
        <div class="ctx-label">{sec.label}</div>
      {/if}
      {#each sec.items as item (item.id)}
        {@render row(item)}
      {/each}
    {/each}
  {/if}
</div>

<style>
  .ctx-menu {
    position: fixed;
    z-index: 3000;
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 4px;
    min-width: 210px;
    max-width: 280px;
    max-height: min(70vh, 620px);
    overflow-y: auto;
    background: rgba(12, 16, 30, 0.98);
    color: var(--gold);
    border: 1px solid rgba(200, 168, 106, 0.55);
    border-radius: 6px;
    box-shadow: 0 14px 34px rgba(0, 0, 0, 0.7);
    font-size: 11px;
  }
  .ctx-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 8px;
    padding: 4px 6px 6px;
    border-bottom: 1px solid rgba(200, 168, 106, 0.3);
    margin-bottom: 2px;
  }
  .ctx-name {
    font-weight: 700;
    color: var(--gold-strong, #ffd07a);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .ctx-zone {
    flex: 0 0 auto;
    font-family: ui-monospace, Menlo, monospace;
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    opacity: 0.6;
  }
  .ctx-label {
    padding: 5px 6px 2px;
    font-family: ui-monospace, Menlo, monospace;
    font-size: 9px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    opacity: 0.55;
  }
  .ctx-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 5px 8px;
    background: transparent;
    color: inherit;
    border: 1px solid transparent;
    border-radius: 4px;
    text-align: left;
    cursor: pointer;
    font-size: 11px;
    line-height: 1.2;
  }
  .ctx-item:hover:not(:disabled),
  .ctx-item:focus-visible:not(:disabled) {
    background: rgba(200, 168, 106, 0.15);
    border-color: rgba(200, 168, 106, 0.4);
    outline: none;
  }
  .ctx-item:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .ctx-item.danger {
    color: #ff9c9c;
  }
  .ctx-item.danger:hover:not(:disabled) {
    background: rgba(120, 20, 20, 0.35);
    border-color: rgba(255, 122, 122, 0.5);
  }
  .ctx-item.back {
    opacity: 0.75;
  }
  .ctx-text {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .ctx-more {
    opacity: 0.6;
    font-weight: 700;
  }
  .ctx-empty {
    margin: 0;
    padding: 6px;
    opacity: 0.7;
    line-height: 1.35;
  }
  .ctx-form {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 4px 6px 6px;
  }
  .ctx-field {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  .ctx-field span {
    font-family: ui-monospace, Menlo, monospace;
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    opacity: 0.6;
  }
  .ctx-field input {
    flex: 1 1 auto;
    min-width: 0;
    max-width: 130px;
    padding: 4px 6px;
    background: rgba(0, 0, 0, 0.35);
    color: inherit;
    border: 1px solid rgba(200, 168, 106, 0.35);
    border-radius: 4px;
    font-size: 11px;
  }
  .ctx-form-row {
    display: flex;
    gap: 6px;
  }
  .ctx-item.apply {
    justify-content: center;
    border-color: rgba(200, 168, 106, 0.45);
  }
</style>

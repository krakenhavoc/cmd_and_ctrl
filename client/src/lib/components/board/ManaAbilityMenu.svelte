<script lang="ts">
  import ManaCost from "./ManaCost.svelte";
  // ManaAbilityMenu — small popover anchored to a battlefield permanent
  // that lists the card's activated mana abilities (S15). One button
  // per entry in card.mana_abilities; click fires the supplied
  // onActivate callback with the ability's index. Tap-cost greys the
  // button when the card is already tapped. Dismisses on Escape or
  // outside click (handled by the parent that mounts it).
  //
  // The menu is position-neutral — the parent wraps it in an
  // absolutely-positioned container anchored near the card. Keeping
  // layout concerns out of this component lets the battlefield
  // figure out overflow / flip-to-top-if-near-edge itself.

  import type { ActivatedAbilityView, CardView, GameView, ManaAbilityView } from "../../protocol";
  import {
    chargedManaCostLabel,
    chargedManaCostNote,
    energyCostSymbols,
    energyCostWords,
    exertCostWords,
    judgeAbilityRows,
    type AbilityRowContext,
    type MenuAction,
    type MenuItem,
  } from "../../contextMenu.logic";
  import { NO_LEGAL_ACTIONS, type LegalActions } from "../../legalActions";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    abilities: ManaAbilityView[];
    tapped: boolean;
    onActivate: (abilityIndex: number) => void;
    // S21 sub-PR 2: CR 602 activated abilities, listed below the
    // mana abilities in the same popover. Costs that need a further
    // choice (sacrifice, target) are collected by the parent after
    // the click.
    activated?: ActivatedAbilityView[];
    onActivateAbility?: (abilityIndex: number) => void;
    summoningSick?: boolean;
    // The WORDS for a shut activation window ("Not your turn", "Stack
    // isn't empty"), or "" when the panel's window reads open.
    // PlayerPanel computes it once per panel and threads it down as
    // `sorcerySpeedBlocked`.
    //
    // Words only, since ADR 0105 sub-PR 3. It used to be a VERDICT too:
    // any `sorcery_speed` row greyed while it was non-empty. That was a
    // client re-derivation of timing. It keyed on the printed clause,
    // so it greyed a Leonin Shikari's equip in combat, which the engine
    // accepts. The verdict is now the server's: the row's own
    // `timing_closed` (#1208) and the legal-action digest (`legalGate`
    // below). This string only explains it, as `timingReason` does in
    // contextMenu.logic.ts.
    timingReason?: string;
    // ADR 0105 (#1789): the card these rows belong to, and two lookups
    // for it. `legal` is what the board may HIGHLIGHT. It is the lookup
    // that knows nothing while highlights are off or autopass is
    // passing. A ready row gets the accent and sorts first. `legalGate`
    // is the frame's full lookup, which the setting never touches: a
    // sorcery-speed row whose ref the exact digest leaves out greys.
    // Both default to "no information", which greys nothing new.
    cardID?: string;
    legal?: LegalActions;
    legalGate?: LegalActions;
    // ADR 0105 sub-PR 4 (#1789): the card's CR 116.2 special actions
    // (foretell, suspend, plot, turn face up), built by Card with the
    // admin menu's own row builder (contextMenu.logic.ts
    // specialActionItems), so the ready accent, the order, the greying
    // and the payload are that function's. Listed first: a face-down
    // permanent has nothing else (CR 708.2a). A row fires
    // `onSpecialAction` with its action unaltered. Empty hides the
    // section.
    special?: MenuItem[];
    onSpecialAction?: (action: MenuAction) => void;
    // ADR 0117 §3: the Sandbox section, one row at the bottom: Tap or
    // Untap, whichever applies (CR 701.26a, 701.26b). It sends the raw
    // tap / untap, which adds no mana and activates nothing. Card sets
    // it on every permanent the viewer controls. On a mana source the
    // Tap row keeps #1438's label, "Tap (no mana)". Undefined hides it.
    onRawTap?: () => void;
    // ADR 0118 §2: the Sandbox section's "Cast anyway (don't pay)" row,
    // built by contextMenu.logic.ts castAnywayItem (greyed with its
    // reason when something other than mana would refuse the cast).
    // Choosing it fires `onCastAnyway`, which opens the dock's
    // confirmation; nothing is sent from here. On a hand card it is the
    // section's only row. Undefined hides it.
    castAnyway?: MenuItem;
    onCastAnyway?: () => void;
    // ADR 0131 §4: the "Pay life for {B}…" row on a hand card whose cost
    // has a symbol a grant (K'rrik) lets life pay. Choosing it fires
    // `onPayLife`, which starts the cast with the life stepper open;
    // nothing is sent from here. Undefined hides it.
    payLife?: MenuItem;
    onPayLife?: () => void;
    // ADR 0117 §2: the card the rows belong to, for its restrictions
    // (Arrest, Faith's Fetters) and, with `view` and `viewerID`, a
    // planeswalker's "already activated this turn" and the −N it cannot
    // pay. Absent (a popover mounted on its own) reads none of those.
    card?: CardView;
    view?: GameView | null;
    viewerID?: string | null;
    // ADR 0117 §3: an uncatalogued planeswalker's manual loyalty rows
    // (contextMenu.logic.ts manualLoyaltyRows), listed with the
    // activated rows. A row fires `onMenuAction` with its action.
    manualLoyalty?: MenuItem[];
    onMenuAction?: (action: MenuAction) => void;
    onClose?: () => void;
    // #1695: the paying player's current life — the card's
    // controller, same as the right-click context menu's rule. A
    // battlefield row only ever holds one seat's own permanents, so
    // PlayerPanel hands down that seat's life once per panel, the same
    // way it hands down `sorcerySpeedBlocked`. Undefined leaves an
    // unpayable life cost unblocked, like every other advisory check
    // here with nothing to judge against — the server's CR 119.4
    // refusal is still the real gate.
    payerLife?: number;
    // ADR 0106 §1 decision 6 (#1793): the rows are another player's
    // permanent's "Any player may activate this ability" rows, opened
    // by a viewer who does not control it. Their verdict is the
    // digest's (decision 5), so any row the exact digest leaves out
    // greys, not only a sorcery-speed one. No digest greys nothing new.
    across?: boolean;
  }

  const {
    abilities,
    tapped,
    onActivate,
    activated = [],
    onActivateAbility,
    summoningSick = false,
    timingReason = "",
    cardID = "",
    legal = NO_LEGAL_ACTIONS,
    legalGate = NO_LEGAL_ACTIONS,
    special = [],
    onSpecialAction,
    onRawTap,
    castAnyway,
    onCastAnyway,
    payLife,
    onPayLife,
    onClose,
    payerLife,
    across = false,
    card,
    view,
    viewerID,
    manualLoyalty = [],
    onMenuAction,
  }: Props = $props();

  function fireSpecial(item: MenuItem): void {
    if (item.disabled || !item.action) return;
    onSpecialAction?.(item.action);
    onClose?.();
  }

  function activate(index: number): void {
    onActivate(index);
    onClose?.();
  }

  // Whether each row can be used is ADR 0117 §2's one predicate
  // (contextMenu.logic.ts abilityRowBlocked), the same one the
  // left-click rule, the mana picker and the override menu ask, so a
  // click never disagrees with this menu. #1695 began it: this popover
  // used to keep a hand-maintained copy that never grew a life check.
  //
  // Timing is the server's answer (ADR 0105 sub-PR 3): the row's own
  // `timing_closed` (#1208), said in the panel's words, and the digest,
  // which greys a sorcery-speed row (or, `across`, any any-player row)
  // whose ref the exact digest leaves out. With no digest only the row
  // fields judge, so nothing is newly greyed. A ready row (ADR 0105
  // §2) takes the accent and sorts first, and a greyed row never does.
  const ctx = $derived<AbilityRowContext>({
    card,
    cardID,
    tapped,
    sick: summoningSick,
    payerLife,
    timingWords: timingReason,
    legalGate,
    across,
    loyalty: card ? { card, view, viewerID: viewerID ?? null } : undefined,
  });
  const manaRows = $derived(judgeAbilityRows(abilities, "mana", ctx, legal.readyManaRefs(cardID)));
  const activatedRows = $derived(
    judgeAbilityRows(activated, "activated", ctx, legal.readyAbilityRefs(cardID)),
  );

  function fireLoyalty(item: MenuItem): void {
    if (item.disabled || !item.action) return;
    onMenuAction?.(item.action);
    onClose?.();
  }

  // ADR 0117 §3: the Sandbox row's words. On a mana source the Tap row
  // keeps "Tap (no mana)" (#1438), so nobody mistakes it for the mana
  // row.
  const makesMana = $derived(abilities.length > 0);
  const sandboxLabel = $derived(tapped ? "Untap" : makesMana ? "Tap (no mana)" : "Tap");
  const sandboxTitle = $derived(
    tapped
      ? "Untap it by hand: a manual change that activates nothing"
      : "Turn it sideways by hand: a manual change that adds no mana and activates nothing",
  );

  function firePayLife(): void {
    onPayLife?.();
    onClose?.();
  }

  function fireCastAnyway(): void {
    if (!castAnyway || castAnyway.disabled) return;
    onCastAnyway?.();
    onClose?.();
  }

  // The Sandbox section's divider: always on a permanent's popover, as
  // before (ADR 0117 §3), and on a hand card's only when a row sits
  // above Cast anyway.
  const sandboxDivider = $derived(
    !!onRawTap ||
      special.length > 0 ||
      abilities.length > 0 ||
      activated.length > 0 ||
      manualLoyalty.length > 0,
  );

  function activateAbility(index: number): void {
    onActivateAbility?.(index);
    onClose?.();
  }

  function onKey(ev: KeyboardEvent): void {
    if (ev.key === "Escape") {
      ev.preventDefault();
      onClose?.();
    }
  }
</script>

<svelte:window onkeydown={onKey} />

<ModalLayer />

<div class="mana-menu" role="menu" aria-label="abilities">
  {#each special as item (item.id)}
    <button
      type="button"
      class="menu-item special"
      class:ready={item.ready}
      role="menuitem"
      disabled={item.disabled}
      title={item.hint || item.label}
      data-special={item.id}
      onclick={(ev) => {
        ev.stopPropagation();
        fireSpecial(item);
      }}
    >
      <span class="label">{item.label}</span>
      {#if item.ready}<span class="sr-only">, available</span>{/if}
      <span class="cost" aria-hidden="true">✦</span>
    </button>
  {/each}
  {#if special.length > 0 && (abilities.length > 0 || activated.length > 0 || manualLoyalty.length > 0)}
    <div class="divider" role="separator"></div>
  {/if}
  {#each manaRows as { a, blocked, ready } (a.index)}
    {@const costNote = chargedManaCostNote(a)}
    <button
      type="button"
      class="menu-item"
      class:ready
      role="menuitem"
      disabled={!!blocked}
      title={blocked || (a.produced ? `produces ${a.produced}` : a.label)}
      data-kind="mana"
      onclick={(ev) => {
        ev.stopPropagation();
        if (!blocked) activate(a.index);
      }}
    >
      <span class="label">{a.label || a.produced || "activate"}</span>
      {#if ready}<span class="sr-only">, available</span>{/if}
      {#if a.tap_cost}
        <span class="cost" aria-label="tap cost">↻</span>
      {/if}
      {#if a.mana_cost}
        <!-- #1190: the ability's OWN mana component (the Signet
             cycle's "{1}", Loot's exhaust "{G}"). Shows what the
             engine actually charges (charged_mana_cost); the tooltip
             names the printed cost only when a discount made the two
             differ. -->
        <span class="cost" title={costNote || `mana cost ${a.mana_cost}`}>
          <ManaCost cost={chargedManaCostLabel(a)} size={13} />
        </span>
      {/if}
      {#if a.sacrifice_cost || a.sacrifice_options}
        <span class="cost" aria-label="sacrifice cost">†</span>
      {/if}
      {#if a.life_cost}
        <!-- S22: a "Pay N life" cost component (Mana Confluence).
             Advisory — the server does the CR 119.4 check. The
             painlands' "deals 1 damage to you" is a RIDER, not a
             cost, so it shows up in the label instead of here. -->
        <span class="cost" aria-label={`pay ${a.life_cost} life`}>♥{a.life_cost}</span>
      {/if}
      {#if a.energy_cost}
        <!-- ADR 0129 §5: "Pay {E}" (Aether Hub). A seat short of energy
             has the row greyed by the server's cant_activate. -->
        <span class="cost" title={energyCostWords(a)}>
          <ManaCost cost={energyCostSymbols(a)} size={13} label={energyCostWords(a)} />
        </span>
      {/if}
      {#if a.exert}
        <!-- ADR 0130 §4: "Exert this land" (Arena of Glory). Never
             greyed: an exert can always be paid. -->
        <span class="cost exert" title={exertCostWords(a)} aria-label={exertCostWords(a)}
          >exert</span
        >
      {/if}
    </button>
  {/each}
  {#if activated.length > 0 || manualLoyalty.length > 0}
    {#if abilities.length > 0}
      <div class="divider" role="separator"></div>
    {/if}
    {#each activatedRows as { a, blocked, ready } (a.index)}
      {@const costNote = chargedManaCostNote(a)}
      <button
        type="button"
        class="menu-item"
        class:ready
        role="menuitem"
        disabled={!!blocked}
        title={blocked || a.label}
        data-kind="activated"
        onclick={(ev) => {
          ev.stopPropagation();
          if (!blocked) activateAbility(a.index);
        }}
      >
        <span class="label">{a.label || "activate"}</span>
        {#if ready}<span class="sr-only">, available</span>{/if}
        {#if a.tap_cost}
          <span class="cost" aria-label="tap cost">↻</span>
        {/if}
        {#if a.mana_cost}
          <!-- #1190: same chip as the mana list above — the row's
               Label text already prints the ability's cost baked in
               by hand, so this is the ONE place a discount that made
               the printed text stale is visible. -->
          <span class="cost" title={costNote || `mana cost ${a.mana_cost}`}>
            <ManaCost cost={chargedManaCostLabel(a)} size={13} />
          </span>
        {/if}
        {#if a.energy_cost || a.energy_cost_x}
          <!-- ADR 0129 §8: "Pay N {E}" / "Pay X {E}" as energy pips.
               Advisory; a seat short of energy has the row greyed by
               the server's cant_activate. -->
          {@const energy = energyCostSymbols(a)}
          <span class="cost" title={energyCostWords(a)}>
            <ManaCost cost={energy} size={13} label={energyCostWords(a)} />
          </span>
        {/if}
        {#if a.exert}
          <!-- ADR 0130 §4: "Exert this creature" (Steward of Solidarity). -->
          <span class="cost exert" title={exertCostWords(a)} aria-label={exertCostWords(a)}
            >exert</span
          >
        {/if}
      </button>
    {/each}
    {#each manualLoyalty as item (item.id)}
      <!-- ADR 0117 §3: a real activate_loyalty with a delta, whose text
           the players resolve. Greyed by canActivateLoyalty. -->
      <button
        type="button"
        class="menu-item"
        role="menuitem"
        disabled={item.disabled}
        title={item.hint || item.label}
        data-loyalty={item.id}
        onclick={(ev) => {
          ev.stopPropagation();
          fireLoyalty(item);
        }}
      >
        <span class="label">{item.label}</span>
      </button>
    {/each}
  {/if}
  {#if payLife && onPayLife}
    <!-- ADR 0131 §4: the name is a label contract (AGENTS.md §5). It is a
         payment choice, not a sandbox override, so it has its own group. -->
    <div class="sandbox" role="group" aria-label="payment">
      <span class="section-label" aria-hidden="true">Payment</span>
      <button
        type="button"
        class="menu-item"
        role="menuitem"
        title={payLife.hint || payLife.label}
        data-pay-life
        onclick={(ev) => {
          ev.stopPropagation();
          firePayLife();
        }}
      >
        <span class="label">{payLife.label}</span>
      </button>
    </div>
  {/if}
  {#if onRawTap || (castAnyway && onCastAnyway)}
    <!-- ADR 0117 §3: the Sandbox section. Never a pip, never counted by
         the click rule: it is on every permanent the viewer controls.
         ADR 0118 §2: and on every card the viewer could cast, with
         "Cast anyway (don't pay)". -->
    {#if sandboxDivider}
      <div class="divider" role="separator"></div>
    {/if}
    <div class="sandbox" role="group" aria-label="sandbox">
      <span class="section-label" aria-hidden="true">Sandbox</span>
      {#if castAnyway && onCastAnyway}
        <!-- ADR 0118 §2: the name is a label contract (AGENTS.md §5). It
             draws no pip and lights no ring. -->
        <button
          type="button"
          class="menu-item"
          role="menuitem"
          disabled={castAnyway.disabled}
          title={castAnyway.hint || castAnyway.label}
          data-cast-anyway
          onclick={(ev) => {
            ev.stopPropagation();
            fireCastAnyway();
          }}
        >
          <span class="label">{castAnyway.label}</span>
        </button>
      {/if}
      {#if onRawTap}
        <button
          type="button"
          class="menu-item"
          role="menuitem"
          title={sandboxTitle}
          data-raw-tap
          data-sandbox={tapped ? "untap" : "tap"}
          onclick={(ev) => {
            ev.stopPropagation();
            onRawTap?.();
            onClose?.();
          }}
        >
          <span class="label">{sandboxLabel}</span>
          <span class="cost" aria-hidden="true">↻</span>
        </button>
      {/if}
    </div>
  {/if}
</div>

<style>
  .divider {
    height: 1px;
    margin: 4px 2px;
    background: color-mix(in srgb, var(--accent) 35%, transparent);
  }

  .sandbox {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .section-label {
    padding: 0 8px;
    font-size: 9px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    opacity: 0.6;
  }

  .mana-menu {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 4px;
    background: rgba(12, 16, 30, 0.96);
    color: var(--accent);
    border: 1px solid color-mix(in srgb, var(--accent) 55%, transparent);
    border-radius: 6px;
    box-shadow: 0 10px 24px rgba(0, 0, 0, 0.6);
    min-width: 180px;
    font-size: 11px;
    z-index: 60;
  }
  .menu-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
    padding: 6px 8px;
    background: transparent;
    color: inherit;
    border: 1px solid transparent;
    border-radius: 4px;
    text-align: left;
    cursor: pointer;
    font-size: 11px;
    line-height: 1.2;
  }
  .menu-item:hover:not(:disabled),
  .menu-item:focus-visible:not(:disabled) {
    background: color-mix(in srgb, var(--accent) 15%, transparent);
    border-color: color-mix(in srgb, var(--accent) 40%, transparent);
    outline: none;
  }
  .menu-item:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  /* ADR 0105 (#1789): the server would accept this row right now. The
     one --ready colour, as a left rule plus a tint, so it reads in a
     list of gold rows without relying on hue alone. */
  .menu-item.ready {
    color: var(--ready);
    background: var(--ready-soft);
    box-shadow: inset 2px 0 0 var(--ready);
  }
  .menu-item.ready:hover,
  .menu-item.ready:focus-visible {
    background: var(--ready-soft);
    border-color: var(--ready);
  }
  .label {
    flex: 1;
  }
  .cost {
    font-weight: 700;
    opacity: 0.75;
  }
  /* ADR 0130 §4: the exert chip, a word rather than a symbol. */
  .cost.exert {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }
  /* ADR 0105 §7 (sub-PR 6): a ready row's accessible name gains
     "available". Spoken, not drawn: the accent already draws it. */
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
</style>

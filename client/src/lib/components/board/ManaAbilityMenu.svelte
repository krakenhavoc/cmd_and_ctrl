<script lang="ts">
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

  import type { ActivatedAbilityView, ManaAbilityView } from "../../protocol";
  import {
    abilityBlocked as sharedAbilityBlocked,
    chargedManaCostLabel,
    chargedManaCostNote,
    type AbilityCost,
    type MenuAction,
    type MenuItem,
  } from "../../contextMenu.logic";
  import {
    NO_LEGAL_ACTIONS,
    digestRefusesRow,
    readyFirst,
    type LegalActions,
  } from "../../legalActions";
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
    // #1438: a left-click on a mana source taps it FOR mana now, so
    // turning it sideways WITHOUT making mana lives here, labelled so
    // nobody mistakes it for the mana row. Undefined hides it.
    onRawTap?: () => void;
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
    onClose,
    payerLife,
  }: Props = $props();

  function fireSpecial(item: MenuItem): void {
    if (item.disabled || !item.action) return;
    onSpecialAction?.(item.action);
    onClose?.();
  }

  // The fallback sentence when the server shut a window the panel's
  // words call open: a per-player restriction, or a digest that
  // leaves the row out for a reason other than timing (the cost).
  const NOT_RIGHT_NOW = "Can't activate this right now";

  function activate(index: number): void {
    onActivate(index);
    onClose?.();
  }

  // An ability is unavailable when its tap cost can't be paid, its
  // life cost is more than the payer has, a sacrifice / return /
  // tap-other cost has nothing to pay it with, or any of the other
  // reasons the right-click context menu already judges through
  // `abilityBlocked` in contextMenu.logic.ts (#1695 — this used to be
  // a second, hand-maintained copy of that whole predicate, and the
  // copy never grew a life-cost check at all, which is what let an
  // unaffordable "Pay N life" ability stay clickable here after #1690
  // fixed the context menu's copy).
  //
  // Timing is the server's answer, in two parts (ADR 0105 sub-PR 3):
  //
  //   - `timing_closed`, the row's own verdict (#1208). Checked here
  //     rather than left to the shared predicate so that it covers
  //     loyalty rows too: the shared predicate reads those only with a
  //     LoyaltyContext, which this popover does not have. The words
  //     are the panel's.
  //   - the digest, after every row-field reason so that a more
  //     specific sentence wins. A sorcery-speed row the exact digest
  //     leaves out greys: the server would refuse it, whether for
  //     priority or for the cost. Only sorcery-speed rows, which are
  //     the rows the old client gate covered. An instant-speed row the
  //     digest misses stays live, so a gap in the enumerator leaves
  //     the row unmarked but never greys a move the engine accepts
  //     (ADR 0105, Consequences).
  //
  // With no digest (no decision owed, an older server, the capped
  // fallback) only the row fields judge, so nothing is newly greyed.
  function abilityBlocked(a: AbilityCost & { ref?: string }, activated: boolean): string {
    if (a.tap_cost && tapped) return "already tapped";
    if (a.tap_cost && summoningSick) return "summoning sickness";
    if (a.timing_closed) return timingReason || NOT_RIGHT_NOW;
    const fromRow = sharedAbilityBlocked(a, tapped, summoningSick, undefined, payerLife);
    if (fromRow) return fromRow;
    if (activated && a.sorcery_speed && digestRefusesRow(legalGate, cardID, a.ref)) {
      return NOT_RIGHT_NOW;
    }
    return "";
  }

  // ADR 0105 §2: a row the server would accept right now takes the
  // ready accent and sorts first. Legality is the digest's ref list
  // and nothing else. A row the row fields grey is never accented,
  // even if the two ever disagreed: an accent on a disabled button
  // would say two things at once.
  interface Row<T> {
    a: T;
    blocked: string;
    ready: boolean;
  }
  function rows<T extends AbilityCost & { ref?: string }>(
    list: readonly T[],
    refs: readonly string[],
    activated: boolean,
  ): Row<T>[] {
    const out = list.map((a) => {
      const blocked = abilityBlocked(a, activated);
      return { a, blocked, ready: !blocked && !!a.ref && refs.includes(a.ref) };
    });
    return readyFirst(out, (r) => r.ready);
  }
  const manaRows = $derived(rows(abilities, legal.readyManaRefs(cardID), false));
  const activatedRows = $derived(rows(activated, legal.readyAbilityRefs(cardID), true));

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
      <span class="cost" aria-hidden="true">✦</span>
    </button>
  {/each}
  {#if special.length > 0 && (abilities.length > 0 || activated.length > 0)}
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
      onclick={(ev) => {
        ev.stopPropagation();
        if (!blocked) activate(a.index);
      }}
    >
      <span class="label">{a.label || a.produced || "activate"}</span>
      {#if a.tap_cost}
        <span class="cost" aria-label="tap cost">↻</span>
      {/if}
      {#if a.mana_cost}
        <!-- #1190: the ability's OWN mana component (the Signet
             cycle's "{1}", Loot's exhaust "{G}"). Shows what the
             engine actually charges (charged_mana_cost); the tooltip
             names the printed cost only when a discount made the two
             differ. -->
        <span class="cost" aria-label="mana cost" title={costNote || `mana cost ${a.mana_cost}`}>
          {chargedManaCostLabel(a)}
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
    </button>
  {/each}
  {#if activated.length > 0}
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
        onclick={(ev) => {
          ev.stopPropagation();
          if (!blocked) activateAbility(a.index);
        }}
      >
        <span class="label">{a.label || "activate"}</span>
        {#if a.tap_cost}
          <span class="cost" aria-label="tap cost">↻</span>
        {/if}
        {#if a.mana_cost}
          <!-- #1190: same chip as the mana list above — the row's
               Label text already prints the ability's cost baked in
               by hand, so this is the ONE place a discount that made
               the printed text stale is visible. -->
          <span class="cost" aria-label="mana cost" title={costNote || `mana cost ${a.mana_cost}`}>
            {chargedManaCostLabel(a)}
          </span>
        {/if}
      </button>
    {/each}
  {/if}
  {#if onRawTap}
    <div class="divider" role="separator"></div>
    <button
      type="button"
      class="menu-item"
      role="menuitem"
      title="Turn it sideways without adding mana"
      data-raw-tap
      onclick={(ev) => {
        ev.stopPropagation();
        onRawTap?.();
        onClose?.();
      }}
    >
      <span class="label">Tap (no mana)</span>
      <span class="cost" aria-hidden="true">↻</span>
    </button>
  {/if}
</div>

<style>
  .divider {
    height: 1px;
    margin: 4px 2px;
    background: rgba(200, 168, 106, 0.35);
  }

  .mana-menu {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 4px;
    background: rgba(12, 16, 30, 0.96);
    color: var(--gold);
    border: 1px solid rgba(200, 168, 106, 0.55);
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
    background: rgba(200, 168, 106, 0.15);
    border-color: rgba(200, 168, 106, 0.4);
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
</style>

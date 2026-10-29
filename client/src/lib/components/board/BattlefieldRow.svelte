<script lang="ts">
  // BattlefieldRow renders a horizontal strip of battlefield cards
  // for one type bucket (CREATURES at the top, LANDS on the middle
  // left). The row owns layout (flex with wrap) and per-card visual
  // state derivation; the parent supplies the click handler so combat
  // / tap routing stays centralised.
  //
  // Cards are sorted by battle_x so a future row-reorder UX has a
  // stable handle. battle_y is intentionally ignored: typed rows make
  // a 2-D placement meaningless, and the server already defaults
  // both axes to 0 for cards that have never been positioned.

  import type { CantAttackChip } from "../../cantAttack";
  import type { CardView, GameView } from "../../protocol";
  import Card from "./Card.svelte";
  import { abilityPopover, popoverSurface } from "../../abilityPopover";
  import { settings } from "../../settings";
  import { etbPulse } from "../../animations";
  import { emit as tutorialEmit } from "../../tutorialBus";
  import { rowEntries, type RowEntry } from "../../tokenGroups";
  import {
    blockBoard,
    blockHighlight,
    fannedPiles,
    fanPile,
    fansOnClick,
    type BlockBoard,
  } from "../../blockPick";
  import { targeting, isLegalCardTarget, isPicked, type TargetingState } from "../../targeting";
  import {
    boardChoicePick,
    isBoardPickable,
    isBoardPicked,
    pickOnBoard,
    pileToggleID,
    type BoardChoicePick,
  } from "../../boardChoicePick";
  import { fitRow, MIN_CARD_H, type RowReport } from "../../rowFit";
  import { namePileKey, pileDepths, STRIP_DEPTH } from "../../landStrip";
  import {
    NO_COMBAT_RINGS,
    NO_LEGAL_ACTIONS,
    combatPipFor,
    readyPips,
    type CombatRings,
    type LegalActions,
    type ReadyPips,
  } from "../../legalActions";
  import type { MenuAction } from "../../contextMenu.logic";

  interface Props {
    label: string;
    cards: CardView[];
    viewerID: string | null;
    selectedCombatCardID?: string | null;
    // #3266: staged attackers, drawn with the attacking ring until the
    // dock's Attack with N sends them.
    stagedAttackers?: ReadonlySet<string>;
    onCardClick?: (card: CardView, ev: MouseEvent) => void;
    // onActivateManaAbility — fires `activate_mana_ability` for a
    // battlefield permanent the viewer controls. Routed down to
    // Card.svelte so the right-click / context menu can hit it.
    // Undefined suppresses the menu entirely (opponent panels).
    // ADR 0131 §2: `phyrexianLife` is the "Pay life for {B}…" row's claim,
    // sent as `phyrexian_life`; the third slot is the colours named up
    // front, which that row never names.
    onActivateManaAbility?: (
      card: CardView,
      abilityIndex: number,
      colors?: string[],
      phyrexianLife?: number,
    ) => void;
    // ADR 0117 §3: the popover's Sandbox row (Tap / Untap; "Tap (no
    // mana)" on a mana source, #1438). Wired on the viewer's own panel.
    onRawTap?: (card: CardView) => void;
    // The frame, for the popover's manual loyalty rows (ADR 0117 §3).
    view?: GameView | null;
    // ADR 0117 §1: true when a left-click on this card would do nothing
    // right now (PlayerPanel's click rule, after its intercepts). The
    // card then drops its pointer affordance. Undefined: every card
    // keeps it.
    clickInert?: (card: CardView) => boolean;
    // S21 sub-PR 2: CR 602 activated abilities, same menu.
    onActivateAbility?: (card: CardView, abilityIndex: number) => void;
    // S31: why the CR 307.1 sorcery-speed window is shut, or "" when
    // it is open. Pass-through to Card → ManaAbilityMenu, which greys
    // `sorcery_speed` abilities with it.
    sorcerySpeedBlocked?: string;
    // #1695: the paying player's current life — always this row's own
    // seat, since a BattlefieldRow only ever holds one seat's
    // controlled cards. PlayerPanel hands it down once per panel, the
    // same way it hands down sorcerySpeedBlocked, so the mana
    // popover's life-cost check has something to grey against.
    payerLife?: number;
    // compact — one card size down (PlayerPanel's --card-w-sm). Used
    // for the middle band: non-creature permanents and lands.
    compact?: boolean;
    // strip — piles instead of a wrapping grid. Lands: untapped
    // copies of the same name stack into one pile with a count
    // badge (four Forests take the width of one), tapped lands sort
    // to the right of every pile so the untapped count reads at a
    // glance, and hovering a pile spreads it so each copy can be
    // clicked on its own.
    strip?: boolean;
    // #4087: never overlap the piles (planeswalkers and battles). A row
    // that does not fit shrinks its cards, then wraps, then scrolls;
    // the cards never sit on one another.
    spread?: boolean;
    // #3028 / #4087: the share of the back row this zone may take, 0..1
    // (backRow.ts). Unset: the whole row.
    share?: number;
    // S24: the Equipment and Auras attached to each host, keyed by
    // the host's instance ID. Derived by PlayerPanel from the
    // one-directional `attached_to` on the wire. They are drawn
    // inside the host's own listitem, offset behind it — the same
    // negative-margin overlap the land strip uses — so a sword reads
    // as being ON the creature rather than as a separate permanent.
    attachmentsByHost?: Record<string, CardView[]>;
    // S24: the name of the player each Curse-style permanent enchants,
    // keyed by the permanent's instance ID. A card attached to a
    // PLAYER has no host to be drawn behind, so the relation would be
    // invisible without this. Derived by PlayerPanel, which is the
    // component that can see the seat list.
    curseTargets?: Record<string, string>;
    // ADR 0104: the OWNER's name for each permanent a different player
    // controls — a stolen permanent, or the permanent a stolen spell
    // became. Derived by PlayerPanel, for curseTargets' reason.
    takenFrom?: Record<string, string>;
    // ADR 0114 owner decision 1: the controller's name for each
    // Ring-bearer, for its marker's title. Derived by PlayerPanel, for
    // curseTargets' reason.
    ringBearers?: Record<string, string>;
    // ADR 0106 §2 (#1794): the CAN'T ATTACK chip for each creature that
    // can't attack its owner, keyed by instance ID. Derived by
    // PlayerPanel from the card views, for curseTargets' reason.
    cantAttack?: Record<string, CantAttackChip>;
    // #1724: identical tokens fold into a group drawn as at most two
    // cards (untapped, tapped) with a count; clicking one calls this
    // with the group's key and the panel opens the member list
    // (TokenGroupModal). Undefined turns grouping off, so a row with
    // nowhere to open the list never hides a token behind a count.
    onGroupClick?: (groupKey: string) => void;
    // #3028: how many piles the row draws, read back by PlayerPanel to
    // share the back row's width between the lands and the rest.
    pileCount?: number;
    // #3044: the height the panel gives this row (allocateRowHeights),
    // which a crowded row may wrap into rather than shrink; 0 or unset
    // keeps it on one line. `demand` reads back what the row asks for.
    budget?: number;
    demand?: RowReport | null;
    // #3893: an empty row keeps one card's line, so the rows after it
    // sit where they would beside a card (the creature row). Without
    // it an empty row is one label high.
    hold?: boolean;
    // ADR 0105 (#1789): the frame's legal-action lookups. `legal` is
    // what may be highlighted ("nothing" while highlights are off or
    // autopass is passing; on an opponent's panel only that panel's
    // any-player rows, ADR 0106 §1, legalActions.ts acrossActions). It
    // draws each permanent's bolt / drop / star pips, and the ring that
    // goes with a bolt or a star (a drop pip alone gets no ring, §2).
    // `legalGate` is the full lookup, for the ability popover's gates
    // only. Both default to "no information".
    legal?: LegalActions;
    legalGate?: LegalActions;
    // ADR 0105 sub-PR 4: sends a CR 116.2 special action chosen from a
    // permanent's ability popover: a face-down permanent's turn face
    // up, a Room's unlock. Wired on the viewer's own panel only.
    // Undefined keeps those rows, and their star pip, out of the
    // popover.
    onSpecialAction?: (action: MenuAction) => void;
    // ADR 0105 sub-PR 5: the combat rings (legalActions.ts
    // combatRings), already empty while highlights are off. A
    // candidate (a creature that may attack or block) wears the ready
    // ring. A target of the selected creature (a planeswalker or
    // battle it may attack, an attacker it may block) wears it too,
    // drawn so it reads over the red attacking ring.
    combat?: CombatRings;
  }

  let {
    label,
    cards,
    viewerID,
    selectedCombatCardID = null,
    stagedAttackers,
    onCardClick,
    onActivateManaAbility,
    onRawTap,
    view,
    clickInert,
    onActivateAbility,
    sorcerySpeedBlocked = "",
    payerLife,
    compact = false,
    strip = false,
    spread = false,
    share = 1,
    attachmentsByHost = {},
    curseTargets = {},
    takenFrom = {},
    ringBearers = {},
    cantAttack = {},
    onGroupClick,
    legal = NO_LEGAL_ACTIONS,
    legalGate = NO_LEGAL_ACTIONS,
    onSpecialAction,
    combat = NO_COMBAT_RINGS,
    pileCount = $bindable(0),
    budget = 0,
    demand = $bindable(null),
    hold = false,
  }: Props = $props();

  const held = $derived(hold && cards.length === 0);

  // #4087: the least width a spread row asks of the back row: every card
  // at the smallest size a row shrinks to (rowFit MIN_CARD_H, 5:7), a
  // gap between them, and the row's padding. A crowded back row shrinks
  // the other zones before it takes that from walkers, which never
  // overlap and so cannot give up more. An inline min-width, because the
  // panel's flipped layout sets its own on every zone.
  const spreadMin = (n: number): number =>
    Math.ceil((MIN_CARD_H * 5) / 7) * n + 12 * Math.max(0, n - 1) + 12;

  // ADR 0120 §3: the surface this row is drawn on, table or expanded.
  const surface = popoverSurface();

  // ADR 0105 §2: the special-action rows go only to a card that offers
  // any. The server strips `special_actions` from every seat but the
  // card's controller, so an opponent's card has none anyway.
  const specialFor = (c: CardView) =>
    onSpecialAction && (c.special_actions?.length ?? 0) > 0 ? onSpecialAction : undefined;
  // The ready ring on a permanent: a live activated ability (the
  // bolt), or a live special action (the star: a face-down permanent's
  // turn face up). A drop pip alone gets no ring (§2). Each only where
  // its popover is wired, as the pips are.
  const ringFor = (c: CardView, p: ReadyPips) =>
    (!!onActivateAbility && p.abilities > 0) || (!!specialFor(c) && p.special > 0);
  // A token group's card stands for every member, so it lights when
  // any member does, the way its selection ring does.
  const anyIn = (ids: ReadonlySet<string>, c: CardView, memberIDs: string[] | undefined) =>
    ids.size > 0 && (memberIDs ?? [c.instance_id]).some((id) => ids.has(id));

  const sorted = $derived.by(() => {
    const byX = [...cards].sort((a, b) => (a.battle_x ?? 0) - (b.battle_x ?? 0));
    if (!strip) return byX;
    // Stable partition: untapped first, tapped after, order kept within each.
    return [...byX.filter((c) => !c.tapped), ...byX.filter((c) => c.tapped)];
  });

  // Piles (strip only). One pile per NAME among the untapped lands,
  // in first-appearance order, then (#3044) one per name among the
  // tapped lands (landStrip.ts namePileKey). A card that carries an
  // attachment (an enchanted land) or is a commander keeps a pile of
  // its own: stacking it under identical basics would hide the very
  // thing that makes it different. The non-strip rows render `sorted`
  // as one pile each so the markup has a single shape.
  //
  // #1724: before any of that, identical tokens fold into groups
  // (lib/tokenGroups.ts), and so do identical nontoken permanents
  // outside the strip (#3046). A group half is always a pile of its own —
  // one card, with the group's count — and its members never reach
  // the name piles.
  interface Pile {
    key: string;
    tapped: boolean;
    cards: CardView[];
    // Set for a token group's half: every member it stands for, the
    // drawn one first.
    group?: { groupKey: string; members: CardView[] };
  }
  // #3046: what the board lights on one card, so a pile of printed
  // cards draws one card per distinct highlight (tokenGroups.ts
  // GroupOptions). Every ring and pip a card can wear on its own is in
  // it: the ready pips, the combat rings and pip, a targeting or
  // choice ring, and a combat selection.
  function highlightOf(
    c: CardView,
    t: TargetingState | null,
    pick: BoardChoicePick | null,
    blocks: BlockBoard | null,
  ): string {
    const id = c.instance_id;
    const p = readyPips(legal, c, "battlefield");
    return [
      p.abilities,
      p.mana ? 1 : 0,
      p.special,
      combat.candidates.has(id) ? "c" : "",
      combat.targets.has(id) ? "t" : "",
      JSON.stringify(combatPipFor(legal, combat, [id]) ?? null),
      t && isLegalCardTarget(t, id) ? "L" : "",
      t && isPicked(t, id) ? "P" : "",
      isBoardPickable(pick, id) ? "b" : "",
      isBoardPicked(pick, id) ? "B" : "",
      selectedCombatCardID === id ? "s" : "",
      stagedAttackers?.has(id) ? "a" : "",
      blockHighlight(blocks, id),
    ].join("|");
  }
  const entries = $derived.by((): RowEntry[] => {
    // #4087: a spread row never folds identical permanents into a pile
    // either; two copies of one walker are two loyalty totals.
    if (!onGroupClick || spread) {
      return sorted.map((c) => ({ kind: "card" as const, key: c.instance_id, card: c }));
    }
    const t = $targeting;
    const pick = $boardChoicePick;
    const blocks = $blockBoard;
    const out = rowEntries(sorted, attachmentsByHost, {
      nontoken: !strip,
      highlight: (c) => highlightOf(c, t, pick, blocks),
    });
    // #3898: while blockers are declared, a pile the player clicked is
    // drawn as its cards, so the copy itself is what they click next.
    const fanned = $fannedPiles;
    if (!blocks || fanned.size === 0) return out;
    return out.flatMap((e): RowEntry[] =>
      e.kind === "group" && fanned.has(e.groupKey)
        ? e.members.map((m) => ({ kind: "card", key: m.instance_id, card: m }))
        : [e],
    );
  });
  // A pile's click: fanned out while blockers are declared (#3898), its
  // member list otherwise.
  function groupClick(groupKey: string, members: readonly CardView[]): void {
    // #4028: while a choice is picked on the board, a click on a pile of
    // identical tokens picks (or puts back) one of them, as a click on a
    // single permanent does.
    const pickID = pileToggleID(
      $boardChoicePick,
      members.map((m) => m.instance_id),
    );
    if (pickID !== null) {
      pickOnBoard(pickID);
      return;
    }
    if (fansOnClick($blockBoard, members)) {
      fannedPiles.update((s) => fanPile(s, groupKey));
      return;
    }
    onGroupClick?.(groupKey);
  }
  const piles = $derived.by((): Pile[] => {
    const out: Pile[] = [];
    const byName = new Map<string, Pile>();
    // ADR 0117 §1: "Use this one" in a group's member list can open
    // the ability popover on a member that is not the drawn one. The
    // group then draws THAT member while its popover is open, so the
    // popover opens at the group's card and its rows act on the member
    // the player chose. Only on the surface the popover was opened
    // from (ADR 0120 §3): the other copy of the row keeps its own rep.
    const popoverID = $abilityPopover?.surface === surface ? $abilityPopover.cardID : undefined;
    for (const e of entries) {
      if (e.kind === "group") {
        const drawn = e.members.find((m) => m.instance_id === popoverID) ?? e.rep;
        out.push({
          key: e.key,
          tapped: e.tapped,
          cards: [drawn],
          group: {
            groupKey: e.groupKey,
            members: [drawn, ...e.members.filter((m) => m !== drawn)],
          },
        });
        continue;
      }
      const c = e.card;
      const key = strip
        ? namePileKey(c, (attachmentsByHost[c.instance_id] ?? []).length > 0)
        : null;
      if (!key) {
        out.push({ key: c.instance_id, tapped: !!c.tapped, cards: [c] });
        continue;
      }
      const hit = byName.get(key);
      if (hit) {
        hit.cards.push(c);
        continue;
      }
      const p: Pile = { key, tapped: !!c.tapped, cards: [c] };
      byName.set(key, p);
      out.push(p);
    }
    return out;
  });
  // #3044: a tapped pile still counts each land in it, as each was a
  // pile of its own before tapped lands piled. A turned land is wide,
  // and the back row's share by pile count (#3028) would otherwise give
  // the strip less room for piling them.
  $effect(() => {
    pileCount = piles.reduce((n, p) => n + (strip && p.tapped && !p.group ? p.cards.length : 1), 0);
  });

  // The attachments drawn behind a pile's card: its own, or for a
  // group every member's, so an Equipment on a grouped token is still
  // on the board to be clicked.
  function attachmentsFor(p: Pile, c: CardView): CardView[] {
    if (!p.group) return attachmentsByHost[c.instance_id] ?? [];
    return p.group.members.flatMap((m) => attachmentsByHost[m.instance_id] ?? []);
  }
</script>

<div
  class="row"
  class:compact
  class:strip
  class:spread
  class:held
  data-zone={label}
  style:--row-share={share}
  style:min-width={spread ? `${spreadMin(cards.length)}px` : undefined}
>
  <!-- #2995: the dice stay off the label (DiceLayer, data-dice-avoid). -->
  <span class="row-label" aria-hidden="true" data-dice-avoid>
    <span class="row-name">{label}</span>
    {#if cards.length > 0}<span class="row-count">{cards.length}</span>{/if}
  </span>
  <!-- #2438: the row's own card size, which fitRow shrinks the list's
       cards from. It sits outside the list, so the list's override never
       reaches it. The outer box fills the row and clips, so a probe
       taller or wider than the row (an empty row is one label high) is
       never scroll overflow. -->
  <span class="fit-probe" aria-hidden="true"><span class="fit-probe-card"></span></span>
  <div
    class="row-cards"
    role="list"
    aria-label={label}
    use:fitRow={{ strip, spread, budget, hold: held, onDemand: (d) => (demand = d) }}
  >
    {#each piles as p, pi (p.key)}
      {@const depths = pileDepths(p.cards.length)}
      <!-- ADR 0076 §2.5: a pointer-only hover signal for the tutorial; not a control. -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        class="pile"
        class:tapped={p.tapped}
        class:multi={p.cards.length > 1}
        class:deep={p.cards.length > STRIP_DEPTH}
        class:group={!!p.group}
        style:--pile-n={p.cards.length}
        style:--pi={pi}
        onpointerenter={p.cards.length > 1 ? () => tutorialEmit("pile-hovered") : undefined}
        title={p.group
          ? `${p.group.members.length} × ${p.cards[0].name} — click to choose which`
          : p.cards.length > 1
            ? `${p.cards.length} × ${p.cards[0].name}`
            : undefined}
      >
        {#each p.cards as c, i (c.instance_id)}
          {@const attached = attachmentsFor(p, c)}
          {@const memberIDs = p.group?.members.map((m) => m.instance_id)}
          {@const cPips = readyPips(legal, c, "battlefield")}
          {@const cTarget = anyIn(combat.targets, c, memberIDs)}
          <div
            role="listitem"
            class:tapped={!!c.tapped}
            class:buried={i > 0 && depths[i] === depths[i - 1]}
            style:--i={depths[i]}
            style:--k={i}
            use:etbPulse
          >
            <div class="host-stack" class:has-attachments={attached.length > 0}>
              {#each attached as a (a.instance_id)}
                {@const aPips = readyPips(legal, a, "battlefield")}
                <div class="attachment">
                  <Card
                    artOnly={$settings.display.battlefieldArt}
                    card={a}
                    ready={ringFor(a, aPips)}
                    pips={aPips}
                    onSpecialAction={specialFor(a)}
                    {legal}
                    {legalGate}
                    enchantedPlayer={curseTargets[a.instance_id]}
                    takenFrom={takenFrom[a.instance_id]}
                    ringBearerOf={ringBearers[a.instance_id]}
                    cantAttack={cantAttack[a.instance_id]}
                    onClick={onCardClick}
                    onActivateManaAbility={onActivateManaAbility
                      ? (idx, life) => onActivateManaAbility(a, idx, undefined, life)
                      : undefined}
                    onRawTap={onRawTap ? () => onRawTap(a) : undefined}
                    onMenuAction={onSpecialAction}
                    {view}
                    inert={clickInert?.(a) ?? false}
                    onActivateAbility={onActivateAbility
                      ? (idx) => onActivateAbility(a, idx)
                      : undefined}
                    {sorcerySpeedBlocked}
                    {payerLife}
                    {viewerID}
                  />
                </div>
              {/each}
              <div class="host">
                {#if p.group && p.group.members.length > 1}
                  <!-- #4431: paint-only ghost copies behind the group's card, so a
                       stack reads as a stack. Absolutely placed inside the host's
                       own box: they take no layout width, so fitRow never re-fits. -->
                  <span class="fan" class:tapped={!!c.tapped} aria-hidden="true">
                    <span class="ghost g2"></span>
                    {#if p.group.members.length > 2}<span class="ghost g3"></span>{/if}
                  </span>
                {/if}
                <Card
                  artOnly={$settings.display.battlefieldArt}
                  card={c}
                  ready={ringFor(c, cPips) || cTarget || anyIn(combat.candidates, c, memberIDs)}
                  combatTarget={cTarget}
                  combatPip={combatPipFor(legal, combat, memberIDs ?? [c.instance_id])}
                  pips={cPips}
                  onSpecialAction={specialFor(c)}
                  {legal}
                  {legalGate}
                  enchantedPlayer={curseTargets[c.instance_id]}
                  takenFrom={takenFrom[c.instance_id]}
                  ringBearerOf={ringBearers[c.instance_id]}
                  cantAttack={cantAttack[c.instance_id]}
                  selected={memberIDs
                    ? !!selectedCombatCardID && memberIDs.includes(selectedCombatCardID)
                    : selectedCombatCardID === c.instance_id}
                  attacking={!!c.attacking_target ||
                    (memberIDs ?? [c.instance_id]).some((id) => stagedAttackers?.has(id))}
                  blocking={!!c.blocking_target}
                  {memberIDs}
                  onClick={p.group && onGroupClick
                    ? () => groupClick(p.group!.groupKey, p.group!.members)
                    : onCardClick}
                  onActivateManaAbility={onActivateManaAbility
                    ? (idx, life) => onActivateManaAbility(c, idx, undefined, life)
                    : undefined}
                  onRawTap={onRawTap ? () => onRawTap(c) : undefined}
                  onMenuAction={onSpecialAction}
                  {view}
                  inert={!p.group && (clickInert?.(c) ?? false)}
                  onActivateAbility={onActivateAbility
                    ? (idx) => onActivateAbility(c, idx)
                    : undefined}
                  {sorcerySpeedBlocked}
                  {payerLife}
                  {viewerID}
                />
                {#if p.group}
                  <!-- Every other member keeps an element with its
                       instance ID where the group is drawn, so an arrow,
                       a picker or an animation that looks a card up by
                       [data-instance-id] lands on the group's card
                       rather than on nothing. -->
                  {#each p.group.members.slice(1) as m (m.instance_id)}
                    <span class="member-anchor" data-instance-id={m.instance_id} aria-hidden="true"
                    ></span>
                  {/each}
                {/if}
              </div>
            </div>
          </div>
        {/each}
        {#if p.group}
          <span class="group-count" data-testid="group-count">×{p.group.members.length}</span>
        {:else if p.cards.length > 1}
          <span class="pile-count" aria-hidden="true">{p.cards.length}</span>
        {/if}
      </div>
    {/each}
  </div>
</div>

<style>
  .row {
    position: relative;
    padding: 18px 6px 6px;
    min-height: 0;
    min-width: 0;
    overflow: auto;
  }
  /* #3893: an empty creature row is as tall as a row of one card (the
     padding adds the label's band), so the back row behind it does not
     move up into its place. It still shrinks on a short panel, as a
     row with a card does; fitRow asks the panel for the same height. */
  .row.held {
    height: var(--card-h, 168px);
  }
  .row.compact {
    --card-w: var(--card-w-sm, 88px);
    --card-h: var(--card-h-sm, 123px);
  }
  .fit-probe {
    position: absolute;
    inset: 0;
    overflow: hidden;
    visibility: hidden;
    pointer-events: none;
  }
  .fit-probe-card {
    position: absolute;
    top: 0;
    left: 0;
    width: var(--card-w, 88px);
    height: var(--card-h, 123px);
  }
  .row-label {
    position: absolute;
    /* #4090: the label's 16px line fills the row's 18px top padding, so
       nothing in the row (a card, a pile badge) reaches into it. */
    top: 1px;
    left: 6px;
    display: inline-flex;
    align-items: baseline;
    gap: 6px;
    font-family: var(--font-ui);
    font-size: 11px;
    letter-spacing: 0.01em;
    color: var(--fg-dim);
    pointer-events: none;
    font-weight: 600;
    white-space: nowrap;
    /* Never wider than the row: an absolute label past the row's edge
       is scroll overflow, which gave a narrow row a scrollbar and could
       scroll its own label out of view. A long name ends in "…". */
    max-width: calc(100% - 12px);
  }
  /* Sentence case from the lower-case zone names ("creatures",
     "enchant / artifact"), which stay the rows' accessible names. */
  .row-name {
    display: inline-block;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .row-name::first-letter {
    text-transform: uppercase;
  }
  .row-count {
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    color: var(--fg-muted);
    background: var(--overlay);
    border-radius: 999px;
    padding: 0 6px;
    line-height: 16px;
  }
  /* A host and everything attached to it. The attachments are laid
     out first and overlapped leftwards behind the host, which is
     drawn last so it sits on top; the wrapper claims only the host's
     width, so a creature carrying two swords does not reflow the row
     out from under its neighbours. */
  .host-stack {
    display: flex;
    flex-direction: row;
    align-items: flex-start;
  }
  .host-stack.has-attachments {
    margin-left: calc(var(--card-w, 88px) * 0.34);
  }
  .host-stack .attachment {
    flex: 0 0 auto;
    margin-right: calc(var(--card-w, 88px) * -0.72);
    transform: translateY(8px);
    filter: brightness(0.88);
  }
  .host-stack .attachment:hover {
    z-index: 7;
    filter: none;
  }
  /* An attachment's failed-art pip (#33) goes in its bottom-left
     corner, tapped or not. The pip cannot paint over the tile after
     it — each tile is its own stacking context — so it has to sit
     where that tile leaves the attachment uncovered, whatever is
     tapped. Upright next to an upright tile, that is the left 28% of
     the attachment. A tapped host turns to cover a band across the
     middle of its height, leaving the bottom of the attachment (the
     8px drop helps) — Card's 22px, the old spot, went under it. A
     tapped attachment turns its bottom edge to the left, clear of an
     upright host, and its bottom-left corner to the top-left, left of a
     turned host. 2px up keeps it clear of the band at the smallest
     size, and the left edge keeps it inside the sliver. It covers the
     start of an AUTO badge or keyword row there, which on an Aura or
     Equipment is rare. boardArtPip.test.ts checks every combination. */
  .host-stack .attachment :global(.card) {
    --art-error-top: calc(100% - 16px);
    --art-error-left: 0px;
  }
  /* Rows grow from the middle of the table: cards centre in the row
     rather than piling up in the top-left corner, so a two-creature
     board reads as a board and not as a corner. */
  .row-cards {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: center;
    gap: 8px;
    align-content: flex-start;
    align-items: flex-start;
    height: 100%;
  }
  /* #2336: a row stays on one line. Cards that do not fit first shrink
     (#2438: fitRow sets --card-w / --card-h on the list, down to a 56px
     card), then overlap, each by --fit-overlap (lib/rowFit.ts),
     instead of wrapping onto a line the panel has no room for. A
     hovered card comes to the front. Past the overlap cap a huge board
     scrolls sideways, and `safe` keeps its first card reachable. */
  .row:not(.strip) .row-cards {
    flex-wrap: nowrap;
    justify-content: safe center;
  }
  .row:not(.strip) .pile + .pile {
    margin-left: calc(-1 * var(--fit-overlap, 0px));
  }
  .row:not(.strip) .pile:hover {
    z-index: 6;
  }
  /* #3044: except a row fitRow wraps, because one line would shrink its
     cards a long way and the panel holds the height for more lines of
     larger ones (lib/rowFit.ts fitWrap). It has no overlap. An
     across-table opponent's list is turned 180° (PlayerPanel), so its
     first line sits against the table's centre, as yours does. */
  .row:not(.strip) .row-cards:global([data-wrap]) {
    flex-wrap: wrap;
    align-content: flex-start;
  }
  /* #2442: a tapped card turns 90 degrees about its centre, so it
     overhangs its box by (h - w) / 2 on each side while its box keeps
     the upright width, and the turned card painted over its
     neighbours. The room is padding, not margin, so it is part of the
     pile's width: fitRow measures it, scales it with the card and
     counts it in the overlap (lib/rowFit.ts tappedRoom). It eases in
     over the turn (animations.ts TAP_DURATION). */
  .row:not(.strip) .pile {
    transition: padding-inline 180ms var(--ease);
  }
  .row:not(.strip) .pile.tapped {
    padding-inline: calc((var(--card-h, 123px) - var(--card-w, 88px)) / 2);
  }
  @media (prefers-reduced-motion: reduce) {
    .row:not(.strip) .pile {
      transition: none;
    }
  }
  /* Outside the strip a pile is one card and nothing more; the
     wrapper exists so the markup has one shape. */
  .pile {
    position: relative;
    display: flex;
    flex-direction: row;
    align-items: flex-start;
    flex: 0 0 auto;
  }
  /* Land strip: no wrap. Between piles each overlaps the previous so
     the name band stays readable; tapped piles (rotated by
     Card.svelte) are sorted to the end and given room for their
     rotated width. The margin rules below are read by
     boardArtPip.test.ts, which resolves their cascade by hand — keep
     each on its own `.row.strip .pile…` selector. */
  .row.strip {
    /* Room for the pile count badges, which sit 6px above the tiles,
       under the row label. */
    padding-top: 26px;
  }
  .row.strip .row-cards {
    flex-wrap: nowrap;
    justify-content: flex-start;
    gap: 0;
    padding-right: calc(var(--card-h, 123px) * 0.2);
  }
  .row.strip .pile {
    margin-left: calc(
      var(--card-w, 88px) * -0.6 * var(--strip-k, 1) + 6px * (1 - var(--strip-k, 1))
    );
  }
  .row.strip .pile:first-child {
    margin-left: 0;
  }
  .row.strip .pile.tapped {
    margin-left: calc(var(--card-w, 88px) * -0.35 * var(--strip-k, 1));
  }
  .row.strip .pile:not(.tapped) + .pile.tapped {
    margin-left: calc(var(--card-h, 123px) * 0.2);
  }
  /* Every land tapped: the first is tapped too. Without this rule the
     .tapped overlap above, as specific as :first-child and later,
     won, and pushed the first land a third of its turned width out of
     the row's clip, its failed-art pip (#33) with it. The turned tile
     overhangs its box by (h - w) / 2 on each side; that is its room. */
  .row.strip .pile.tapped:first-child {
    margin-left: calc((var(--card-h, 123px) - var(--card-w, 88px)) / 2);
  }
  .row.strip .pile:hover {
    z-index: 6;
  }
  /* Inside a pile every copy after the first sits 4px right and 3px
     up of the one before, so the pile reads as a stack and takes the
     width of one card plus a sliver per copy. The top copy is the
     last in the DOM and paints over the rest, which is where the
     count badge and the pip live. */
  .row.strip .pile > [role="listitem"] {
    position: relative;
    top: calc(var(--i, 0) * -3px);
  }
  /* #4422: hovering a pile spreads its copies apart, but only as paint.
     The spread used to be a hover margin, which widened the pile, so
     fitRow (which measures the strip's width) read a wider strip while a
     pile was hovered, tightened every pile's overlap (--strip-k) to
     compensate, and the piles slid left out from under the pointer; the
     hover dropped, the strip measured narrower, the piles slid back
     right, and the pile flickered left and right for as long as the
     pointer rested there. A translate moves no layout box: the pile, its
     neighbours and the strip measure the same hovered or not. Copy k
     rests at its depth's 4px step and moves to k spread steps. The pile
     stays hovered because the copies are its descendants, and it
     paints over its neighbours (z-index above). The spread eases open
     on hover only. #3044: an always-on transition animated every
     copy's move whenever fitRow resized the cards. */
  .row.strip .pile.multi:hover > [role="listitem"] {
    transform: translateX(calc(var(--k, 0) * var(--spread-step) - var(--i, 0) * 4px));
    transition: transform 140ms var(--ease);
  }
  .row.strip .pile > [role="listitem"] + [role="listitem"] {
    margin-left: calc(var(--card-w, 88px) * -1 + 4px);
  }
  /* #3044: a pile shows at most STRIP_DEPTH steps (lib/landStrip.ts), so
     28 Forests read as one card and a count, not a fan of slivers. The
     copies past that sit exactly under the bottom one. */
  .row.strip .pile > [role="listitem"] + [role="listitem"].buried {
    margin-left: calc(var(--card-w, 88px) * -1);
  }
  /* The spread step: 40% of each card, the same the strip already leaves
     visible between names, so each copy can be told apart, hovered and
     clicked (a specific Forest for a fetch, say). */
  .row.strip .pile.multi:hover {
    --spread-step: calc(var(--card-w, 88px) * 0.4);
  }
  /* #3044: tapped lands pile by name too. A turned copy is a height
     wide, so a spread tapped pile steps by a share of the height, which
     leaves each copy's pip (#33) clear of the next. */
  .row.strip .pile.tapped.multi:hover {
    --spread-step: calc(var(--card-h, 123px) * 0.45);
  }
  /* #3044: a deep pile spreads no wider than four more cards, so a
     hover over 28 Forests stays near the pile; each copy still shows a
     sliver to hover (it zooms) and click. */
  .row.strip .pile.deep:hover {
    --spread-step: min(var(--card-w, 88px) * 0.4, var(--card-w, 88px) * 4 / (var(--pile-n, 2) - 1));
  }
  .row.strip .pile.tapped.deep:hover {
    --spread-step: min(
      var(--card-h, 123px) * 0.45,
      var(--card-w, 88px) * 4 / (var(--pile-n, 2) - 1)
    );
  }
  /* #1724: a token group's card. The host box is what the hidden
     members' anchors cover, so a lookup by any member's instance ID
     measures the card that stands for it. */
  .host {
    position: relative;
    /* #4431: the fan's ghosts sit at z-index -1, behind the card but
       inside this box's stacking context. */
    isolation: isolate;
  }
  .fan,
  .ghost {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }
  /* The card's own face, a shade darker per layer, with an edge lifted
     enough to read against the board. Each ghost steps 6px further up and
     left (the row's padding holds it); a tapped fan turns with the card,
     so the steps stay up and left, inside the tapped room above it. */
  .ghost {
    z-index: -1;
    border-radius: 8px;
    box-sizing: border-box;
    background: #0a0f1c;
    border: 1px solid color-mix(in srgb, var(--fg-dim, #5d7187) 70%, #0a0e1a);
    box-shadow: 0 1px 3px rgb(0 0 0 / 0.5);
    transform: translate(var(--fx), var(--fy)) rotate(var(--fan-tap, 0deg));
    transition: transform 180ms var(--ease);
  }
  .ghost.g2 {
    --fx: -4px;
    --fy: -6px;
  }
  .ghost.g3 {
    --fx: -8px;
    --fy: -12px;
    z-index: -2;
    background: #080c17;
  }
  .fan.tapped {
    --fan-tap: 90deg;
  }
  @media (prefers-reduced-motion: reduce) {
    .ghost,
    .group-count {
      transition: none;
    }
  }
  .member-anchor {
    position: absolute;
    inset: 0;
    visibility: hidden;
    pointer-events: none;
  }
  /* The group's count sits on the card's top-right corner, clear of
     the top-left badges (CMD, GOAD) and the P/T pip at the bottom. */
  .group-count {
    position: absolute;
    top: -8px;
    right: -8px;
    z-index: 8;
    min-width: 24px;
    height: 20px;
    padding: 0 6px;
    box-sizing: border-box;
    border-radius: 10px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
    color: var(--accent);
    background: var(--surface, #0b0a09);
    border: 1px solid color-mix(in srgb, var(--accent) 60%, transparent);
    pointer-events: none;
    transition: top 180ms var(--ease);
  }
  /* #4431: a tapped pile turns its card about the box's centre, so the card's
     top-right corner is (h - w) / 2 down the pile. The badge follows it and
     stays upright; only its offset moves, never the pile's width. */
  .row:not(.strip) .pile.tapped .group-count {
    top: calc((var(--card-h, 123px) - var(--card-w, 88px)) / 2 - 8px);
  }
  /* #2209: the count badge overhangs the top-left corner, where an art
     tile's name strip starts, so the strip's text starts clear of it. */
  .row.strip .pile.multi {
    --art-name-inset: 15px;
  }
  .pile-count {
    position: absolute;
    top: -6px;
    left: -6px;
    z-index: 8;
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    box-sizing: border-box;
    border-radius: 9px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    color: var(--accent);
    background: var(--surface, #0b0a09);
    border: 1px solid color-mix(in srgb, var(--accent) 60%, transparent);
    pointer-events: none;
  }
</style>

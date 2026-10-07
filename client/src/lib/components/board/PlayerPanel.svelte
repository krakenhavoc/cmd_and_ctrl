<script lang="ts">
  // PlayerPanel is one player's full board, laid out in a CSS Grid
  // matching the S16.5 avatar-centric redesign:
  //
  //   ┌─────────────────────────────────────────────┐
  //   │  creatures (full width)                     │
  //   ├──────────────────────┬──────────────────────┤
  //   │  lands               │  enchant / artifact  │
  //   ├─────────────────────────────────────────────┤
  //   │  ⭕ piles  hand (peek) ……………… dock cell (self)│
  //   └─────────────────────────────────────────────┘
  //
  // The bottombar is a single flex row: avatar anchors bottom-left,
  // piles immediately beside it, the hand peeks up from the same base-
  // line (top ~55% of each card visible, bottom-clipped at the pile
  // bottom), and the action dock (ADR 0111) sits bottom-right over the viewer's own
  // panel. Hovering the self hand lifts the whole fan up over the board
  // to reveal full cards.
  //
  // The same component is reused for self + opponents; Board.svelte
  // wraps opponent instances in a transform container that rotates
  // them 180° across the top so they read "across the table". The
  // panel itself is layout-only and doesn't know whether it's rotated.
  //
  // Click routing for battlefield cards lives here so the combat
  // intercepts and the click rule stay in one place: targeting, combat
  // select on your own creature, an attack on a listed planeswalker or
  // battle, declare-block on an incoming attacker, and then ADR 0117's
  // click rule: the card's one usable ability activated (#2201), its
  // ability popover when two or more rows are usable and one is not
  // mana, its mana when only mana rows are usable (#1438), and
  // otherwise nothing. The rule itself is battlefieldClickPlan, in
  // contextMenu.logic, so it is testable without rendering Svelte.

  import type {
    ActionPayload,
    ActionType,
    CardView,
    GameView,
    ManaAbilityView,
    PlayerView,
    ZoneView,
  } from "../../protocol";
  import { defendingPlayerOf } from "../../attackTargets";
  import { takenFromByCard } from "../../takenFrom";
  import { ringBearerNames } from "../../ringEmblem";
  import { cantAttackByCard } from "../../cantAttack";
  import { bucketForBattlefield, isCreature, isLand } from "../../cardTypes";
  import {
    battlefieldClickPlan,
    type BattlefieldClickPlan,
    type LoneAbilityRow,
  } from "../../contextMenu.logic";
  import { canActivateSorcerySpeedAbility } from "../../timing";
  import { manaAbilityNeedsPrompt } from "../../manaAbilityCost";
  import { untrack } from "svelte";
  import {
    closeAbilityPopover,
    openAbilityPopover,
    setPopoverSurface,
    type PopoverSurface,
  } from "../../abilityPopover";
  import { findCardAnchor } from "../../boardAnchor";
  import { manaClickPlan, manaColorParams, type AnchorRect } from "../../manaSource";
  import { manaAbilityRef } from "../../abilityRef";
  import {
    closeManaSourcePicker,
    manaSourcePickerOpenFor,
    openManaSourcePicker,
  } from "../../manaSourcePicker";
  import BattlefieldRow from "./BattlefieldRow.svelte";
  import PileBar from "./PileBar.svelte";
  import Hand from "./Hand.svelte";
  import ExileStrip from "./ExileStrip.svelte";
  import CommandStrip from "./CommandStrip.svelte";
  import type { CastSourceZone } from "../../targeting";
  import PlayerIdentity from "./PlayerIdentity.svelte";
  import { seatColor } from "../../colors";
  import PromisesRow from "./PromisesRow.svelte";
  import TokenGroupModal from "./TokenGroupModal.svelte";
  import { groupMembersOf } from "../../tokenGroups";
  import { L } from "../../labels";
  import { canOverride, type MenuAction } from "../../contextMenu.logic";
  import {
    NO_COMBAT_RINGS,
    NO_LEGAL_ACTIONS,
    acrossActions,
    attackTargetListed,
    attackTargetOpen,
    combatRings,
    type LegalActions,
  } from "../../legalActions";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seat: PlayerView;
    isSelf: boolean;
    isActive: boolean;
    hasPriority: boolean;
    viewerID: string | null;
    isAdmin: boolean;
    sendAction: ActionSender;
    isMonarch: boolean;
    isInitiative: boolean;
    view: GameView;
    // Battlefield slice already filtered to cards with controller === seat.id
    controlledCards: CardView[];
    // Player-owned slice of the shared exile zone (filtered by Board)
    exile: ZoneView;
    combatMode: "idle" | "attack" | "block";
    selectedCombatCardID: string | null;
    onSelectCombatCard: (cardID: string) => void;
    onDeclareAttack: (targetPlayerID: string) => void;
    onDeclareBlock: (attackerCardID: string) => void;
    onTapToggle: (card: CardView) => void;
    // `fromZone` / `face` ride along for a cast out of the #1389
    // exile strip; a hand cast passes the card alone. #1508: `viaDrag`
    // is set by the hand's drag-to-cast gesture only.
    onPlayCard: (
      card: CardView,
      fromZone?: CastSourceZone,
      face?: number,
      viaDrag?: boolean,
    ) => void;
    onDrawCard: () => void;
    onTargetPlayer?: (targetPlayerID: string) => void;
    // onTargetCard returns true when a cast-targeting prompt
    // consumed the click (so the caller stops propagating into
    // the tap-toggle / combat default). Returns false when no
    // prompt is active or the card isn't a legal target.
    onTargetCard?: (card: CardView) => boolean;
    // S21 sub-PR 2: a CR 602 activated ability on one of this
    // seat's permanents was chosen from the card menu. Board owns
    // the follow-up (sacrifice pick, targeting) because those are
    // board-wide modals. On an opponent's panel the popover reaches it
    // only for an "Any player may activate this ability" row, the one
    // kind of row the viewer may activate there (ADR 0106 §1).
    onActivateAbility?: (card: CardView, abilityIndex: number) => void;
    // S21, widened in #789: a mana ability on one of this seat's
    // permanents needs a cost choice before it can be activated — a
    // sacrifice (Ashnod's Altar), or which counters come off and how
    // many (Mage-Ring Network, Iron Spider's cousin on a land). Board
    // owns those modals, so the panel forwards the click instead of
    // sending the action. Only wired for the viewer's own panel.
    // #1443: `colors` is the answer the anchored picker already has,
    // carried through the cost pickers into the one action.
    onManaAbilityCost?: (card: CardView, ability: ManaAbilityView, colors?: string[]) => void;
    // ADR 0111 §4 (owner decision 1): the action dock sits in the
    // screen's bottom-right corner, over this panel's corner. Set on the
    // viewer's own panel while the dock is mounted: the rail then stops
    // above the bottom row, and the bottom row keeps an empty cell the
    // dock's size (--dock-w × --dock-h) where the dock sits.
    docked?: boolean;
    // ADR 0076 §2.3 (amended 2026-10-02): the tutorial's coach card sits
    // in the screen's bottom-left corner, over this panel's. Set on the
    // viewer's own panel while the card shows: the bottom row then keeps
    // an empty cell the card's size (--coach-w × --coach-h) at its left,
    // so the hand centres in what is left and nothing sits under the card.
    // Unset — every game that is not the tutorial — the panel is unchanged.
    coached?: boolean;
    // flipped — top-row opponents. The panel keeps its zones in the
    // same grid but reverses the row order (hand at the top edge,
    // creatures toward the table centre) instead of rotating 180°,
    // so text and card art stay upright.
    flipped?: boolean;
    // #1071: a spectator seat (Board's "uniform grid" path — every
    // seat isSelf=false, none flipped). The 200px opponent ceiling
    // exists so an opponent panel doesn't rival the viewer's own —
    // a spectator has no panel of their own for it to defer to, so
    // that reason doesn't apply and the ceiling can match self's.
    // See .panel.opponent.spectator below.
    spectator?: boolean;
    // #1307: threaded straight through to PlayerIdentity — see its
    // prop doc.
    considering?: boolean;
    // #1724: "Attack <seat> with N" from a token group's list — one
    // declare_attackers action for exactly those attackers. Game owns
    // it because it is the same bulk declaration "attack with all"
    // sends, with the same refusal handling (the attack-tax picker).
    // Undefined hides the list's attack buttons; "Use this one" still
    // selects a single attacker the two-click way.
    onDeclareAttackers?: (attackerIDs: string[], defenderSeatID: string) => void;
    // ADR 0105 (#1789): the frame's legal-action lookup, already
    // "nothing" while highlights are off or autopass is about to pass
    // (Game.svelte). Handed to the cast surfaces for their ready rings
    // and counts; it gates nothing.
    legal?: LegalActions;
    // ADR 0105 sub-PR 3: the frame's FULL lookup, which the highlight
    // setting never touches. Read only by the ability popover's gate,
    // which greys a sorcery-speed row the server's digest leaves out.
    legalGate?: LegalActions;
    // ADR 0120 §3: this panel is the second copy of a seat's board,
    // drawn larger in the expanded overlay over the table. It is then
    // upright (never `flipped`), its cards cap at 240px for every seat,
    // `docked` and `coached` are off (the dock and the coach card belong
    // to the table's panel), and it is a `group` with no name, because
    // the overlay around it is the named region and "<name> board"
    // must stay unique. Its cards open the ability popover on the
    // `expanded` surface. Read once, at mount: a panel is the overlay's
    // or the table's for its whole life.
    expanded?: boolean;
  }

  const {
    seat,
    isSelf,
    isActive,
    hasPriority,
    viewerID,
    isAdmin,
    sendAction,
    isMonarch,
    isInitiative,
    view,
    controlledCards,
    exile,
    combatMode,
    selectedCombatCardID,
    onSelectCombatCard,
    onDeclareAttack,
    onDeclareBlock,
    onTapToggle,
    onPlayCard,
    onDrawCard,
    onTargetPlayer,
    onTargetCard,
    docked: dockedProp = false,
    coached: coachedProp = false,
    onActivateAbility,
    onManaAbilityCost,
    flipped: flippedProp = false,
    spectator = false,
    considering = false,
    onDeclareAttackers,
    legal = NO_LEGAL_ACTIONS,
    legalGate = NO_LEGAL_ACTIONS,
    expanded = false,
  }: Props = $props();

  // ADR 0120 §3: the expanded copy is upright, and leaves the dock's
  // and the coach card's cells to the table's panel.
  const flipped = $derived(flippedProp && !expanded);
  const docked = $derived(dockedProp && !expanded);
  const coached = $derived(coachedProp && !expanded);
  // The surface this panel's cards are drawn on, for the ability
  // popover (abilityPopover.ts): set as context for every Card below,
  // and used directly by this panel's own click router.
  const surface: PopoverSurface = untrack(() => expanded) ? "expanded" : "table";
  setPopoverSurface(surface);

  // ADR 0105: the battlefield rows read the lookups on the viewer's own
  // panel. A spectator never has a digest.
  //
  // ADR 0106 §1 decisions 5 and 6 (#1793): on an opponent's panel they
  // read `acrossActions`, which answers only for that panel's
  // "Any player may activate this ability" rows. The digest is the
  // viewer's own moves, so it lists another player's permanent exactly
  // when the viewer may activate such a row on it right now; that
  // permanent then wears the bolt pip and ring, and its popover's gate
  // reads the same answer. Nothing else lights through these lookups on
  // an opponent's panel; the combat rings below have their own.
  const battlefieldCards = $derived(view.battlefield?.cards ?? []);
  const rowLegal = $derived(
    spectator
      ? NO_LEGAL_ACTIONS
      : isSelf
        ? legal
        : acrossActions(legal, battlefieldCards, viewerID),
  );
  const rowGate = $derived(
    spectator
      ? NO_LEGAL_ACTIONS
      : isSelf
        ? legalGate
        : acrossActions(legalGate, battlefieldCards, viewerID),
  );
  // #1695: the life the popover's life-cost check reads. On the
  // viewer's own panel that is the seat's. On an opponent's panel the
  // only rows the popover lists are any-player rows, and the player
  // who activates one pays its cost (CR 602.1a): the viewer.
  const payerLife = $derived(isSelf ? seat.life : view.seats.find((s) => s.id === viewerID)?.life);

  // ADR 0105 sub-PR 5: the combat rings. Unlike the pips these reach
  // an opponent's panel, because what the SELECTED creature may be
  // declared against is on the other side of the table: the
  // planeswalkers and battles a selected attacker may attack, the
  // attackers a selected blocker may block. Both are facts of the
  // viewer's own digest about cards it already sees, so they reveal
  // nothing. The candidates half only ever lights the viewer's own
  // creatures, which are the only ones the digest has entries for. A
  // spectator has no digest.
  const rings = $derived(
    spectator
      ? NO_COMBAT_RINGS
      : combatRings(legal, combatMode, selectedCombatCardID, controlledCards),
  );

  // ADR 0105 sub-PR 4 (#1789): a CR 116.2 special action chosen from a
  // card's ability popover: foretell, suspend or plot from the hand,
  // turn face up on a face-down permanent. Until this PR these rows
  // lived only in the admin override menu. The row is built by the
  // same function that menu uses (contextMenu.logic.ts
  // specialActionItems), so this sends exactly what that menu's row
  // sends, through the same guarded sender. The viewer's own panel
  // only; the server strips `special_actions` from everyone else.
  const sendSpecialAction = $derived(
    isSelf && !spectator
      ? (action: MenuAction) =>
          sendAction(
            action.type,
            action.params as ActionPayload["params"],
            action.player ?? seat.id,
          )
      : undefined,
  );

  // The seat's commander, wherever it is right now: the command zone
  // first, then the shared zones (battlefield, stack, exile) and its
  // graveyard. Feeds the art-crop avatar fallback in PlayerIdentity.
  const commanderScryfallID = $derived.by((): string | null => {
    const inCommand = seat.command?.cards?.find((c) => c.scryfall_id);
    if (inCommand?.scryfall_id) return inCommand.scryfall_id;
    const zones = [view.battlefield, view.stack, view.exile, seat.graveyard];
    for (const z of zones) {
      const hit = z?.cards?.find(
        (c) => c.is_commander && (c.owner === seat.id || c.controller === seat.id) && c.scryfall_id,
      );
      if (hit?.scryfall_id) return hit.scryfall_id;
    }
    return null;
  });

  // S24 (ADR 0036 decisions 13 + 14): the wire carries attachment in
  // one direction — each Equipment / Aura names its host — and the
  // reverse list is derived here rather than shipped, so the two can
  // never disagree. Keyed by host instance ID over the WHOLE
  // battlefield, not just this seat's cards: an Aura you control on
  // a creature an opponent controls is drawn on the creature, which
  // is where the rules put it.
  const attachmentsByHost = $derived.by(() => {
    const out: Record<string, CardView[]> = {};
    for (const c of view.battlefield?.cards ?? []) {
      if (c.attached_to?.kind !== "card" || !c.attached_to.id) continue;
      (out[c.attached_to.id] ??= []).push(c);
    }
    return out;
  });

  // S24 (ADR 0036 decision 14 item 3): a Curse enchants a PLAYER, so
  // it has no host card to hide behind and stays in its controller's
  // "enchant / artifact" row. Without a badge naming its victim the
  // board says nothing about who is being cursed, which is the whole
  // card — so each such permanent gets the enchanted seat's name.
  //
  // Drawing it in the ENCHANTED player's panel would read better
  // still, but it makes "where a card is drawn" diverge from
  // card.controller, and that is a new concept for this UI.
  const curseTargets = $derived.by(() => {
    const names = new Map((view.seats ?? []).map((s) => [s.id, s.name]));
    const out: Record<string, string> = {};
    for (const c of view.battlefield?.cards ?? []) {
      if (c.attached_to?.kind !== "player" || !c.attached_to.id) continue;
      out[c.instance_id] = names.get(c.attached_to.id) ?? "a player";
    }
    return out;
  });

  // ADR 0104 (owner decision 6): the owner of each permanent another
  // player controls, for Card's TAKEN FROM badge.
  const takenFrom = $derived(takenFromByCard(view.battlefield?.cards, view.seats));
  // ADR 0114: "Alice's Ring-bearer", for the marker's title.
  const ringBearers = $derived(ringBearerNames(view.battlefield?.cards, view.seats));
  // ADR 0106 §2 (owner decision 3): whom each creature can't attack,
  // read off the card view, for Card's CAN'T ATTACK chip.
  const cantAttack = $derived(cantAttackByCard(view.battlefield?.cards, view.seats));

  // Every battlefield card that is drawn behind a host rather than in
  // its own type row. A dangling attachment — the host has left but
  // the state-based action has not swept the relation yet — keeps its
  // own row, so an Equipment never vanishes mid-frame.
  const hostedCardIDs = $derived.by(() => {
    const onBattlefield = new Set((view.battlefield?.cards ?? []).map((c) => c.instance_id));
    const out = new Set<string>();
    for (const c of view.battlefield?.cards ?? []) {
      const host = c.attached_to;
      if (host?.kind === "card" && host.id && onBattlefield.has(host.id)) {
        out.add(c.instance_id);
      }
    }
    return out;
  });

  // S31: the CR 307.1 sorcery-speed window, derived once per panel
  // and handed down to every Card so the ability popover can grey an
  // "activate only as a sorcery" row. The flag has ridden the wire as
  // ActivatedAbilityView.sorcery_speed since S21 and nothing read it,
  // so those abilities stayed clickable through combat and an
  // opponent's turn and came back rejected — the live example
  // ADR 0033 §1 cites for why the client stopped re-deriving timing.
  //
  // ADR 0105 sub-PR 3: it is the popover's WORDS now, not its verdict.
  // Whether a row is shut is the server's answer, from the row's
  // `timing_closed` and the legal-action digest (`legalGate`). This
  // string only says why.
  //
  // Empty string means "open, no opinion"; opponents' panels are
  // never gated on the VIEWER's window, so they get "" too.
  const sorcerySpeedBlocked = $derived(
    isSelf ? (canActivateSorcerySpeedAbility(view, viewerID).reason ?? "") : "",
  );
  // ADR 0106 §1: the battlefield rows' words. An opponent's permanent
  // offers the viewer only its any-player rows, whose timing is the
  // ACTIVATOR's (CR 109.5), so a row the server shut is explained in the
  // viewer's window there too. A spectator opens no popover.
  const rowTimingWords = $derived(
    isSelf || spectator
      ? sorcerySpeedBlocked
      : (canActivateSorcerySpeedAbility(view, viewerID).reason ?? ""),
  );

  // #1724: the token group whose member list is open, by group key.
  // The members are re-derived from every snapshot, so the list
  // follows the board while it is open; a group that has broken up
  // (down to one token, or none) closes it.
  let openGroupKey = $state<string | null>(null);
  const openGroupMembers = $derived.by(() => {
    if (!openGroupKey) return [];
    const rowCards = controlledCards
      .filter((c) => !hostedCardIDs.has(c.instance_id))
      .sort((a, b) => (a.battle_x ?? 0) - (b.battle_x ?? 0));
    return groupMembersOf(rowCards, attachmentsByHost, openGroupKey);
  });
  $effect(() => {
    if (openGroupKey && openGroupMembers.length === 0) openGroupKey = null;
  });
  // Tap / Untap in the list: the same permission a click has — the
  // controller, or an admin.
  const canTapGroup = $derived(
    openGroupMembers.length > 0 && canOverride(openGroupMembers[0], viewerID, isAdmin),
  );
  function declareGroupBlockers(blockerIDs: string[], attackerID: string): void {
    for (const id of blockerIDs) {
      sendAction("declare_blocker", { blocker: id, attacker: attackerID });
    }
  }

  const buckets = $derived.by(() => {
    const out = { creature: [] as CardView[], land: [] as CardView[], right: [] as CardView[] };
    for (const c of controlledCards) {
      if (hostedCardIDs.has(c.instance_id)) continue;
      out[bucketForBattlefield(c)].push(c);
    }
    return out;
  });

  // S15: mana-ability activation handler for battlefield permanents
  // the viewer controls. Only installed on the viewer's own panel;
  // opponent panels pass undefined down so the context menu stays
  // closed on cards they don't control.
  //
  // S21, widened in #789: a mana ability whose cost sacrifices
  // ANOTHER permanent (Ashnod's Altar), or whose counter component
  // still has a choice in it (which permanent, which kind, how many),
  // needs an answer first — and those modals are board-wide. So the
  // click is handed to Board, which owns the same SacrificeCostModal
  // and CounterCostModal the CR 602 abilities use and sends the
  // action itself once the cost is settled.
  //
  // A card-shaped cost goes the same way: Skirge Familiar's discard
  // (#1213) and Cadaverous Bloom's "Exile a card from your hand"
  // (#1283). This click path used to skip the discard, so a board
  // click on Skirge Familiar sent no `discard_ids` and was refused.
  //
  // manaAbilityNeedsPrompt is the one predicate; a plain "{T}: Add
  // {G}", and a Vivid land with charge counters on it, go straight to
  // the action as they always did.
  const activateManaAbility = $derived(
    isSelf
      ? (card: CardView, abilityIndex: number, colors?: string[]) => {
          // #1228: a permanent publishes `mana_abilities` and a card
          // in hand whose mana ability functions there publishes
          // `zone_mana_abilities` — never both. One lookup reads
          // whichever is present, exactly as the card menu does.
          const rows = card.mana_abilities ?? card.zone_mana_abilities ?? [];
          const ability = rows.find((a) => a.index === abilityIndex);
          if (ability && onManaAbilityCost && manaAbilityNeedsPrompt(ability)) {
            onManaAbilityCost(card, ability, colors);
            return;
          }
          // #1443: a colour chosen at the card rides the activation,
          // so the server produces it with no second question.
          sendAction(
            "activate_mana_ability",
            {
              card_id: card.instance_id,
              ability_index: abilityIndex,
              // ADR 0093: the row this click meant, so a stale one is refused.
              ...manaAbilityRef(card, abilityIndex),
              ...manaColorParams(colors),
            },
            seat.id,
          );
        }
      : undefined,
  );

  // ADR 0105 sub-PR 5: the player half of "the selected attacker's
  // defenders light" — the identity's green ring, which is also the
  // click. Read off the FULL lookup: it withholds a click the server
  // would refuse, and the highlight setting must not change that. No
  // list, or a re-point of an attacker already declared, keeps the old
  // rule (attackTargetOpen).
  const attackTargetable = $derived(
    !isSelf &&
      !seat.eliminated &&
      combatMode === "attack" &&
      !!selectedCombatCardID &&
      attackTargetOpen(
        legalGate,
        view.battlefield?.cards?.find((c) => c.instance_id === selectedCombatCardID),
        selectedCombatCardID,
        seat.id,
      ),
  );

  function handleCardClick(card: CardView, ev?: MouseEvent): void {
    // The S17 sub-PR 5 Shift+click / Shift+Alt+click +1/+1 debug
    // chord used to live here. #170 retired it: the right-click
    // override menu offers both directions on every counter type
    // (and every other manual override) from a surface the player
    // can find without being told it exists. Enable it under
    // Settings → Gameplay → "Enable admin overrides".
    //
    // S14: targeting intercept. If a cast-targeting prompt is live
    // and this card is a legal target (battlefield creature for
    // "any" / "creature" modes), route through onTargetCard. Board
    // clears the targeting state when cast_spell fires.
    if (onTargetCard) {
      // onTargetCard itself checks the targeting store; only call
      // when a prompt is waiting. Board wires this to the
      // targeting-aware completion.
      // To avoid double-handling, we check the targeting store here
      // via a light dynamic import — but easier: the callback
      // returns a boolean "handled" signal. Lacking that, just
      // call unconditionally: Board's handler is a no-op when no
      // prompt is active.
      if (onTargetCard(card)) return;
    }
    // Combat select on viewer's own creature wins ahead of tap/untap.
    if (
      (combatMode === "attack" || combatMode === "block") &&
      card.controller === viewerID &&
      isCreature(card)
    ) {
      onSelectCombatCard(card.instance_id);
      return;
    }
    // ADR 0105 sub-PR 5: an attack-mode click on a planeswalker or a
    // battle the selected attacker may attack declares that attack
    // (CR 508.1d). It is the card the ready ring is on, and a ring
    // that pointed at nothing would be worse than none. Only what the
    // server lists: with no list the click stays what it was.
    if (
      combatMode === "attack" &&
      attackTargetListed(legalGate, selectedCombatCardID, card.instance_id)
    ) {
      onDeclareAttack(card.instance_id);
      return;
    }
    // Block-mode click on an incoming attacker commits the block.
    // "Incoming" is CR 802.4a's: attacking the viewer, a planeswalker
    // they control or a battle they protect (#1339).
    if (combatMode === "block" && defendingPlayerOf(card) === viewerID) {
      onDeclareBlock(card.instance_id);
      return;
    }
    // ADR 0117 §1: the click rule. A left-click acts on what the card
    // does: its one usable ability, activated as its popover row would
    // activate it (#2201); its popover when two or more rows are usable
    // and one is not mana (at the card, never the override menu); its
    // mana when only mana rows are usable; nothing otherwise. Alt-click
    // still raw-taps wherever the viewer may drive the card.
    const plan = clickPlan(card, !!ev?.altKey);
    switch (plan.intent) {
      case "none":
        return;
      case "activate":
        activateLoneRow(card, plan.row);
        return;
      case "popover":
        openAbilityPopover(card.instance_id, surface);
        return;
      case "mana":
        clickForMana(card, ev);
        return;
      case "tap":
        onTapToggle(card);
    }
  }

  // clickPlan asks the click rule with exactly what this panel hands
  // the popover (BattlefieldRow → Card → ManaAbilityMenu): the same
  // wiring, life, timing words and digest gate, so a click never
  // disagrees with the menu it would open (ADR 0117 §2).
  function clickPlan(card: CardView, rawTap: boolean): BattlefieldClickPlan {
    return battlefieldClickPlan(card, viewerID, isAdmin, {
      manaClick: !!activateManaAbility,
      special: !!sendSpecialAction,
      rawTap,
      view,
      payerLife,
      timingWords: rowTimingWords,
      legalGate: rowGate,
    });
  }

  // ADR 0117 §1, the cursor: a card whose left-click would do nothing
  // right now drops its pointer affordance. Every intercept above that
  // could take the click keeps it (a live targeting ring is Card's own
  // to read).
  function clickInert(card: CardView): boolean {
    if (
      (combatMode === "attack" || combatMode === "block") &&
      card.controller === viewerID &&
      isCreature(card)
    ) {
      return false;
    }
    if (
      combatMode === "attack" &&
      attackTargetListed(legalGate, selectedCombatCardID, card.instance_id)
    ) {
      return false;
    }
    if (combatMode === "block" && defendingPlayerOf(card) === viewerID) return false;
    return clickPlan(card, false).intent === "none";
  }

  // activateLoneRow is the "activate" branch (#2201): the card's one
  // usable row, sent down the very callback its popover row calls
  // (BattlefieldRow → Card → ManaAbilityMenu), so its costs, X, modes
  // and targets run exactly as they do from the popover. An activated
  // row (own, granted, loyalty, any-player) goes to the board's
  // activation flow; a special action and a manual loyalty row send
  // the action their row sends. No menu opens, so nothing is said to
  // the tutorial. A popover left open on another card closes, as
  // choosing a row closes it.
  function activateLoneRow(card: CardView, row: LoneAbilityRow): void {
    closeAbilityPopover();
    if (row.kind === "activated") {
      onActivateAbility?.(card, row.index);
      return;
    }
    sendSpecialAction?.(row.action);
  }

  // clickForMana is the "mana" branch of the click router (#1438): one
  // ability goes out now, several open the anchored picker. Clicking
  // the card whose picker is already open closes it instead, so the
  // card is its own toggle.
  function clickForMana(card: CardView, ev?: MouseEvent): void {
    if (manaSourcePickerOpenFor(card.instance_id)) {
      closeManaSourcePicker();
      return;
    }
    const plan = manaClickPlan(card, { payerLife });
    if (!plan || !activateManaAbility) return;
    if (plan.kind === "activate") {
      activateManaAbility(card, plan.index, plan.colors);
      return;
    }
    // ADR 0117 §4: one ability with two or more picking slots opens
    // the picker straight on its per-colour stepper.
    openManaSourcePicker({
      cardID: card.instance_id,
      anchor: anchorFor(card, ev),
      ...(plan.kind === "split" ? { abilityIndex: plan.index } : {}),
    });
  }

  // The clicked card's box: the element the click landed on, or the
  // tile found by its instance ID (a keyboard activation has no
  // pointer), or failing both a point at the cursor. The lookup goes
  // through boardAnchor, so an expanded overlay's copy wins (ADR 0120 §3).
  function anchorFor(card: CardView, ev?: MouseEvent): AnchorRect {
    const target = ev?.currentTarget ?? ev?.target;
    let el = target instanceof Element ? target.closest("[data-instance-id]") : null;
    if (!el && typeof document !== "undefined") {
      el = findCardAnchor(document, card.instance_id);
    }
    if (el) {
      const r = el.getBoundingClientRect();
      return { left: r.left, top: r.top, right: r.right, bottom: r.bottom };
    }
    const x = ev?.clientX ?? 0;
    const y = ev?.clientY ?? 0;
    return { left: x, top: y, right: x, bottom: y };
  }
</script>

<div
  class="panel seat-panel"
  class:self={isSelf}
  class:seat-self={isSelf}
  class:opponent={!isSelf}
  class:flipped
  class:spectator
  class:docked
  class:expanded
  class:seat-turn={isActive}
  style:--seat-color={seatColor(seat.seat)}
  role={expanded ? "group" : "region"}
  aria-label={expanded ? undefined : isSelf ? L.yourBoard : L.seatBoard(seat.name)}
>
  <div class="grid-creatures">
    <BattlefieldRow
      label={L.creatures}
      {attachmentsByHost}
      {curseTargets}
      {takenFrom}
      {ringBearers}
      {cantAttack}
      cards={buckets.creature}
      {viewerID}
      {selectedCombatCardID}
      onCardClick={handleCardClick}
      onActivateManaAbility={activateManaAbility}
      onRawTap={activateManaAbility ? onTapToggle : undefined}
      {view}
      {clickInert}
      {onActivateAbility}
      sorcerySpeedBlocked={rowTimingWords}
      {payerLife}
      legal={rowLegal}
      combat={rings}
      legalGate={rowGate}
      onSpecialAction={sendSpecialAction}
      onGroupClick={(k) => (openGroupKey = k)}
    />
  </div>
  <!-- The back row: land piles first, then the other permanents,
       both shrink-wrapped and centred as one group so the row grows
       out from the middle like the creature row above it. -->
  <div class="grid-middle">
    <BattlefieldRow
      label={L.lands}
      {attachmentsByHost}
      {curseTargets}
      {takenFrom}
      {ringBearers}
      {cantAttack}
      cards={buckets.land}
      compact
      strip
      {viewerID}
      {selectedCombatCardID}
      onCardClick={handleCardClick}
      onActivateManaAbility={activateManaAbility}
      onRawTap={activateManaAbility ? onTapToggle : undefined}
      {view}
      {clickInert}
      {onActivateAbility}
      sorcerySpeedBlocked={rowTimingWords}
      {payerLife}
      legal={rowLegal}
      combat={rings}
      legalGate={rowGate}
      onSpecialAction={sendSpecialAction}
      onGroupClick={(k) => (openGroupKey = k)}
    />
    <BattlefieldRow
      label="enchant / artifact"
      {attachmentsByHost}
      {curseTargets}
      {takenFrom}
      {ringBearers}
      {cantAttack}
      cards={buckets.right}
      compact
      {viewerID}
      {selectedCombatCardID}
      onCardClick={handleCardClick}
      onActivateManaAbility={activateManaAbility}
      onRawTap={activateManaAbility ? onTapToggle : undefined}
      {view}
      {clickInert}
      {onActivateAbility}
      sorcerySpeedBlocked={rowTimingWords}
      {payerLife}
      legal={rowLegal}
      combat={rings}
      legalGate={rowGate}
      onSpecialAction={sendSpecialAction}
      onGroupClick={(k) => (openGroupKey = k)}
    />
  </div>
  <div class="grid-bottom">
    {#if coached}
      <!-- ADR 0076 §2.3: the coach card's cell. Empty on purpose, like
           .dock-spacer — the card is Game.svelte's, positioned over this
           exact rectangle. -->
      <div class="coach-spacer" aria-hidden="true"></div>
    {/if}
    <!-- Hand sits inline with the dock cell; its clipped-bottom
         line coincides with the panel edge. flex: 1 lets it absorb
         the width the rail freed up. -->
    <div class="hand-zone">
      <Hand
        hand={seat.hand}
        {isSelf}
        onPlayCard={isSelf ? onPlayCard : undefined}
        // #1920: a dragged land goes the click path's way (no strict/auto_tap
        // stamp: there is nothing to pay), so the action is the same one.
        onDragCast={isSelf ? (c) => onPlayCard(c, undefined, undefined, !isLand(c)) : undefined}
        onActivateAbility={isSelf ? onActivateAbility : undefined}
        onActivateManaAbility={activateManaAbility}
        {sorcerySpeedBlocked}
        snap={view}
        {viewerID}
        {legal}
        legalGate={rowGate}
        onSpecialAction={sendSpecialAction}
      />
    </div>
    {#if isSelf}
      <!-- #1389: the exiled cards this seat may cast, as a second
           hand. Renders nothing when there are none. -->
      <ExileStrip
        {view}
        {viewerID}
        onCastCard={onPlayCard}
        onDragCast={(c, zone, face) => onPlayCard(c, zone, face, true)}
        {legal}
        {onActivateAbility}
        {sorcerySpeedBlocked}
        legalGate={rowGate}
      />
    {/if}
    {#if !isSelf}
      <!-- #2349: this seat's commander, beside its hand, where yours
           is beside yours. -->
      <CommandStrip {seat} />
      <PromisesRow {view} {viewerID} opponentID={seat.id} {sendAction} />
    {/if}
    {#if docked}
      <!-- ADR 0111 §4: the action dock's cell. Empty on purpose — the
           dock is Game.svelte's, positioned over this exact rectangle,
           so the hand and the rail can never sit under it. -->
      <div class="dock-spacer" aria-hidden="true"></div>
    {/if}
  </div>
  <!-- The rail is the player card: identity, floating mana and
       markers on top, the four piles below. On the viewer's own panel
       it stops above the bottom row, whose right end is the action
       dock's (ADR 0111 owner decision 1). -->
  <div class="rail">
    <PlayerIdentity
      {seat}
      {commanderScryfallID}
      {isSelf}
      {isActive}
      {hasPriority}
      {attackTargetable}
      {isMonarch}
      {isInitiative}
      {sendAction}
      {onDeclareAttack}
      {onTargetPlayer}
      {considering}
    />
    <div class="rail-gap"></div>
    <PileBar
      {seat}
      {exile}
      {isSelf}
      {sendAction}
      onDrawCard={isSelf ? onDrawCard : undefined}
      {onPlayCard}
      {legal}
    />
  </div>
  <TokenGroupModal
    groupKey={openGroupKey ?? ""}
    members={openGroupMembers}
    {attachmentsByHost}
    {view}
    {viewerID}
    {combatMode}
    canTap={canTapGroup}
    onUse={handleCardClick}
    onTarget={onTargetCard}
    {onTapToggle}
    onAttack={isSelf ? onDeclareAttackers : undefined}
    onBlock={isSelf ? declareGroupBlockers : undefined}
    onClose={() => (openGroupKey = null)}
    legalGate={rowGate}
  />
</div>

<style>
  .panel {
    /* Sept 2026 redesign: zones on the left, a player rail on the
       right. Creatures on top at full size; the middle band is the
       land piles and the other permanents, one size down; the
       bottom row is the hand plus (self only) the action dock's cell.
       Top-row opponents set `flipped`, which reverses the row order
       instead of rotating the panel.

       Card size is a function of the PANEL'S HEIGHT, not a constant
       (board ratios, Sept 2026): the panel is a size container and
       every card height is a share of it, clamped so a short panel
       never drops below the fixed sizes the redesign shipped with
       (120×168 / 88×123 / 64×90) and a tall one stops at 240px. The
       share is the panel's budget solved for the creature card: the
       creature row (1×), the back row (0.74×) and the hand's peek
       (0.62× self, 0.55× opponents) plus ~72px of row padding and
       gaps must fit, so cre ≤ 0.42·H − 30 (0.43·H − 31 with the
       tighter peek). The card-size setting scales the share, so
       "large" can outgrow a 900px-tall window — the rows scroll
       rather than clip, as they did with the fixed 1.25× sizes. The
       5:7 card aspect is kept by deriving the width.

       --card-w / --card-h cascade into every nested Card; the
       compact rows override them with --card-w-sm / --card-h-sm.
       --thumb-w / --thumb-h size the pile thumbnails in the rail. */
    container-type: size;
    /* --card-h-max is the clamp's ceiling, pulled into its own token
       (#1071) so a variant below can raise it without restating the
       floor and slope — which would drift the two out of step the
       next time either changes. */
    --card-h-max: 240px;
    --card-h: clamp(168px, calc((42cqh - 30px) * var(--card-scale, 1)), var(--card-h-max));
    --card-w: calc(var(--card-h) * 5 / 7);
    --card-h-sm: clamp(123px, calc(var(--card-h) * 0.74), 178px);
    --card-w-sm: calc(var(--card-h-sm) * 5 / 7);
    --thumb-w: calc(40px * var(--card-scale, 1));
    --thumb-h: calc(56px * var(--card-scale, 1));
    --avatar-size-base: 84px;
    --rail-w: 112px;
    display: grid;
    grid-template-columns: minmax(0, 1fr) var(--rail-w);
    grid-template-rows: minmax(0, 1fr) auto auto;
    grid-template-areas:
      "creatures rail"
      "middle    rail"
      "bottom    rail";
    gap: 6px;
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    padding: 8px;
    border-radius: 14px;
    overflow: hidden;
    position: relative;
  }
  /* ADR 0111 §4 (owner decision 1): the action dock owns the screen's
     bottom-right corner, which is this panel's. The rail spans the
     creature and middle rows only; the bottom row runs the full width
     and ends in .dock-spacer, an empty cell the dock's live size. So
     nothing of the panel sits under the dock at rest, and the piles end
     above it (the rail scrolls sooner on a short panel). */
  .panel.docked {
    grid-template-areas:
      "creatures rail"
      "middle    rail"
      "bottom    bottom";
  }
  .dock-spacer {
    flex: 0 0 var(--dock-w, 0px);
    height: var(--dock-h, 0px);
    align-self: flex-end;
  }
  /* ADR 0076 §2.3 (amended again, #1081 follow-up): the coach's cell
     takes WIDTH from the bottom row only, never height. It stretches to
     whatever height the hand and the dock's cell already give the row,
     so the battlefield rows above keep every pixel they have without a
     tutorial. The card is taller than the row, so it rises past the
     row's top edge over the battlefield's left margin, which is empty:
     both battlefield rows centre their cards. Giving the cell the card's
     height instead cost the creature row 40px at 1280×800 and at
     1920×1080, nearly a third of it at 1280. */
  .coach-spacer {
    flex: 0 0 var(--coach-w, 0px);
    align-self: stretch;
  }
  /* ADR 0111 §8: on a phone the dock is a bar under the board, not in
     this corner, so the cell goes and the rail keeps its full height. */
  @media (max-width: 599px) {
    .panel.docked {
      grid-template-areas:
        "creatures rail"
        "middle    rail"
        "bottom    rail";
    }
    .dock-spacer,
    .coach-spacer {
      display: none;
    }
  }
  .panel.opponent {
    /* Upright opponent (the "next" seat) gets the medium scale — it
       has the tall row to itself. Scales by --card-scale-opponent
       (settings) with a gentler curve than self. The 200px ceiling
       (vs. self's 240px) is deliberate: an opponent panel shares the
       screen with the viewer's own and shouldn't rival it. */
    --card-h-max: 200px;
    /* #2336: sized by relevance, like your own board. The face-down
       hand shows only a count, so it peeks 30%; lands and other
       permanents are half a creature. The creature card is what is
       left: H = 1.92h + ~64px (the commander beside the hand, CommandStrip, is 0.42h), so h ≈ 52cqh - 33px. The floor is low so
       a short panel shrinks its cards rather than scrolling them. */
    --hand-peek: 0.3;
    --card-h: clamp(90px, calc((52cqh - 33px) * var(--card-scale-opponent, 1)), var(--card-h-max));
    --card-h-sm: clamp(56px, calc(var(--card-h) * 0.5), 110px);
    --thumb-w: calc(36px * var(--card-scale-opponent, 1));
    --thumb-h: calc(50px * var(--card-scale-opponent, 1));
    --avatar-size-base: 72px;
    --rail-w: 104px;
  }
  /* #1071: a spectator has no panel of their own for an opponent
     panel to defer to, so the reason for the 200px ceiling above
     doesn't apply — every seat is equal, and the ceiling can match
     self's. This rule and .panel.opponent live in the same
     stylesheet (same Svelte scope hash on both), so there is no
     cross-component cascade tie to reason about — one extra class
     over .panel.opponent is ordinary CSS specificity. The floor and
     slope are untouched, so nothing changes below ~1340px of panel
     height. Board never sets `flipped` on a spectator panel today
     (every spectator seat mounts isSelf=false with no `flipped`), so
     .panel.opponent.flipped below never competes with this rule in
     practice — the dedicated four-class rule after it is what keeps
     that true even if a spectator panel ever does gain `flipped`. */
  .panel.opponent.spectator {
    --card-h-max: 240px;
  }
  .panel.opponent.flipped {
    /* Across-table seats: one more size down (the top row is the
       short one), rows reversed so the hand hugs the top edge and
       creatures face the centre of the table. */
    --card-h-max: 168px;
    --card-h: clamp(70px, calc((52cqh - 33px) * var(--card-scale-opponent, 1)), var(--card-h-max));
    --card-h-sm: clamp(48px, calc(var(--card-h) * 0.5), 90px);
    --thumb-w: calc(32px * var(--card-scale-opponent, 1));
    --thumb-h: calc(45px * var(--card-scale-opponent, 1));
    --avatar-size-base: 60px;
    grid-template-rows: auto auto minmax(0, 1fr);
    grid-template-areas:
      "bottom    rail"
      "middle    rail"
      "creatures rail";
  }
  /* Defensive only — unreachable today (see above), kept so a future
     spectator panel that gains `flipped` still gets the raised
     ceiling instead of silently falling back to 168px. Four classes
     beats both .panel.opponent.flipped (three) and
     .panel.opponent.spectator (three) outright, so this can never
     lose to either regardless of source order. */
  .panel.opponent.flipped.spectator {
    --card-h-max: 240px;
  }
  /* ADR 0120 §3: the expanded overlay's copy of a board caps its cards
     at 240px for every seat, as on a spectator's panel. Self is 240px
     already, and an expanded panel is never flipped, so only
     .panel.opponent's 200px has to be beaten. The floor and slope are
     the seat's own; the overlay's height does the rest. */
  .panel.opponent.expanded {
    --card-h-max: 240px;
  }
  /* The creature area is the grid's minmax(0, 1fr) row, so on a short
     panel it is shorter than one card (whose height floors at 168px).
     The area is a flex column and the row may shrink, so the row's own
     overflow: auto scrolls it inside its area instead of the cards
     painting over the lands below (ADR 0111 PR 2: with the action dock
     in the bottom row this happens on more screens than before). */
  .grid-creatures {
    grid-area: creatures;
    min-height: 0;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .grid-creatures > :global(.row) {
    flex: 0 1 auto;
  }
  /* A flipped panel's creature row is its LAST row, and it must sit
     at the bottom of its area — against the middle of the table,
     facing the viewer's own creatures — not float up under the
     lands. The area keeps the leftover height; the row is pushed to
     its far edge. */
  .flipped .grid-creatures {
    justify-content: flex-end;
  }
  .grid-middle {
    grid-area: middle;
    min-height: 0;
    min-width: 0;
    display: flex;
    /* safe: a back row wider than its area starts at its left edge
       rather than spilling left over the piles beside it (#2438). */
    justify-content: safe center;
    align-items: flex-start;
    gap: 24px;
  }
  /* Each back-row zone takes the width of its cards, no more, so the
     two read as one centred group. An empty zone keeps enough width
     to show its label. */
  .grid-middle > :global(.row) {
    flex: 0 1 auto;
    min-width: 160px;
    max-width: 100%;
  }
  .grid-bottom {
    grid-area: bottom;
    min-height: 0;
    min-width: 0;
    display: flex;
    align-items: flex-end;
    gap: 8px;
  }
  .flipped .grid-bottom {
    align-items: flex-start;
  }
  .hand-zone {
    flex: 1 1 0;
    min-width: 0;
    align-self: flex-end;
    /* Fixed rest-state height matches the Hand's clipped peek. Keeping
       it explicit means the Hand's hover lift (max-height: none plus a
       translateY transform) doesn't grow this wrapper and push the
       row taller — the lift stays purely visual. */
    height: calc(var(--card-h, 168px) * var(--hand-peek, 0.55));
    overflow: visible;
    position: relative;
  }
  /* The viewer's own hand shows 62% of each card at rest (name, cost,
     art and the type line); opponents' face-down fans keep the
     tighter 55%. Hand.svelte's peek and lift use the same share. */
  .panel.self .hand-zone {
    height: calc(var(--card-h, 168px) * var(--hand-peek, 0.62));
  }
  .flipped .hand-zone {
    align-self: flex-start;
  }
  .rail {
    grid-area: rail;
    min-height: 0;
    min-width: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    padding: 6px 0 2px 8px;
    border-left: 1px solid var(--border);
  }
  .rail-gap {
    flex: 1 1 0;
  }
  .flipped .rail-gap {
    flex: 0 0 4px;
  }
  /* Short panels (the top row at 900px tall) must never clip the
     piles: the rail scrolls before it hides anything, and the
     across-table rails drop the pile labels (the tile's title and
     aria-label still carry them). */
  .rail {
    overflow-y: auto;
    overflow-x: hidden;
  }
  /* #2336: your board is sized by relevance, as in Arena, in every
     layout. Creatures get the largest cards. Lands and other permanents
     behind them are about half that size. The hand shows its top 40%
     and lifts on hover. The piles leave the rail for the empty corner
     left of the back row, four across, and the rail keeps only your
     identity. The rail is `display: contents`, so its children are this
     grid's items.

     The creature card's height is what the panel has left once the
     rest is placed, so the row fits instead of scrolling: with the back
     row at half a card and the hand at 0.4 of one, H = 1.9h + ~64px of
     padding, gaps and labels, so h ≈ 52cqh - 34px. Desktop only: a
     phone keeps the docked layout above. */
  @media (min-width: 600px) {
    .panel.self.docked {
      --card-h: clamp(120px, calc((52cqh - 40px) * var(--card-scale, 1)), var(--card-h-max));
      --card-h-sm: clamp(64px, calc(var(--card-h) * 0.5), 110px);
      --hand-peek: 0.4;
      grid-template-columns: auto minmax(0, 1fr) var(--rail-w);
      grid-template-areas:
        "creatures creatures rail"
        "piles     middle    rail"
        "bottom    bottom    bottom";
    }
    .panel.self.docked .rail {
      display: contents;
    }
    /* The bottom row is the hand's peek, not the dock's height: the
       dock rises over the back row's right end instead, which the back
       row keeps clear with padding. That height goes to the creatures. */
    .panel.self.docked .dock-spacer {
      height: 0;
    }
    .panel.self.docked .grid-middle {
      padding-right: max(0px, calc(var(--dock-w, 0px) - var(--rail-w)));
    }
    .panel.self.docked .rail-gap {
      display: none;
    }
    .panel.self.docked .rail > :global(.identity) {
      grid-area: rail;
      align-self: start;
    }
    .panel.self.docked .rail > :global(.pile-bar) {
      grid-area: piles;
      grid-template-columns: repeat(3, 52px);
      width: auto;
      align-self: end;
    }
    /* As on the top-row panels: the tile's title and aria-label still
       carry the pile's name, so the piles stay one short row. */
    .panel.self.docked .rail :global(.pile .label) {
      display: none;
    }
  }
  /* #2438, #2483: an across-table opponent sits across from you. Their
     board is yours turned 180° about the table's centre, with the cards
     themselves left upright to read: their hand along the top edge with
     their commander on its left (yours is on your hand's right), the
     back row under it reading enchantments/artifacts then lands with
     the piles in its right corner (yours are in your back row's left),
     and the creatures facing yours, with their identity at the bottom
     of the rail on the left, where yours is at the top of your rail on
     the right. The rail is `display: contents`, as on yours, so the
     piles leave it and the creatures keep its full height. */
  @media (min-width: 600px) {
    .panel.opponent.flipped {
      grid-template-columns: var(--rail-w) minmax(0, 1fr) auto;
      grid-template-rows: auto auto minmax(0, 1fr);
      grid-template-areas:
        "bottom bottom    bottom"
        "rail   middle    piles"
        "rail   creatures creatures";
    }
    .panel.opponent.flipped .grid-bottom,
    .panel.opponent.flipped .grid-middle {
      flex-direction: row-reverse;
    }
    .panel.opponent.flipped .rail {
      display: contents;
    }
    .panel.opponent.flipped .rail-gap {
      display: none;
    }
    .panel.opponent.flipped .rail > :global(.identity) {
      grid-area: rail;
      align-self: end;
    }
    .panel.opponent.flipped .rail > :global(.pile-bar) {
      grid-area: piles;
      grid-template-columns: repeat(3, 44px);
      width: auto;
      align-self: start;
    }
    /* A top-row panel is narrow: the two back-row zones may give up
       more of their width, and shrink their cards to fit (rowFit),
       before either crowds the piles. */
    .panel.opponent.flipped .grid-middle {
      gap: 12px;
    }
    /* Room for "Enchant / artifact 3" when the panel has it; under that
       the label ends in "…" rather than scrolling the row. */
    .panel.opponent.flipped .grid-middle > :global(.row) {
      min-width: min(124px, 44%);
    }
  }
  /* A narrow self panel (the quadrant's bottom right): the dock is
     wider than the room right of the back row, so the back row keeps
     its full width and the bottom row keeps the dock's height, as
     before. The creature card is then what is left above the dock:
     H = 1.5h + dock + ~64px. The panel is a size container, so these
     rules on its children read its width; the sizes are redeclared
     on them because a custom property's var() is fixed where it is
     declared. */
  @media (min-width: 600px) {
    @container (max-width: 1099px) {
      .panel.self.docked .dock-spacer {
        height: var(--dock-h, 0px);
      }
      .panel.self.docked .grid-middle {
        padding-right: 0;
      }
      .panel.self.docked :is(.grid-creatures, .grid-middle, .grid-bottom) {
        --card-h: clamp(
          100px,
          calc((100cqh - var(--dock-h, 170px) - 64px) / 1.5 * var(--card-scale, 1)),
          var(--card-h-max)
        );
        --card-w: calc(var(--card-h) * 5 / 7);
        --card-h-sm: clamp(64px, calc(var(--card-h) * 0.5), 110px);
        --card-w-sm: calc(var(--card-h-sm) * 5 / 7);
      }
    }
  }
  .flipped .rail :global(.pile .label) {
    display: none;
  }
</style>

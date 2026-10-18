<script lang="ts">
  // PlayerPanel is one player's full board, laid out in a CSS Grid
  // matching the S16.5 avatar-centric redesign:
  //
  //   ┌─────────────────────────────────────────────┐
  //   │  creatures (full width)                     │
  //   ├──────────────────────┬──────────────────────┤
  //   │  lands · planeswalkers · battles · artifacts · enchantments │
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
  import { graveyardHostLabels } from "../../seatSummary";
  import {
    bucketForBattlefield,
    isCreature,
    isLand,
    type BattlefieldBucket,
  } from "../../cardTypes";
  import {
    BACK_ZONES,
    ZONE_LABEL,
    ZONE_WEIGHT,
    backRowShares,
    isSpreadZone,
    type BackZone,
  } from "../../backRow";
  import {
    anyPlayerManaRows,
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
  import CompanionSlot from "./CompanionSlot.svelte";
  import type { CastSourceZone } from "../../targeting";
  import PlayerIdentity from "./PlayerIdentity.svelte";
  import Icon from "../Icon.svelte";
  import { seatColor } from "../../colors";
  import { settings } from "../../settings";
  import { playmatShownFor, playmatSrc } from "../../playmat";
  import PromisesRow from "./PromisesRow.svelte";
  import TokenGroupModal from "./TokenGroupModal.svelte";
  import { groupMembersOf } from "../../tokenGroups";
  import { L } from "../../labels";
  import { allocateRowHeights, type RowReport } from "../../rowFit";
  import {
    backRowSplit,
    handNotch,
    partBackRow,
    seatOverhang,
    type BackRowSplit,
    type HandNotch,
  } from "../../seatNotch";
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
    // #3266: the attackers the viewer has staged, drawn as attacking.
    stagedAttackers?: ReadonlySet<string>;
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
    // The piles (and so the back row) start right of the card too, since it
    // is taller than the hand row (#2990).
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
    //
    // ADR 0120 as amended 2026-10-18 (#4122): it is also laid out as a
    // reading surface rather than a table panel. A header names the seat
    // (its one identity, `PlayerIdentity` `header`), every battlefield
    // group draws at one card size capped at 240px, the command zone is
    // the back row's last group, and the piles and the hand sit in a
    // short strip along the bottom.
    expanded?: boolean;
    // ADR 0120 as amended 2026-10-18: the reflow turn chip (NEXT UP,
    // AFTER YOU, AFTER <NAME>) for a seat that is an avatar card, shown
    // in the expanded header. Null shows none; the table's panel never
    // shows it (its avatar card does).
    turnChip?: string | null;
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
    stagedAttackers,
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
    turnChip = null,
  }: Props = $props();

  // ADR 0120 §3: the expanded copy is upright, and leaves the dock's
  // and the coach card's cells to the table's panel.
  const flipped = $derived(flippedProp && !expanded);
  const docked = $derived(dockedProp && !expanded);
  const coached = $derived(coachedProp && !expanded);
  // ADR 0128: this seat's playmat, drawn behind its battlefield. The
  // per-device setting decides whose are drawn; a mat that fails to
  // load is dropped for that URL (no broken-image box), leaving the
  // plain board. Keyed by the wire URL, not the src, so a renewed
  // session token does not retry a mat that already failed.
  const playmatURL = $derived(playmatShownFor($settings.display.playmats, seat, isSelf));
  let playmatFailed = $state<string | null>(null);
  const playmat = $derived(
    playmatURL !== null && playmatFailed !== playmatURL ? playmatSrc(playmatURL) : null,
  );
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
  // "enchantments" row. Without a badge naming its victim the
  // board says nothing about who is being cursed, which is the whole
  // card — so each such permanent gets the enchanted seat's name.
  //
  // Drawing it in the ENCHANTED player's panel would read better
  // still, but it makes "where a card is drawn" diverge from
  // card.controller, and that is a new concept for this UI.
  //
  // ADR 0154 (#1780): an Aura enchanting a card in a graveyard (Animate
  // Dead while its enters trigger waits) has no host on the battlefield
  // either, and gets the same badge naming the card and whose graveyard
  // it is in.
  const curseTargets = $derived.by(() => {
    const names = new Map((view.seats ?? []).map((s) => [s.id, s.name]));
    const out: Record<string, string> = graveyardHostLabels(view.battlefield?.cards, view.seats);
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

  // #3028: the back row's zones share its width by how many piles each
  // draws. The land strip takes what the others leave (#2960), and at
  // full card size a long artifact row left it a sliver, so each other
  // zone is capped at its share. With no land piles they split the row.
  //
  // #4087: every permanent type has a zone of its own: planeswalkers,
  // battles, artifacts, enchantments, and "other" for a card whose type
  // line names none. A zone is drawn only while it holds a card, so a
  // count left by a row that has gone is ignored (backZones below).
  // backRowShares weighs a planeswalker or battle pile up, since it is
  // never overlapped.
  let landPiles = $state(0);
  let zonePiles = $state<Record<BackZone, number>>({
    planeswalker: 0,
    battle: 0,
    artifact: 0,
    enchantment: 0,
    other: 0,
  });
  const backZones = $derived(BACK_ZONES.filter((z) => buckets[z].length > 0));
  // Your back row parted around your avatar (backRowSplit), where the
  // two share height; null elsewhere.
  let backSplit = $state<BackRowSplit | null>(null);
  const pilesOf = (zones: readonly BackZone[]) =>
    Object.fromEntries(zones.map((z) => [z, zonePiles[z]])) as Partial<Record<BackZone, number>>;
  // ADR 0204 §3: where your back row parts around your avatar (backSplit,
  // below), its groups are shared between the part left of the avatar and
  // the part right of it by the room each has, and each part shares its
  // width among its own groups. Otherwise it is one part, as before.
  const backParts = $derived(
    backSplit
      ? partBackRow(
          landPiles,
          backZones.map((z) => ({ zone: z, weight: zonePiles[z] * ZONE_WEIGHT[z] })),
          backSplit.left,
          backSplit.zones,
        )
      : { left: backZones, right: [] as BackZone[] },
  );
  const sharesLeft = $derived(backRowShares(landPiles, pilesOf(backParts.left)));
  const sharesRight = $derived(backRowShares(0, pilesOf(backParts.right)));
  const shareOf = (z: BackZone): number =>
    backParts.right.includes(z) ? sharesRight[z] : sharesLeft[z];

  // #3044: the height the creature row and the back row share, and
  // what each asks of it (rowFit.ts rowDemand). A crowded row may wrap
  // into its budget instead of shrinking its cards toward the floor;
  // the budgets never add up to more than the panel holds, so a wrapped
  // row takes no height from the hand or the other row. The back row's
  // two zones sit side by side and share one budget.
  type Demand = RowReport | null;
  let creatureDemand = $state<Demand>(null);
  let landDemand = $state<Demand>(null);
  let zoneDemand = $state<Record<BackZone, Demand>>({
    planeswalker: null,
    battle: null,
    artifact: null,
    enchantment: null,
    other: null,
  });
  let fieldH = $state(0);
  const budgets = $derived.by(() => {
    if (fieldH <= 0 || !creatureDemand) return [0, 0];
    const back = [landDemand, ...backZones.map((z) => zoneDemand[z])].filter(
      (d): d is NonNullable<Demand> => !!d,
    );
    const backDemand = {
      base: Math.max(0, ...back.map((d) => d.base)),
      want: Math.max(0, ...back.map((d) => d.want)),
    };
    // ADR 0204: the band the row beside the hand keeps clear under the
    // avatar is height the two rows do not get.
    return allocateRowHeights(Math.max(0, fieldH - seatNotch), [creatureDemand, backDemand]);
  });

  // ADR 0204: what the seated identity keeps clear. `seatNotch` is the
  // depth of the band the battlefield row beside the hand keeps clear
  // under the avatar (call 3); `ownHandNotch` is where the viewer's
  // avatar sits over their own fan, so the cards under it dip (call 4).
  // Both are read off the laid-out boxes (seatNotch.ts decides), and are
  // zero and null where the identity is not seated (a phone's rail).
  let seatNotch = $state(0);
  let ownHandNotch = $state<HandNotch | null>(null);
  function sameSplit(a: BackRowSplit | null, b: BackRowSplit | null): boolean {
    return (
      a === b ||
      (!!a &&
        !!b &&
        a.left === b.left &&
        a.notch === b.notch &&
        a.zones === b.zones &&
        a.right === b.right)
    );
  }
  function sameNotch(a: HandNotch | null, b: HandNotch | null): boolean {
    return (
      a === b || (!!a && !!b && a.left === b.left && a.right === b.right && a.depth === b.depth)
    );
  }
  function seatRoom(node: HTMLElement, isFlipped: boolean) {
    if (typeof ResizeObserver === "undefined") return {};
    let flip = isFlipped;
    let frame = 0;
    const measure = () => {
      frame = 0;
      const id = node.querySelector<HTMLElement>(":scope > .rail > .identity");
      const bottom = node.querySelector<HTMLElement>(":scope > .grid-bottom");
      if (!id || !bottom || getComputedStyle(id).position !== "absolute") {
        if (seatNotch !== 0) seatNotch = 0;
        if (ownHandNotch) ownHandNotch = null;
        if (backSplit) backSplit = null;
        return;
      }
      const p = node.getBoundingClientRect();
      const ir = id.getBoundingClientRect();
      const disc = id.querySelector(".avatar-wrap")?.getBoundingClientRect();
      const life = id.querySelector(".life-row")?.getBoundingClientRect();
      const kw = id.querySelector(".seat-keywords")?.getBoundingClientRect();
      const lowest = Math.max(ir.bottom, life?.bottom ?? 0, kw?.bottom ?? 0);
      const br = bottom.getBoundingClientRect();
      // Every seat keeps the band its avatar reaches into clear, yours
      // included (call 3): the back row is laid out as it always was, so
      // its groups shrink-fit their width and nothing scrolls, and the
      // rows' height budget gives up the band. Only on a short narrow
      // board, whose back row already spills into the dock's row the
      // avatar stands in, can no band clear it; there your back row parts
      // around your avatar instead (backRowSplit).
      const docked = node.classList.contains("docked");
      const wide = docked && node.clientWidth - 16 >= 1100;
      const mid = node.querySelector<HTMLElement>(":scope > .grid-middle");
      const spacer = bottom
        .querySelector<HTMLElement>(":scope > .dock-spacer")
        ?.getBoundingClientRect();
      const rows = mid
        ? [...mid.querySelectorAll<HTMLElement>(".row")].map(
            (e) => e.getBoundingClientRect().bottom,
          )
        : [];
      const tight = Math.max(0, ...rows) > br.top + 1;
      const split = docked && !!mid && !!spacer && (wide || tight);
      const v = split
        ? 0
        : seatOverhang({
            identity: { top: ir.top - p.top, bottom: lowest - p.top },
            handRow: { top: br.top - p.top, bottom: br.bottom - p.top },
            rowGap: parseFloat(getComputedStyle(node).rowGap) || 0,
            handOnTop: flip,
          });
      if (v !== seatNotch) seatNotch = v;
      let nextSplit: BackRowSplit | null = null;
      if (split && mid && spacer) {
        const sides = [...id.querySelectorAll<HTMLElement>(".side")]
          .map((e) => e.getBoundingClientRect())
          .filter((r) => r.width > 0);
        const mr = mid.getBoundingClientRect();
        nextSplit = backRowSplit({
          panelW: p.width,
          start: mr.left - p.left + (parseFloat(getComputedStyle(mid).paddingLeft) || 0),
          avatar: {
            left: Math.min(ir.left, ...sides.map((r) => r.left)) - p.left,
            right: Math.max(ir.right, ...sides.map((r) => r.right)) - p.left,
          },
          dockLeft: spacer.left - p.left,
        });
      }
      if (!sameSplit(nextSplit, backSplit)) backSplit = nextSplit;
      const zone = bottom.querySelector<HTMLElement>(":scope > .hand-zone");
      const hand = zone?.querySelector<HTMLElement>(":scope > .hand");
      let next: HandNotch | null = null;
      if (isSelf && !flip && disc && zone && hand) {
        next = handNotch({
          avatar: {
            left: Math.min(disc.left, life?.left ?? disc.left),
            right: Math.max(disc.right, life?.right ?? disc.right),
          },
          avatarBottom: lowest,
          // A card at rest shows from the strip's top, under its padding.
          handTop: zone.getBoundingClientRect().top + 4,
          handLeft: hand.getBoundingClientRect().left,
        });
      }
      if (!sameNotch(next, ownHandNotch)) ownHandNotch = next;
    };
    const schedule = () => {
      if (!frame) frame = requestAnimationFrame(measure);
    };
    const ro = new ResizeObserver(schedule);
    ro.observe(node);
    for (const sel of [
      ":scope > .rail > .identity",
      ":scope > .grid-bottom",
      ":scope > .grid-bottom > .hand-zone",
      ":scope > .grid-middle",
    ]) {
      const el = node.querySelector(sel);
      if (el) ro.observe(el);
    }
    schedule();
    return {
      update(f: boolean) {
        flip = f;
        schedule();
      },
      destroy() {
        if (frame) cancelAnimationFrame(frame);
        ro.disconnect();
      },
    };
  }

  // fieldHeight measures the panel's height less its hand row (the
  // first grid track on an across-table panel, the last otherwise), its
  // padding and its row gaps: what the creature and back rows share. It
  // reads the panel and the hand row, never the battlefield rows, so a
  // row wrapping does not move it.
  function fieldHeight(node: HTMLElement, isFlipped: boolean) {
    if (typeof ResizeObserver === "undefined") return {};
    let flip = isFlipped;
    let frame = 0;
    const measure = () => {
      frame = 0;
      const cs = getComputedStyle(node);
      const tracks = cs.gridTemplateRows.split(/\s+/).map((t) => parseFloat(t) || 0);
      if (tracks.length < 3) return;
      const hand = flip ? tracks[0] : tracks[tracks.length - 1];
      // ADR 0120 as amended 2026-10-18: the expanded panel's first track
      // is its header, which the rows do not get either.
      const head = node.classList.contains("expanded") && tracks.length > 3 ? tracks[0] : 0;
      const gap = parseFloat(cs.rowGap) || 0;
      const pad = (parseFloat(cs.paddingTop) || 0) + (parseFloat(cs.paddingBottom) || 0);
      const h = node.clientHeight - pad - hand - head - gap * (tracks.length - 1);
      if (Math.abs(h - fieldH) > 0.5) fieldH = Math.max(0, h);
    };
    const schedule = () => {
      if (!frame) frame = requestAnimationFrame(measure);
    };
    const ro = new ResizeObserver(schedule);
    ro.observe(node);
    for (const sel of [":scope > .grid-bottom", ":scope > .head"]) {
      const el = node.querySelector(sel);
      if (el) ro.observe(el);
    }
    schedule();
    return {
      update(f: boolean) {
        flip = f;
        schedule();
      },
      destroy() {
        if (frame) cancelAnimationFrame(frame);
        ro.disconnect();
      },
    };
  }

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
    const out: Record<BattlefieldBucket, CardView[]> = {
      creature: [],
      land: [],
      planeswalker: [],
      battle: [],
      artifact: [],
      enchantment: [],
      other: [],
    };
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
  //
  // #1864 (ADR 0106 §1 amendment 2026-10-16): an opponent's panel gets
  // the handler too when a seated viewer looks at it, for the "Any player
  // may activate this ability" mana rows alone (Mana Cache). The viewer
  // activates them, so the action is the viewer's and the mana lands in
  // the viewer's pool; every other row of a permanent the viewer does
  // not control is ignored here, as the server would refuse it.
  function sendManaActivation(
    card: CardView,
    ability: ManaAbilityView | undefined,
    abilityIndex: number,
    player: string,
    colors?: string[],
    phyrexianLife?: number,
  ): void {
    if (ability && onManaAbilityCost && manaAbilityNeedsPrompt(ability)) {
      onManaAbilityCost(card, ability, colors);
      return;
    }
    // #1443: a colour chosen at the card rides the activation, so the
    // server produces it with no second question.
    sendAction(
      "activate_mana_ability",
      {
        card_id: card.instance_id,
        ability_index: abilityIndex,
        // ADR 0093: the row this click meant, so a stale one is refused.
        ...manaAbilityRef(card, abilityIndex),
        ...manaColorParams(colors),
        // ADR 0131 §2: the "Pay life for {B}…" row's claim.
        ...(phyrexianLife ? { phyrexian_life: phyrexianLife } : {}),
      },
      player,
    );
  }
  const activateManaAbility = $derived(
    isSelf
      ? (card: CardView, abilityIndex: number, colors?: string[], phyrexianLife?: number) => {
          // #1228: a permanent publishes `mana_abilities` and a card
          // in hand whose mana ability functions there publishes
          // `zone_mana_abilities` — never both. One lookup reads
          // whichever is present, exactly as the card menu does.
          const rows = card.mana_abilities ?? card.zone_mana_abilities ?? [];
          const ability = rows.find((a) => a.index === abilityIndex);
          sendManaActivation(card, ability, abilityIndex, seat.id, colors, phyrexianLife);
        }
      : viewerID && !spectator
        ? (card: CardView, abilityIndex: number, colors?: string[], phyrexianLife?: number) => {
            const ability = anyPlayerManaRows(card).find((a) => a.index === abilityIndex);
            if (!ability) return;
            sendManaActivation(card, ability, abilityIndex, viewerID, colors, phyrexianLife);
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
    // #1864: on a permanent the viewer does not control, only its
    // any-player mana rows (Mana Cache) are the viewer's to click.
    const own = isSelf || (card.controller || card.owner) === viewerID;
    const plan = manaClickPlan(own ? card : { ...card, mana_abilities: anyPlayerManaRows(card) }, {
      payerLife,
    });
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
  class:coached
  class:expanded
  class:seat-turn={isActive}
  style:--seat-color={seatColor(seat.seat)}
  style:--creature-min={creatureDemand?.used ? `${creatureDemand.used}px` : undefined}
  style:--seat-notch={seatNotch ? `${seatNotch}px` : undefined}
  class:back-split={backSplit !== null}
  style:--seat-split-left={backSplit ? `${backSplit.left}px` : undefined}
  style:--seat-split-notch={backSplit ? `${backSplit.notch}px` : undefined}
  style:--seat-split-right={backSplit ? `${backSplit.right}px` : undefined}
  role={expanded ? "group" : "region"}
  aria-label={expanded ? undefined : isSelf ? L.yourBoard : L.seatBoard(seat.name)}
  use:fieldHeight={flipped}
  use:seatRoom={flipped}
>
  {#if playmat}
    <!-- ADR 0128: the owner's playmat, under everything. An <img> and
         not a CSS background so a failed load is detectable. -->
    <div
      class="playmat"
      aria-hidden="true"
      style:--playmat-wash={seat.playmat_wash ? `${seat.playmat_wash}%` : undefined}
    >
      <!-- #3934: the whole mat, contained; the strips it leaves are
           this same image (one request, same src) blurred and dimmed. -->
      <img
        class="backdrop"
        src={playmat}
        alt=""
        draggable="false"
        decoding="async"
        referrerpolicy="no-referrer"
        onerror={() => (playmatFailed = playmatURL)}
      />
      <img
        class="mat"
        src={playmat}
        alt=""
        draggable="false"
        decoding="async"
        referrerpolicy="no-referrer"
        onerror={() => (playmatFailed = playmatURL)}
      />
    </div>
  {/if}
  {#if expanded}
    <!-- ADR 0120 as amended 2026-10-18 (#4122): the header names the
         seat. It holds the seat's one identity (call 12), so the
         avatar's target, attack and peek handlers and its data-seat-id
         are here, and the reflow turn chip when the seat has one. The
         overlay's pin sits over its right end (Board.svelte). Not a
         heading: the overlay region already carries the name. -->
    <div class="head">
      {@render identity(true)}
      {#if turnChip && !seat.eliminated}
        <span class="turn-chip" class:next={turnChip === L.reflowNextUp} title={turnChip}
          >{turnChip}</span
        >
      {/if}
    </div>
  {/if}
  <div class="grid-creatures">
    <BattlefieldRow
      label={L.creatures}
      {attachmentsByHost}
      {curseTargets}
      {takenFrom}
      {ringBearers}
      {cantAttack}
      cards={buckets.creature}
      budget={budgets[0]}
      bind:demand={creatureDemand}
      hold
      {viewerID}
      {selectedCombatCardID}
      {stagedAttackers}
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
    <!-- ADR 0204 §3: the back row's two parts either side of your
         avatar's notch, where it parts (`back-split`); otherwise
         `display: contents`, one row as before. -->
    <div class="back-part back-left">
      <BattlefieldRow
        bind:pileCount={landPiles}
        label={L.lands}
        {attachmentsByHost}
        {curseTargets}
        {takenFrom}
        {ringBearers}
        {cantAttack}
        cards={buckets.land}
        budget={budgets[1]}
        bind:demand={landDemand}
        compact
        strip
        {viewerID}
        {selectedCombatCardID}
        {stagedAttackers}
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
      {#each backParts.left as z (z)}
        {@render backZone(z)}
      {/each}
    </div>
    <div class="back-part back-right">
      {#each backParts.right as z (z)}
        {@render backZone(z)}
      {/each}
    </div>
    {#if expanded && !isSelf}
      <!-- ADR 0120 as amended 2026-10-18 (call 10): in the overlay the
           command zone is the back row's last group, at the one card
           size, with its revealed companion; not tucked by the hand. -->
      <CommandStrip {seat} row />
      <CompanionSlot {seat} />
    {/if}
  </div>
  {#snippet backZone(z: BackZone)}
    <!-- #4087: one zone per permanent type, drawn only while it holds
           a card. Planeswalkers and battles never overlap or fold into
           a pile (BattlefieldRow `spread`), so two walkers cannot hide
           each other's loyalty. -->
    <BattlefieldRow
      bind:pileCount={zonePiles[z]}
      label={ZONE_LABEL[z]}
      {attachmentsByHost}
      {curseTargets}
      {takenFrom}
      {ringBearers}
      {cantAttack}
      cards={buckets[z]}
      budget={budgets[1]}
      bind:demand={zoneDemand[z]}
      share={shareOf(z)}
      compact
      spread={isSpreadZone(z)}
      {viewerID}
      {selectedCombatCardID}
      {stagedAttackers}
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
  {/snippet}
  <!-- The seat's identity: seated at its hand's edge on the table
       (ADR 0204), or in the expanded overlay's header (ADR 0120 as
       amended 2026-10-18). Drawn once either way. -->
  {#snippet identity(header: boolean)}
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
      seated={!header}
      {header}
      handCount={!header && !isSelf && $settings.display.showOpponentHandCount
        ? seat.hand.count
        : null}
    />
  {/snippet}
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
    <!-- data-hand-owner: where a spell cast from this hand flies from
         (#3381, castFlight.ts). -->
    <div class="hand-zone" data-hand-owner={seat.id}>
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
        notch={isSelf ? ownHandNotch : null}
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
      <!-- ADR 0172 §8: your companion, outside the game, with the {3}
           special action that puts it into your hand. Renders nothing
           without one. -->
      <CompanionSlot
        {seat}
        onSpecialAction={sendSpecialAction}
        {legal}
        legalGate={rowGate}
        {sorcerySpeedBlocked}
      />
    {/if}
    {#if !isSelf}
      {#if expanded}
        {#if $settings.display.showOpponentHandCount}
          <!-- ADR 0120 as amended 2026-10-18: no avatar rides the fan in
               the overlay's strip, so the count sits beside it. -->
          <span
            class="strip-hand-count"
            title={`${seat.hand.count} ${seat.hand.count === 1 ? "card" : "cards"} in hand`}
            aria-label={`${seat.hand.count} ${seat.hand.count === 1 ? "card" : "cards"} in hand`}
            ><Icon name="hand" size={11} />{seat.hand.count}</span
          >
        {/if}
      {:else}
        <!-- #2349: this seat's commander, beside its hand, where yours
             is beside yours. -->
        <CommandStrip {seat} />
        <!-- ADR 0172 §8: and its revealed companion, view only. -->
        <CompanionSlot {seat} />
      {/if}
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
    {#if !expanded}
      {@render identity(false)}
    {/if}
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
       (120×168 / 88×123 / 64×90). Your own board has no fixed
       ceiling (#3028): a fixed 240px left a third of a 1440p panel as
       empty mat, and made the "large" card size a no-op there. The
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
       next time either changes. On your own board it is the panel's
       own height (#3028), which no card reaches: the slope is the
       limit. Opponents' ceilings below grow with the screen instead. */
    --card-h-max: 100cqh;
    --card-h: clamp(168px, calc((42cqh - 30px) * var(--card-scale, 1)), var(--card-h-max));
    --card-w: calc(var(--card-h) * 5 / 7);
    --card-h-sm: max(123px, calc(var(--card-h) * 0.74));
    --card-w-sm: calc(var(--card-h-sm) * 5 / 7);
    --thumb-w: calc(40px * var(--card-scale, 1));
    --thumb-h: calc(56px * var(--card-scale, 1));
    --avatar-size-base: 84px;
    --rail-w: 112px;
    display: grid;
    grid-template-columns: minmax(0, 1fr) var(--rail-w);
    /* #3045: the creature row takes its own height (less on a short
       panel, where it scrolls), the back row sits right under it, and the
       panel's spare height opens between the back row and the hand,
       at the panel's outer edge. A row shrunk to fit its width (rowFit)
       used to leave that spare as a band at the table's centre line.
       #3044: a creature row wrapped into its budget reports its lines'
       height as --creature-min, since an auto track beside a 1fr one is
       not promised its content and the row would scroll instead.
       #3893: the back row's track is a plain 1fr, kept at least its
       content's height by .grid-middle's min-height. A min-content
       minimum also took the height of every item spanning the track:
       the rail's identity, as tall as both rows, so on a short panel the
       creature row gave its height to empty space under the back row,
       and an empty creature row (one label high) gave it all. */
    grid-template-rows: minmax(var(--creature-min, 0px), auto) 1fr auto;
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
    /* The playmat sits at z-index -1, under the zones, so the panel must
       be its own stacking context. container-type: size already makes it
       one (layout containment); this says so where it is relied on. */
    isolation: isolate;
  }
  /* ADR 0128: the seat owner's playmat, a background layer and nothing
     else. It takes no space in the grid (absolute), takes no pointer
     events, and paints under every zone. `contain` and centred over a blurred copy of itself (#3934), so
     none of it is cropped. The scrim is theme-aware: it is the page's own
     background colour at partial opacity, so in a dark skin the mat is
     dimmed toward dark and in the light skin toward light, and the
     cards, chips and text above it keep their contrast either way. */
  .playmat {
    position: absolute;
    inset: 0;
    z-index: -1;
    pointer-events: none;
    overflow: hidden;
    border-radius: inherit;
  }
  .playmat img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: contain;
    object-position: center;
    user-select: none;
  }
  /* #3934: the leftover strips. Same image, filled (cover), blurred and
     darkened; scaled up a little so the blur's soft edge falls outside
     the clip. One static filter, no script. */
  .playmat img.backdrop {
    position: absolute;
    inset: 0;
    object-fit: cover;
    filter: blur(18px) brightness(0.7);
    transform: scale(1.08);
  }
  .playmat img.mat {
    position: relative;
  }
  .playmat::after {
    content: "";
    position: absolute;
    inset: 0;
    /* The owner sets the strength (ADR 0128 amendment); 58% when they
       have not, or the server sent none. */
    background: color-mix(in srgb, var(--bg) var(--playmat-wash, 58%), transparent);
  }
  /* An across-table opponent's board is yours turned 180° about the
     table's centre (#2438, below), so their mat is turned with it: the
     art faces its owner, as it would on paper. Only where that turn
     applies (the same 600px floor as the layout). */
  @media (min-width: 600px) {
    .panel.opponent.flipped .playmat img {
      transform: rotate(180deg);
    }
    .panel.opponent.flipped .playmat img.backdrop {
      transform: rotate(180deg) scale(1.08);
    }
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
       (settings) with a gentler curve than self. The ceiling is
       deliberate: an opponent panel shares the screen with the
       viewer's own and shouldn't rival it. #3028: it is 200px on a
       1080px-tall screen and grows with the screen past that (18.5vh),
       and the card-size setting scales it, so "large" is larger. */
    --card-h-max: calc(max(200px, 18.5vh) * var(--card-scale-opponent, 1));
    /* #2336: sized by relevance, like your own board. The face-down
       hand shows only a count, so it peeks 30%; lands and other
       permanents are half a creature. The creature card is what is
       left: H = 1.92h + ~64px (the commander beside the hand, CommandStrip, is 0.42h), so h ≈ 52cqh - 33px. The floor is low so
       a short panel shrinks its cards rather than scrolling them. */
    --hand-peek: 0.3;
    --card-h: clamp(90px, calc((52cqh - 33px) * var(--card-scale-opponent, 1)), var(--card-h-max));
    --card-h-sm: max(56px, calc(var(--card-h) * 0.5));
    --thumb-w: calc(36px * var(--card-scale-opponent, 1));
    --thumb-h: calc(50px * var(--card-scale-opponent, 1));
    --avatar-size-base: 72px;
    --rail-w: 104px;
  }
  /* #1071: a spectator has no panel of their own for an opponent
     panel to defer to, so the reason for the ceiling above doesn't
     apply — every seat is equal, and the ceiling matches self's: none
     but the panel's own height (#3028). This rule and .panel.opponent live in the same
     stylesheet (same Svelte scope hash on both), so there is no
     cross-component cascade tie to reason about — one extra class
     over .panel.opponent is ordinary CSS specificity. The floor and
     slope are untouched, so nothing changes until the opponent
     ceiling binds. Board never sets `flipped` on a spectator panel today
     (every spectator seat mounts isSelf=false with no `flipped`), so
     .panel.opponent.flipped below never competes with this rule in
     practice — the dedicated four-class rule after it is what keeps
     that true even if a spectator panel ever does gain `flipped`. */
  .panel.opponent.spectator {
    --card-h-max: 100cqh;
  }
  .panel.opponent.flipped {
    /* Across-table seats: your board turned 180° about the table's
       centre (#2438, #3028). Rows reversed so the hand hugs the top
       edge and creatures face the centre of the table; #2491: the top
       row is the tucked commander, half a card.
       #3028: the same size as yours, by the same budget as
       .panel.self.docked below: creatures at 52cqh - 40px with no
       fixed ceiling, the back row half a card, the hand peeking 40%,
       and pile tiles no taller than the peek. In a two-player game
       the two halves of the table are equal (Board.svelte), so the
       two boards are the same size. The floor stays low (44px), as
       before, so a short top-row panel with three or four players
       shrinks its cards rather than scrolling them. */
    --card-h-max: 100cqh;
    --card-h: clamp(
      44px,
      calc(min(52cqh - 40px, (100cqh - 151px) / 1.5) * var(--card-scale-opponent, 1)),
      var(--card-h-max)
    );
    --card-h-sm: max(24px, calc(var(--card-h) * 0.5));
    --hand-peek: 0.4;
    --thumb-h: min(
      calc(56px * var(--card-scale-opponent, 1)),
      max(36px, calc(var(--card-h) * var(--hand-peek) - 30px))
    );
    --thumb-w: calc(var(--thumb-h) * 5 / 7);
    --avatar-size-base: 84px;
    --rail-w: 112px;
    grid-template-rows: auto 1fr minmax(var(--creature-min, 0px), auto);
    grid-template-areas:
      "bottom    rail"
      "middle    rail"
      "creatures rail";
  }
  /* Defensive only — unreachable today (see above), kept so a future
     spectator panel that gains `flipped` still gets the raised
     ceiling instead of silently falling back to the flipped one. Four classes
     beats both .panel.opponent.flipped (three) and
     .panel.opponent.spectator (three) outright, so this can never
     lose to either regardless of source order. */
  .panel.opponent.flipped.spectator {
    --card-h-max: 100cqh;
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
  /* #3045: the back row hugs the creature row, the way yours does turned
     180°; the spare height sits between it and the hand. */
  .flipped .grid-middle {
    align-items: flex-end;
  }
  .back-part {
    display: contents;
  }
  .grid-middle {
    grid-area: middle;
    /* #3893: the back row's track never takes less than the back row
       (the track is a plain 1fr; see .panel). */
    min-height: min-content;
    min-width: 0;
    display: flex;
    /* safe: a back row wider than its area starts at its left edge
       rather than spilling left over the piles beside it (#2438). */
    justify-content: safe center;
    align-items: flex-start;
    gap: 14px;
    /* #4087: with a zone per permanent type a narrow panel (a phone, a
       four-seat across row) cannot hold them side by side. They scroll
       sideways inside the back row, as a crowded row always has, rather
       than spilling out under the player rail. */
    overflow-x: auto;
    overflow-y: hidden;
  }
  /* Each back-row zone takes the width of its cards, no more, so the
     two read as one centred group. An empty zone keeps enough width
     to show its label. */
  .grid-middle > :global(.row),
  .back-part > :global(.row) {
    flex: 0 1 auto;
    min-width: 112px;
    max-width: 100%;
  }
  /* #4087: with a zone per permanent type the back row can hold up to
     six, so a zone other than the lands asks for just enough to show
     its label ("enchantments" and its count). */
  .grid-middle > :global(.row:not(.strip)),
  .back-part > :global(.row:not(.strip)) {
    min-width: 72px;
  }
  /* A phone-wide panel has about 245px beside the player rail: let each
     zone ask for less than a card and a half, so two or three show at
     once and the rest scroll, instead of the lands taking all of it. */
  @container (max-width: 480px) {
    .grid-middle > :global(.row),
    .back-part > :global(.row),
    .grid-middle > :global(.row:not(.strip)),
    .back-part > :global(.row:not(.strip)) {
      min-width: 88px;
    }
  }
  /* #2960: the land strip takes the room the other back-row zone leaves
     instead of shrink-wrapping its overlapped cards. fitRow then opens
     the strip (--strip-k) until the lands sit edge to edge, and
     overlaps them only when they no longer fit. */
  .grid-middle > :global(.row.strip),
  .back-part > :global(.row.strip) {
    flex: 1 1 0;
    /* #4090: but no wider than its lands laid edge to edge (fitRow
       publishes --strip-spread), so the zones beside it sit against it
       rather than across an empty strip. */
    max-width: var(--strip-spread, 100%);
    /* Room for "Lands 1" and a card, not a card and a half: the zones
       beside it are what the spare width is for. */
    min-width: 64px;
  }
  /* #3028: a zone takes no more than its share of the row by pile
     count (--row-share, set on the row from backRow.ts), so a long
     artifact row shrinks its own cards instead of crushing the lands.
     #4087: that holds for each of the zones, one per permanent type. */
  .grid-middle > :global(.row:not(.strip)),
  .back-part > :global(.row:not(.strip)) {
    max-width: calc(var(--row-share, 1) * 100%);
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
      /* #3028: on a short panel (two players at 720px, each half
         ~320px) the rows' fixed parts dominate: the back row is half a
         card plus ~32px, the hand row is at least a pile tile (~65px)
         and the creature row adds ~24px, so H = 1.5h + ~151px and the
         card is the smaller of the two. The floor is as low as an
         across-table seat's, so the cards shrink rather than clip. */
      --card-h: clamp(
        44px,
        calc(min(52cqh - 40px, (100cqh - 151px) / 1.5) * var(--card-scale, 1)),
        var(--card-h-max)
      );
      --card-h-sm: max(64px, calc(var(--card-h) * 0.5));
      --hand-peek: 0.4;
      /* #3028: a pile tile is its thumbnail plus ~30px of padding and
         count, and in the hand's row it must not stand taller than the
         hand's peek, or the row grows and the creatures lose height. */
      --thumb-h: min(
        calc(56px * var(--card-scale, 1)),
        max(36px, calc(var(--card-h) * var(--hand-peek) - 30px))
      );
      --thumb-w: calc(var(--thumb-h) * 5 / 7);
      /* #3028: the piles sit in the hand's row, against the bottom
         edge like the hand, as an across-table opponent's sit in theirs
         against the top edge (#2491), so the two boards mirror.
         ADR 0204 §3: there is no rail; your identity sits over the
         middle of your hand (below), and the battlefield takes the rail's
         width back. */
      grid-template-columns: auto minmax(0, 1fr);
      grid-template-areas:
        "creatures creatures"
        "middle    middle"
        "piles     bottom";
    }
    /* The bottom row is the hand's peek, not the dock's height: the
       dock rises over the back row's right end instead, which the back
       row keeps clear with padding. That height goes to the creatures. */
    .panel.self.docked .dock-spacer {
      height: 0;
    }
    .panel.self.docked .grid-middle {
      padding-right: var(--dock-w, 0px);
    }
    /* #2990, #2409, #3028: the coach card sits in the bottom-left
       corner, where the piles are, and is taller than the hand row, so
       it rises over the back row's left end, where the first lands
       land. While it shows, the piles start right of the card's width,
       and the hand after them, so the bottom row needs no empty cell of
       its own; the back row starts right of the card too. Neither takes
       any height, so the rows keep theirs. The creature row above is
       centred, with an empty left margin the card may rise into. */
    .panel.self.docked .coach-spacer {
      flex-basis: 0;
    }
    .panel.self.docked.coached .rail > :global(.pile-bar) {
      margin-left: calc(var(--coach-w, 0px) + 6px);
    }
    .panel.self.docked.coached .grid-middle {
      padding-left: calc(var(--coach-w, 0px) + 6px);
    }
  }
  /* #2438, #2483: an across-table opponent sits across from you. Their
     board is yours turned 180° about the table's centre, with the cards
     themselves left upright to read: their hand along the top edge with
     their commander on its left (yours is on your hand's right), the
     back row under it reading enchantments/artifacts then lands, the
     piles in the top right corner (yours are bottom left),
     and the creatures facing yours, with their identity at the bottom
     of the rail on the left, where yours is at the top of your rail on
     the right. The rail is `display: contents`, as on yours, so the
     piles leave it and the creatures keep its full height. */
  @media (min-width: 600px) {
    .panel.opponent.flipped {
      grid-template-columns: minmax(0, 1fr) auto;
      grid-template-rows: auto 1fr minmax(var(--creature-min, 0px), auto);
      /* #2491: the piles sit in the hand's row, against the top edge
         like the hand, not in the back row under it. ADR 0204 §2: the
         identity sits over the middle of the hand (below), so the rail's
         column is the battlefield's again. */
      grid-template-areas:
        "bottom    piles"
        "middle    middle"
        "creatures creatures";
    }
    /* #2491, #3028: their commander is tucked half behind the top edge,
       as yours is behind the bottom one, so their board mirrors yours:
       a hand-size card whose middle sits on the panel's clipping edge
       (the strip's top, less the panel's 8px padding). #2995: what
       shows is a card you can recognise, not a sliver, so the card is
       at least 168px and its visible half at least 84px. The hand row
       is only a peek tall, so the half that shows hangs down out of it
       into the empty corner above the identity rather than growing the
       row; the strip keeps the hand's height and lets its card
       overflow. The tax goes to the edge you can see. */
    .panel.opponent.flipped :global(.command-strip) {
      /* ADR 0204 §5: and no wider than the room left of the avatar, less
         the hand count on the portrait's rim. */
      --strip-h: min(
        max(168px, var(--card-h)),
        (var(--seat-side) - var(--seat-badge, 24px) - 4px) * 7 / 5
      );
      height: calc(var(--card-h) * var(--hand-peek, 0.55));
      align-self: flex-start;
      align-items: flex-start;
      position: relative;
      z-index: 6;
      padding: 0;
    }
    .panel.opponent.flipped :global(.command-strip .slot) {
      flex: none;
      margin-top: calc(var(--strip-h) * -0.5 - 8px);
    }
    .panel.opponent.flipped :global(.command-strip .tax-badge) {
      top: auto;
      bottom: 3px;
    }
    /* #4087: the back row is mirrored with `direction: rtl`, not
       row-reverse. A reversed flex line that overflows spills past its
       far edge, where a scroll container cannot reach it, so a crowded
       back row would lose zones for good. In rtl the row starts at the
       right (the lands, then the walkers) and what does not fit spills
       left, into scroll range: the enchantments and artifacts are the
       ones that scroll away. The zones read left to right as ever. */
    .panel.opponent.flipped .grid-middle {
      direction: rtl;
    }
    .panel.opponent.flipped .grid-middle > :global(.row),
    .panel.opponent.flipped .back-part > :global(.row) {
      direction: ltr;
    }
    .panel.opponent.flipped .grid-bottom {
      flex-direction: row-reverse;
    }
    /* #2486: and what they hold faces them, as it would across a real
       table. Their hand, their piles' top cards and their commander are
       turned 180°, so the fan arcs down toward you and each card's top
       points at the table's centre. What you read stays upright: the
       pile counts, the commander's tax, the row labels and their
       identity. */
    .panel.opponent.flipped .hand-zone > :global(.hand),
    .panel.opponent.flipped .rail :global(.pile .thumb),
    .panel.opponent.flipped :global(.command-strip .slot .card) {
      transform: rotate(180deg);
    }
    /* #3028: their permanents face them too, your board turned 180°.
       Each row's card list turns about its own centre, so every card
       is upside down to you and the row runs from your right to your
       left, as theirs would across a real table. Layout is untouched
       (rowFit measures offsets, not the turned box), and the expanded
       overlay, which is never flipped, reads upright. */
    .panel.opponent.flipped :is(.grid-creatures, .grid-middle) :global(.row-cards) {
      transform: rotate(180deg);
    }
    /* #2995: but the empty-hand label is text, and reads upright. */
    .panel.opponent.flipped .hand-zone > :global(.hand > .empty) {
      transform: rotate(180deg);
    }
    .panel.opponent.flipped .rail > :global(.pile-bar) {
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
    .panel.opponent.flipped .grid-middle > :global(.row),
    .panel.opponent.flipped .back-part > :global(.row) {
      min-width: min(88px, 44%);
    }
    /* #4087: up to five other zones sit beside the lands, so they ask
       for no more than their label ("enchantments" ends in "…"). */
    .panel.opponent.flipped .grid-middle > :global(.row:not(.strip)),
    .panel.opponent.flipped .back-part > :global(.row:not(.strip)) {
      min-width: 72px;
    }
    .panel.opponent.flipped .grid-middle > :global(.row.strip),
    .panel.opponent.flipped .back-part > :global(.row.strip) {
      min-width: 64px;
    }
  }
  /* ADR 0204: seats at the table. From 600px up every panel's identity
     sits at the middle of the edge its hand is on, in front of the fan,
     as a player holds their cards: a top-row opponent's at the top centre
     (§2), an upright opponent's at the bottom centre (§1), yours just above
     the middle of your hand (§3). The rail is gone: it is `display:
     contents`, its identity is placed over the panel, and its piles sit at
     the far end of the hand's edge (an upright panel's bottom left, as
     yours are; a top-row panel's top right). A phone keeps the rail. */
  @media (min-width: 600px) {
    .panel:not(.flipped):not(.docked) {
      grid-template-columns: auto minmax(0, 1fr);
      grid-template-areas:
        "creatures creatures"
        "middle    middle"
        "piles     bottom";
    }
    .rail {
      display: contents;
    }
    .rail-gap {
      display: none;
    }
    /* §5: the room either side of the avatar on the hand's edge, for
       the commander strip and the piles. They shrink to it rather than
       slide under the avatar: the piles' tiles narrow (here) and the
       commander card gets smaller (below). A container unit, so it is
       read against this panel wherever it is used. */
    .panel {
      --seat-side: calc(50cqw - var(--avatar-size-base) * var(--card-scale, 1) / 2 - 14px);
    }
    .rail > :global(.pile-bar) {
      --pile-col: clamp(24px, (var(--seat-side) - 8px) / 3, 52px);
      grid-area: piles;
      grid-template-columns: repeat(3, var(--pile-col));
      width: auto;
      align-self: end;
    }
    /* §2: the hand count sits on the commander's side of the portrait,
       away from the wider piles: the left on a top-row panel, the right
       on an upright one (its piles are at the bottom left). */
    .panel:not(.flipped) .rail > :global(.identity .hand-count) {
      right: auto;
      left: calc(50% + 24px);
    }
    /* An upright opponent's commander, at its hand's right end, no wider
       than the room right of the avatar either. */
    .panel.opponent:not(.flipped) :global(.command-strip) {
      --strip-h: min(
        calc(var(--card-h) * 0.42),
        (var(--seat-side) - var(--seat-badge, 24px) - 4px) * 7 / 5
      );
    }
    /* The promise tokens leave the hand's edge, which the avatar now
       holds the middle of: they hang just off it on the battlefield's
       side, at the commander's end (z over the back row's corner). */
    .panel.opponent .grid-bottom {
      position: relative;
    }
    .panel.opponent .grid-bottom > :global(.promises) {
      position: absolute;
      right: 0;
      bottom: calc(100% + 4px);
      z-index: 4;
    }
    .panel.opponent.flipped .grid-bottom > :global(.promises) {
      right: auto;
      left: 0;
      bottom: auto;
      top: calc(100% + 4px);
    }
    /* The tile's title and aria-label still carry the pile's name, so
       the piles stay one short row. Each thumb fits its column, in
       proportion, however narrow the room beside the avatar is. */
    .rail :global(.pile .label) {
      display: none;
    }
    .rail > :global(.pile-bar .thumb) {
      --pile-thumb-w: min(var(--thumb-w), calc(var(--pile-col) - 8px));
      --pile-thumb-h: min(var(--thumb-h), calc((var(--pile-col) - 8px) * 7 / 5));
      width: var(--pile-thumb-w);
      height: var(--pile-thumb-h);
      --card-w: var(--pile-thumb-w);
      --card-h: var(--pile-thumb-h);
    }
    /* The identity is centred on the panel, over the fan (z above the
       hand's 1 and the commander strip's 6). An opponent's sits against
       its hand's edge and overhangs toward the battlefield. */
    .rail > :global(.identity) {
      position: absolute;
      left: 50%;
      translate: -50% 0;
      z-index: 7;
    }
    .panel:not(.flipped) .rail > :global(.identity) {
      bottom: 4px;
    }
    .panel.flipped .rail > :global(.identity) {
      top: 2px;
    }
    /* §3: yours is centred over your hand's edge, which ends at the
       dock (ADR 0111 owns the corner), so it is never under the dock. It
       sits under the hand's hover lift (z 20), which shows whole cards
       over it. On a narrow board (below) the bottom row is the dock's
       height and the avatar stands in it, just above the hand and the
       piles; on a wide one it straddles the line between the battlefield
       and the hand. */
    .panel.self.docked .rail > :global(.identity) {
      /* Its steppers (about 130px right of its centre) end short of the
         dock too. */
      left: clamp(
        56px,
        calc((100% - var(--dock-w, 0px)) / 2),
        calc(100% - var(--dock-w, 0px) - 130px)
      );
      bottom: calc(8px + var(--card-h) * var(--hand-peek, 0.4) + 14px);
      z-index: 5;
    }
    /* §2, §3 (call 3): the battlefield row beside the hand keeps clear
       the depth the identity reaches into it (--seat-notch, measured),
       so no permanent sits under the avatar. The row hugs the far side
       of its area, so the band is the spare height first and costs the
       rows only what that does not cover (the row budgets know). */
    .panel.flipped .grid-middle {
      padding-top: var(--seat-notch, 0px);
    }
    .panel:not(.flipped) .grid-middle {
      padding-bottom: var(--seat-notch, 0px);
    }
  }
  /* ADR 0204 §5: a panel too narrow for the hand count beside the
     portrait as well (three opponents in a row on a tablet). The count
     rides on the portrait over its life chip instead, so the commander
     has all the room left of the avatar; the avatar keeps its centre and
     nothing goes under it. */
  @media (min-width: 600px) {
    @container (max-width: 319px) {
      .panel.flipped .rail > :global(.identity .hand-count),
      .panel:not(.flipped) .rail > :global(.identity .hand-count) {
        left: 50%;
        right: auto;
        bottom: 22px;
        translate: -50% 0;
      }
      .grid-bottom {
        --seat-badge: 0px;
      }
    }
  }
  /* ADR 0204 §3: a board of your own too narrow to seat you left of the
     dock (the quadrant's bottom right on a tablet): you stand just above
     the dock, at the middle of your board, and the back row keeps the
     band you reach into clear (--seat-notch). */
  @media (min-width: 600px) {
    @container (max-width: 489px) {
      .panel.self.docked .rail > :global(.identity) {
        left: 50%;
        bottom: calc(8px + var(--dock-h, 170px) + 8px);
      }
    }
  }
  /* ADR 0204 §3: on a wide board of your own the bottom row is only the
     hand's peek and the dock rises over the back row's right end. Your
     avatar straddles the line between the battlefield and the hand by
     --seat-dip, and the cards under it dip below it (Hand's notch, call
     4). */
  @media (min-width: 600px) {
    @container (min-width: 1100px) {
      .panel.self.docked .rail > :global(.identity) {
        --seat-dip: 10px;
        bottom: calc(8px + var(--card-h) * var(--hand-peek, 0.4) - var(--seat-dip));
      }
    }
    /* Call 3: on a short narrow board of your own the back row already
       spills into the dock's row your avatar stands in, so no band can
       clear it. There it parts around the avatar instead (backRowSplit,
       measured): a part left of the avatar, a clear notch the width of
       the avatar and its steppers, and a part right of it up to the dock.
       The groups are shared between the parts by the room each has
       (partBackRow), and each part shrink-fits its own, as the one row
       does elsewhere. */
    .panel.self.docked.back-split .grid-middle {
      display: grid;
      grid-template-columns: var(--seat-split-left) var(--seat-split-notch) minmax(0, 1fr);
      column-gap: 0;
      align-items: start;
      padding-right: var(--seat-split-right);
    }
    .panel.self.docked.back-split .back-part {
      grid-row: 1;
      display: flex;
      align-items: flex-start;
      gap: 14px;
      min-width: 0;
    }
    .panel.self.docked.back-split .back-left {
      grid-column: 1;
      justify-content: flex-end;
    }
    .panel.self.docked.back-split .back-right {
      grid-column: 3;
      justify-content: flex-start;
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
      .panel.self.docked :is(.grid-creatures, .grid-middle, .grid-bottom, .rail) {
        --card-h: clamp(
          100px,
          calc((100cqh - var(--dock-h, 170px) - 64px) / 1.5 * var(--card-scale, 1)),
          var(--card-h-max)
        );
        --card-w: calc(var(--card-h) * 5 / 7);
        --card-h-sm: max(64px, calc(var(--card-h) * 0.5));
        --card-w-sm: calc(var(--card-h-sm) * 5 / 7);
      }
    }
  }
  .flipped .rail :global(.pile .label) {
    display: none;
  }

  /* ADR 0120 as amended 2026-10-18 (#4122): the expanded overlay's copy
     of a board is a reading surface, not a table panel. A table panel's
     sizes are shares of its height made for 300–400px (the creature at
     52cqh, the back row half that, the commander 0.42), and in a 900px
     overlay they drew 450px creatures over small lands. Here:

     - a header names the seat (.head, its one identity, call 12);
     - every battlefield group is one card size (call 9): what two rows
       of cards get once the header, the strip and the rows' labels,
       padding and gaps (~100px) are placed, never over 240px, times the
       viewer's own card-size setting for every seat (call 16);
     - the command zone is the back row's last group at that size
       (call 10);
     - the piles and the hand sit in a short strip along the bottom
       (call 13), the hand's cards sized so their peek is the strip.

     Every width: the overlay keeps this layout on a phone too (§5).
     The :not(.flipped) is always true here (an expanded panel is never
     flipped); it weighs these rules as much as the seated layout's
     .panel:not(.flipped):not(.docked), which they come after. */
  .panel.expanded:not(.flipped) {
    --x-strip-h: clamp(64px, 10cqh, 96px);
    --card-h: clamp(
      96px,
      calc((100cqh - 52px - var(--x-strip-h) - 100px) / 2 * var(--card-scale, 1)),
      calc(240px * var(--card-scale, 1))
    );
    --card-h-sm: var(--card-h);
    --hand-peek: 0.55;
    --thumb-h: calc(var(--x-strip-h) - 14px);
    --thumb-w: calc(var(--thumb-h) * 5 / 7);
    grid-template-columns: auto minmax(0, 1fr);
    grid-template-rows: auto minmax(var(--creature-min, 0px), auto) 1fr auto;
    grid-template-areas:
      "head      head"
      "creatures creatures"
      "middle    middle"
      "piles     bottom";
  }
  .panel.expanded:not(.flipped) .rail {
    display: contents;
  }
  .panel.expanded:not(.flipped) .rail-gap {
    display: none;
  }
  /* The header: the identity in one line, the turn chip after it, and
     room at the right end for the overlay's pin (Board.svelte), on the
     seat's wash so whose board it is reads before the name. */
  .head {
    grid-area: head;
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
    min-height: 52px;
    box-sizing: border-box;
    padding: 4px 44px 0 6px;
    border-radius: 10px;
    background: color-mix(in srgb, var(--seat-color, var(--fg-dim)) 9%, transparent);
  }
  .turn-chip {
    flex: none;
    padding: 2px 8px;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--seat-color, var(--fg-dim)) 40%, transparent);
    background: color-mix(in srgb, var(--seat-color, var(--fg-dim)) 8%, transparent);
    color: var(--fg-muted);
    font-family: var(--font-mono);
    font-size: 9.5px;
    font-weight: 700;
    letter-spacing: 0.12em;
    line-height: 1.3;
    text-transform: uppercase;
    white-space: nowrap;
  }
  .turn-chip.next {
    color: var(--bg);
    background: var(--seat-color, var(--fg));
    border-color: transparent;
  }
  /* The back row's command group (call 10) and a revealed companion
     draw at the one card size, top-aligned with the zones beside them. */
  .panel.expanded:not(.flipped) .grid-middle :global(.command-strip),
  .panel.expanded:not(.flipped) .grid-middle :global(.companion-slot) {
    --strip-h: var(--card-h);
    align-self: flex-start;
  }
  .panel.expanded:not(.flipped) .grid-middle :global(.companion-slot) {
    padding-top: 18px;
  }
  /* The strip (call 13): the piles at the left and the hand beside
     them, under a hairline, its height fixed. */
  .panel.expanded:not(.flipped) .grid-bottom,
  .panel.expanded:not(.flipped) .rail > :global(.pile-bar) {
    padding-top: 6px;
    border-top: 1px solid var(--border);
  }
  /* Everything in the strip that peeks as the hand does (the hand, your
     castable strip) is sized so its peek is the strip's height. */
  .panel.expanded:not(.flipped) .grid-bottom {
    --card-h: calc(var(--x-strip-h) / var(--hand-peek));
    --card-w: calc(var(--card-h) * 5 / 7);
    align-items: center;
  }
  .panel.expanded:not(.flipped) .hand-zone {
    height: var(--x-strip-h);
  }
  /* The fan's tops are clipped at the strip's hairline, not held over
     the back row: at rest nothing of the hand rises into the board.
     Your own hand still lifts whole on hover. */
  .panel.expanded:not(.flipped) .hand-zone > :global(.hand:not(.opponent):not(.scroll)) {
    --hand-headroom: 0px;
  }
  .strip-hand-count {
    flex: none;
    align-self: center;
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 1px 6px;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: color-mix(in srgb, var(--bg) 82%, transparent);
    color: var(--fg);
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  /* The piles, one row of three, each a thumbnail with its label and
     count beside it, since the strip has the room the seated layout
     does not. */
  .panel.expanded:not(.flipped) .rail > :global(.pile-bar) {
    --pile-col: 1000px;
    grid-area: piles;
    grid-template-columns: repeat(3, auto);
    width: auto;
    align-self: stretch;
    align-content: center;
  }
  .panel.expanded:not(.flipped) .rail :global(.pile) {
    flex-direction: row;
    gap: 6px;
    padding: 3px 8px 3px 3px;
  }
  .panel.expanded:not(.flipped) .rail :global(.pile .label) {
    display: inline-block;
  }
</style>

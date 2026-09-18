<script lang="ts">
  // PlayerPanel is one player's full board, laid out in a CSS Grid
  // matching the S16.5 avatar-centric redesign:
  //
  //   ┌─────────────────────────────────────────────┐
  //   │  creatures (full width)                     │
  //   ├──────────────────────┬──────────────────────┤
  //   │  lands               │  enchant / artifact  │
  //   ├─────────────────────────────────────────────┤
  //   │  ⭕ piles  hand (peek) ……………… phases (self)│
  //   └─────────────────────────────────────────────┘
  //
  // The bottombar is a single flex row: avatar anchors bottom-left,
  // piles immediately beside it, the hand peeks up from the same base-
  // line (top ~55% of each card visible, bottom-clipped at the pile
  // bottom), and PhaseDisplay floats bottom-right on the viewer's own
  // panel. Hovering the self hand lifts the whole fan up over the board
  // to reveal full cards.
  //
  // The same component is reused for self + opponents; Board.svelte
  // wraps opponent instances in a transform container that rotates
  // them 180° across the top so they read "across the table". The
  // panel itself is layout-only and doesn't know whether it's rotated.
  //
  // Click routing for battlefield cards lives here so combat vs.
  // tap-toggle logic stays in one place. The router mirrors the old
  // Pixi wireTapClick: combat select on your own creature, declare-
  // block on an incoming attacker, otherwise tap/untap — except for
  // planeswalkers, whose click opens the card menu (#329). The
  // decision itself is battlefieldClickIntent, in contextMenu.logic,
  // so it is testable without rendering Svelte.

  import type {
    ActionPayload,
    ActionType,
    CardView,
    GameView,
    ManaAbilityView,
    PlayerView,
    ZoneView,
  } from "../../protocol";
  import { bucketForBattlefield, isCreature } from "../../cardTypes";
  import { battlefieldClickIntent } from "../../contextMenu.logic";
  import { canActivateSorcerySpeedAbility } from "../../timing";
  import { counterCostNeedsPrompt } from "../../counterCost";
  import { openCardMenu } from "../../contextMenu";
  import BattlefieldRow from "./BattlefieldRow.svelte";
  import PileBar from "./PileBar.svelte";
  import Hand from "./Hand.svelte";
  import PlayerIdentity from "./PlayerIdentity.svelte";
  import PhaseDisplay from "./PhaseDisplay.svelte";
  import PromisesRow from "./PromisesRow.svelte";

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
    onPlayCard: (card: CardView) => void;
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
    // board-wide modals. Only wired for the viewer's own panel.
    onActivateAbility?: (card: CardView, abilityIndex: number) => void;
    // S21, widened in #789: a mana ability on one of this seat's
    // permanents needs a cost choice before it can be activated — a
    // sacrifice (Ashnod's Altar), or which counters come off and how
    // many (Mage-Ring Network, Iron Spider's cousin on a land). Board
    // owns those modals, so the panel forwards the click instead of
    // sending the action. Only wired for the viewer's own panel.
    onManaAbilityCost?: (card: CardView, ability: ManaAbilityView) => void;
    // Priority controls forwarded to PhaseDisplay — only the
    // self panel mounts the widget, so these only matter when
    // isSelf=true but they're plumbed uniformly for prop typing.
    autopassEnabled?: boolean;
    // #628: the CR 726 loop-breaker banner line, empty when quiet.
    loopNotice?: string;
    onPassPriority?: () => void;
    onToggleAutopass?: () => void;
    // flipped — top-row opponents. The panel keeps its zones in the
    // same grid but reverses the row order (hand at the top edge,
    // creatures toward the table centre) instead of rotating 180°,
    // so text and card art stay upright.
    flipped?: boolean;
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
    autopassEnabled = false,
    loopNotice = "",
    onPassPriority,
    onToggleAutopass,
    onActivateAbility,
    onManaAbilityCost,
    flipped = false,
  }: Props = $props();

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
  // Empty string means "open, no opinion"; opponents' panels are
  // never gated on the VIEWER's window, so they get "" too.
  const sorcerySpeedBlocked = $derived(
    isSelf ? (canActivateSorcerySpeedAbility(view, viewerID).reason ?? "") : "",
  );

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
  // manaAbilityNeedsPrompt is the one predicate; a plain "{T}: Add
  // {G}", and a Vivid land with charge counters on it, go straight to
  // the action as they always did.
  const activateManaAbility = $derived(
    isSelf
      ? (card: CardView, abilityIndex: number) => {
          const ability = (card.mana_abilities ?? []).find((a) => a.index === abilityIndex);
          if (
            ability &&
            onManaAbilityCost &&
            (ability.sacrifice_options || counterCostNeedsPrompt(ability))
          ) {
            onManaAbilityCost(card, ability);
            return;
          }
          sendAction(
            "activate_mana_ability",
            { card_id: card.instance_id, ability_index: abilityIndex },
            seat.id,
          );
        }
      : undefined,
  );

  const attackTargetable = $derived(
    !isSelf && !seat.eliminated && combatMode === "attack" && !!selectedCombatCardID,
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
    // Block-mode click on an incoming attacker commits the block.
    if (combatMode === "block" && card.attacking_target === viewerID) {
      onDeclareBlock(card.instance_id);
      return;
    }
    // #329: a planeswalker's click means "which loyalty ability?",
    // not "tap it". The card menu is the surface that already
    // renders activated abilities (ADR 0028), and it carries the
    // manual loyalty rows for the planeswalkers with no catalog
    // entry — which is still most of them.
    switch (battlefieldClickIntent(card, viewerID, isAdmin)) {
      case "none":
        return;
      case "abilities":
        openCardMenu({ card, x: ev?.clientX ?? 0, y: ev?.clientY ?? 0 });
        return;
      case "tap":
        onTapToggle(card);
    }
  }
</script>

<div
  class="panel"
  class:self={isSelf}
  class:opponent={!isSelf}
  class:flipped
  role="region"
  aria-label={isSelf ? "your board" : `${seat.name} board`}
>
  <div class="grid-creatures">
    <BattlefieldRow
      label="creatures"
      {attachmentsByHost}
      {curseTargets}
      cards={buckets.creature}
      {viewerID}
      {selectedCombatCardID}
      onCardClick={handleCardClick}
      onActivateManaAbility={activateManaAbility}
      {onActivateAbility}
      {sorcerySpeedBlocked}
    />
  </div>
  <!-- The back row: land piles first, then the other permanents,
       both shrink-wrapped and centred as one group so the row grows
       out from the middle like the creature row above it. -->
  <div class="grid-middle">
    <BattlefieldRow
      label="lands"
      {attachmentsByHost}
      {curseTargets}
      cards={buckets.land}
      compact
      strip
      {viewerID}
      {selectedCombatCardID}
      onCardClick={handleCardClick}
      onActivateManaAbility={activateManaAbility}
      {onActivateAbility}
      {sorcerySpeedBlocked}
    />
    <BattlefieldRow
      label="enchant / artifact"
      {attachmentsByHost}
      {curseTargets}
      cards={buckets.right}
      compact
      {viewerID}
      {selectedCombatCardID}
      onCardClick={handleCardClick}
      onActivateManaAbility={activateManaAbility}
      {onActivateAbility}
      {sorcerySpeedBlocked}
    />
  </div>
  <div class="grid-bottom">
    <!-- Hand sits inline with the phase widget; its clipped-bottom
         line coincides with the panel edge. flex: 1 lets it absorb
         the width the rail freed up. -->
    <div class="hand-zone">
      <Hand
        hand={seat.hand}
        {isSelf}
        onPlayCard={isSelf ? onPlayCard : undefined}
        snap={view}
        {viewerID}
      />
    </div>
    {#if !isSelf}
      <PromisesRow {view} {viewerID} opponentID={seat.id} {sendAction} />
    {/if}
    {#if isSelf && view.turn}
      <PhaseDisplay
        turn={view.turn}
        seats={view.seats}
        mulligansOpen={view.mulligans_open === true}
        viewerHasPriority={hasPriority}
        {autopassEnabled}
        {loopNotice}
        onPassPriority={onPassPriority ?? (() => {})}
        onToggleAutopass={onToggleAutopass ?? (() => {})}
      />
    {/if}
  </div>
  <!-- The rail is the player card: identity, floating mana and
       markers on top, the four piles below. It frees the bottom row
       for the hand and the phase widget. -->
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
    />
    <div class="rail-gap"></div>
    <PileBar {seat} {exile} {isSelf} {sendAction} onDrawCard={isSelf ? onDrawCard : undefined} />
  </div>
</div>

<style>
  .panel {
    /* Sept 2026 redesign: zones on the left, a player rail on the
       right. Creatures on top at full size; the middle band is the
       land piles and the other permanents, one size down; the
       bottom row is the hand plus (self only) the phase widget.
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
    --card-h: clamp(168px, calc((42cqh - 30px) * var(--card-scale, 1)), 240px);
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
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    overflow: hidden;
    position: relative;
  }
  /* Self panel carries a gold hairline so the row that's yours reads
     without drawing attention from the cards. */
  .panel.self {
    border-color: rgba(217, 180, 92, 0.28);
    box-shadow: inset 0 0 0 1px rgba(217, 180, 92, 0.08);
  }
  .panel.opponent {
    /* Upright opponent (the "next" seat) gets the medium scale — it
       has the tall row to itself. Scales by --card-scale-opponent
       (settings) with a gentler curve than self. */
    --card-h: clamp(123px, calc((43cqh - 31px) * var(--card-scale-opponent, 1)), 200px);
    --card-h-sm: clamp(90px, calc(var(--card-h) * 0.74), 148px);
    --thumb-w: calc(36px * var(--card-scale-opponent, 1));
    --thumb-h: calc(50px * var(--card-scale-opponent, 1));
    --avatar-size-base: 72px;
    --rail-w: 104px;
  }
  .panel.opponent.flipped {
    /* Across-table seats: one more size down (the top row is the
       short one), rows reversed so the hand hugs the top edge and
       creatures face the centre of the table. */
    --card-h: clamp(90px, calc((43cqh - 31px) * var(--card-scale-opponent, 1)), 168px);
    --card-h-sm: clamp(67px, calc(var(--card-h) * 0.74), 124px);
    --thumb-w: calc(32px * var(--card-scale-opponent, 1));
    --thumb-h: calc(45px * var(--card-scale-opponent, 1));
    --avatar-size-base: 60px;
    grid-template-rows: auto auto minmax(0, 1fr);
    grid-template-areas:
      "bottom    rail"
      "middle    rail"
      "creatures rail";
  }
  .grid-creatures {
    grid-area: creatures;
    min-height: 0;
    min-width: 0;
  }
  /* A flipped panel's creature row is its LAST row, and it must sit
     at the bottom of its area — against the middle of the table,
     facing the viewer's own creatures — not float up under the
     lands. The area keeps the leftover height; the row is pushed to
     its far edge. */
  .flipped .grid-creatures {
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
  }
  .flipped .grid-creatures > :global(.row) {
    flex: 0 1 auto;
  }
  .grid-middle {
    grid-area: middle;
    min-height: 0;
    min-width: 0;
    display: flex;
    justify-content: center;
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
    height: calc(var(--card-h, 168px) * 0.55);
    overflow: visible;
    position: relative;
  }
  /* The viewer's own hand shows 62% of each card at rest (name, cost,
     art and the type line); opponents' face-down fans keep the
     tighter 55%. Hand.svelte's peek and lift use the same share. */
  .panel.self .hand-zone {
    height: calc(var(--card-h, 168px) * 0.62);
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
     aria-label still carry them) and the command-zone hints. */
  .rail {
    overflow-y: auto;
    overflow-x: hidden;
  }
  .flipped .rail :global(.pile .label),
  .flipped .rail :global(.cmd-zone .label),
  .flipped .rail :global(.cmd-zone .hints) {
    display: none;
  }
</style>

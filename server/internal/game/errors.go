// Package game holds the authoritative in-memory game state for a
// single cmd_and_ctrl Commander table: players, zones, cards, and the
// turn/phase/step state machine. It is server-side only; the client
// never imports this package — the client sees a serialised view via
// the v0 protocol (docs/protocol.md). Rules enforcement is deliberately
// absent at S02; mutations are accepted at face value and players
// resolve rules manually until the S13+ B→C rules graft track begins.
package game

import "errors"

var (
	// ErrGameNotInLobby is returned when an operation valid only in the
	// lobby state (like AddPlayer) is attempted on an active or ended game.
	ErrGameNotInLobby = errors.New("game: not in lobby state")

	// ErrGameAlreadyStarted is returned when Start is called on a game
	// that is not in the lobby state.
	ErrGameAlreadyStarted = errors.New("game: already started")

	// ErrGameNotActive is returned when an operation valid only during an
	// active game is attempted in lobby or ended state.
	ErrGameNotActive = errors.New("game: not in active state")

	// ErrGameFull is returned when AddPlayer would exceed MaxPlayers.
	ErrGameFull = errors.New("game: full")

	// ErrNotEnoughPlayers is returned when Start is called with fewer
	// than MinPlayers seats filled.
	ErrNotEnoughPlayers = errors.New("game: not enough players")

	// ErrEmptyName is returned when a player is added without a name.
	ErrEmptyName = errors.New("game: player name cannot be empty")

	// ErrEmptyDeck is returned when a player is added without a deck.
	ErrEmptyDeck = errors.New("game: deck cannot be empty")

	// ErrCardNotFound is returned when a zone operation references a
	// card instance that does not exist in the zone.
	ErrCardNotFound = errors.New("game: card instance not found in zone")

	// ErrZoneEmpty is returned from Top/Bottom/PopTop on an empty zone.
	ErrZoneEmpty = errors.New("game: zone is empty")

	// ErrPlayerNotFound is returned when a mutation references a player
	// ID that is not seated at the table.
	ErrPlayerNotFound = errors.New("game: player not found")

	// ErrZoneNotFound is returned when a ZoneRef cannot be resolved to
	// a concrete Zone on the game (unknown kind or unknown owner).
	ErrZoneNotFound = errors.New("game: zone not found")

	// ErrInvalidParam is returned when an action's parameters are
	// rejected for reasons not covered by a more specific error (e.g.
	// negative hand size on Mulligan).
	ErrInvalidParam = errors.New("game: invalid parameter")

	// ErrIllegalTarget is returned by CastSpell when a target the
	// client sent fails the card's TargetSpec (CR 601.2c) — a black
	// creature for Doom Blade, an opponent's spell for a "target
	// spell you control" clause, a card in the wrong zone. The
	// client's picker only offers legal targets, so hitting this
	// means a stale snapshot or a hand-built payload. Added in S20
	// sub-PR 1.
	ErrIllegalTarget = errors.New("game: illegal target")

	// ErrPlayerEliminated is returned when a mutation targets a player
	// whose Eliminated flag is set, or when an eliminated player tries
	// to concede a second time.
	ErrPlayerEliminated = errors.New("game: player is eliminated")

	// ErrWrongStep is returned when a step-gated mutation is attempted
	// in a step that doesn't permit it (e.g. declare_attacker outside
	// the declare_attackers step). Added in S08.
	ErrWrongStep = errors.New("game: action not legal in current step")

	// ErrNotACreature is returned when a card-targeting mutation
	// requires a creature but the supplied card isn't one (e.g.
	// declaring a land as an attacker). Added in S08.
	ErrNotACreature = errors.New("game: card is not a creature")

	// ErrCardCallerMismatch is returned when a seated player issues a
	// card-instance-scoped action (tap, move_card, add_counter,
	// set_battlefield_position, declare_attacker, declare_blocker) on
	// a card whose Controller is a different seat. Admin / spectator
	// callers (Caller == uuid.Nil) bypass this check. Added in S08.5.
	ErrCardCallerMismatch = errors.New("game: caller does not control this card")

	// ErrNoUndosRemaining is returned by SpendUndo when the named
	// player's per-turn undo budget is already at 0. Refreshed when
	// the cursor enters that player's untap step. Added in S11.
	ErrNoUndosRemaining = errors.New("game: no undos remaining this turn")

	// ErrNoPriority is returned by PassPriority when called during a
	// step that does not grant priority (currently Untap and Cleanup
	// per CR 502.4 / 514.3). Added in S13.
	ErrNoPriority = errors.New("game: no player holds priority this step")

	// ErrSplitSecondActive is returned by cast_spell and
	// activate_ability when an item with SplitSecond is on the stack
	// (CR 702.61). Mana abilities and special actions are still
	// allowed. Added in S13.1.
	ErrSplitSecondActive = errors.New("game: split second is active")

	// ErrLoyaltyAlreadyActivated is returned by activate_loyalty when
	// the planeswalker's loyalty ability has already been activated
	// this turn (CR 606.3). The flag clears when the turn cursor
	// wraps to the next ActiveSeat. Added in S13.1.
	ErrLoyaltyAlreadyActivated = errors.New("game: planeswalker loyalty already activated this turn")

	// ErrInsufficientLoyalty is returned when a loyalty ability's
	// cost would remove more loyalty counters than the planeswalker
	// has (CR 606.6). Paying down to exactly zero is legal — the
	// 704.5i SBA takes it from there — so this fires only on a
	// genuine overpayment. Added in S27 (#329, #334).
	ErrInsufficientLoyalty = errors.New("game: not enough loyalty to pay that cost")

	// ErrNotAPlaneswalker is returned when a loyalty cost is
	// activated on something that isn't a planeswalker (CR 606.2).
	// Guards both the catalog path (a miswritten Spec) and the
	// S13.1 sandbox action, which used to take any battlefield card
	// and hand it loyalty counters. Added in S27 (#329, #334).
	ErrNotAPlaneswalker = errors.New("game: source is not a planeswalker")

	// ErrInvalidStackDestination is returned by counter_spell when the
	// requested destination zone is invalid (battlefield, stack — a
	// counter must move the spell off the stack). Added in S13.1.
	ErrInvalidStackDestination = errors.New("game: invalid stack-counter destination")

	// ErrSorcerySpeedRequired is returned by cast_spell / activate_*
	// when the action is sorcery-speed only (sorceries, loyalty
	// abilities, casts of cards lacking flash) and the sorcery-speed
	// gate is not currently open: stack non-empty, caller is not the
	// active player, or current step isn't a main phase. Added in
	// S13.1.
	ErrSorcerySpeedRequired = errors.New("game: sorcery speed required")

	// ErrLandDropUnavailable is returned by cast_spell when a player
	// plays a land having already used every land play they get this
	// turn (CR 305.2). The allowance is not always one — a controlled
	// Exploration or a one-turn grant raises it — so the message says
	// "no land plays left", not "one land per turn"; the client pairs
	// it with the per-seat count on the wire
	// (PlayerView.LandDropsPerTurn / LandsPlayedThisTurn).
	//
	// Its own sentinel rather than ErrSorcerySpeedRequired or
	// ErrInvalidParam: the player's timing was fine and their payload
	// was fine — they are simply out of land plays, which is a
	// different sentence and a different fix. Added for #500, the
	// first refusal in the engine's move away from the sandbox
	// posture on land drops.
	ErrLandDropUnavailable = errors.New("game: no land plays left this turn")

	// ErrNoPlayPermission is returned when a player tries to play a
	// card from exile without a live impulse-exile grant — the grant
	// belongs to someone else, has expired, was never made, or is
	// cast-only and the card is a land (CR 305.1: playing a land is
	// not casting). Added in S21 sub-PR 6.
	ErrNoPlayPermission = errors.New("game: no permission to play this card from exile")

	// ErrCastZoneNotAllowed is returned by cast_spell when the card
	// does not declare the source zone as one it can be cast from —
	// an ordinary sorcery named as a graveyard cast — or when a
	// zone-bound alternative cost is claimed from the wrong zone
	// (flashback named on a card in hand). Distinct from
	// ErrZoneNotFound, which means the `from_zone` string itself did
	// not name a zone. Added in S29.
	ErrCastZoneNotAllowed = errors.New("game: card cannot be cast from that zone")

	// ErrCastCostRequired is returned by cast_spell when the card is
	// castable from the source zone only by paying a cost bound to
	// that zone, and the cast named none — a graveyard cast of a
	// flashback card that did not claim flashback. Paying the
	// printed cost instead would be strictly better than the card.
	// Added in S29.
	ErrCastCostRequired = errors.New("game: casting from that zone requires its alternative cost")

	// ErrNoManaCost is returned by cast_spell when a non-land card
	// with no mana cost (Ancestral Vision, Living End) is cast by
	// paying that cost. CR 118.6: no mana cost is an unpayable cost,
	// and paying it is illegal, so the cast is refused at announce
	// unless an alternative cost replaces it (CR 118.6a; "without
	// paying its mana cost" counts). Distinct from a {0} cost, which
	// is a real cost of zero. Mode-independent, for the #289 reason:
	// permissive mode's "pay it on paper" has nothing to pay.
	ErrNoManaCost = errors.New("game: a spell with no mana cost can't be cast by paying it")

	// ErrCardNotOnStack is returned by counter_spell / counter_ability
	// when the targeted item is not currently on the stack (already
	// resolved, never cast, or wrong instance ID). Added in S13.1.
	ErrCardNotOnStack = errors.New("game: card is not on the stack")

	// ErrPendingChoiceNotFound is returned by ResolvePendingChoice
	// when the supplied choice ID isn't in the queue (already
	// resolved, wrong ID, or queue was drained by an earlier
	// resolution path). Added in S14 for Thoughtseize-style
	// effects.
	ErrPendingChoiceNotFound = errors.New("game: pending choice not found")

	// ErrNotTheChooser is returned by ResolvePendingChoice when the
	// caller is not the player the choice was addressed to. Added
	// in S14.
	ErrNotTheChooser = errors.New("game: caller is not the chooser of this pending choice")

	// ErrChoiceSetRejected is returned by ResolveChooseCards when the
	// picks are individually fine (right count, all candidates, all
	// still where the prompt found them) but the SET breaks a rule
	// the card prints about them together — ChooseCardsPrompt.
	// Validate refused it. "Discard two cards unless you discard a
	// creature card" answered with one land is the shape.
	//
	// Its own sentinel rather than ErrInvalidParam because the player
	// can fix it by choosing again, and the prompt stays open for
	// exactly that; "invalid parameter" in the prompt's error line
	// would read as a broken client. The player-facing sentence lives
	// in ws.classifyActionError, as it does for ErrUnparseableCost and
	// ErrInvalidFace. The prompt's own Question carries the card's
	// words, so neither needs to repeat them. Added for #624.
	ErrChoiceSetRejected = errors.New("game: chosen cards rejected by the prompt's set rule")

	// ErrAlreadyTapped is returned by ActivateManaAbility when the
	// ability has a tap cost and the permanent is already tapped —
	// the mana-ability rules require the cost to be payable (CR
	// 605.1 + 118.3). Added in S15 sub-PR 2.
	ErrAlreadyTapped = errors.New("game: card is already tapped")

	// ErrInsufficientMana is the sentinel for a cast_spell gated by
	// the S15 strict-mana mode when the caster's pool can't cover
	// the effective cost. CastSpell wraps it in an
	// InsufficientManaError whose Missing slice lists the unpaid
	// symbols; callers that only need to discriminate the error
	// class can still `errors.Is(err, ErrInsufficientMana)`.
	// Added in S15 sub-PR 3.
	ErrInsufficientMana = errors.New("game: insufficient mana")

	// ErrUnparseableCost is returned by CastSpell when the card's
	// printed ManaCost can't be parsed. Split and adventure cards
	// import a joined cost ("{1}{R} // {1}{U}") that ParseCost
	// rightly rejects; before this sentinel existed the cost gate
	// swallowed the parse error and let the spell through for FREE
	// — see #289. Refusing the cast is the honest answer: the
	// engine does not know what the card costs, so neither the
	// strict gate nor permissive paper-tracking can be trusted.
	// Callers wrap it with the offending cost string for the
	// client-facing message.
	ErrUnparseableCost = errors.New("game: unparseable mana cost")

	// ErrCostModifier is returned by CastSpell when a cost modifier
	// (CR 601.2f — "spells cost {1} more / {2} less to cast")
	// produces an amount the engine won't price: a negative
	// increase, a negative reduction, a negative floor.
	//
	// A refusal rather than a clamp, for the #289 reason. The
	// failure mode of clamping is a spell that comes out CHEAPER
	// than printed because a card file's Amount hook had a sign
	// error, and a free spell nobody ordered is the worst thing
	// this path can produce. The cast is rejected, the card stays
	// in hand, and the player sees which permanent misbehaved.
	// Added in S28.
	ErrCostModifier = errors.New("game: invalid cost modifier")

	// ErrInvalidFace is returned by CastSpell when the requested
	// printed face is not one this card offers (ADR 0034): a
	// negative or out-of-range index, or the back face of anything
	// that is not a modal DFC — a transform card's back is reached
	// by transforming the permanent, never by casting it (CR 712.11),
	// and an adventure's second half needs the exile-and-recast
	// permission that is not built yet.
	//
	// Rejecting rather than clamping to the front face is
	// deliberate. A player who meant to play Sea Gate, Reborn as a
	// land and silently got a seven-mana sorcery on the stack has
	// been handed the worst available failure; an error toast is
	// strictly better.
	ErrInvalidFace = errors.New("game: invalid card face")

	// ErrSummoningSick is returned when a creature that entered
	// the battlefield this turn is asked to attack or activate a
	// tap-cost ability without haste (CR 302.6, 702.10). Added in
	// S18 sub-PR 2.
	ErrSummoningSick = errors.New("game: creature has summoning sickness")

	// ErrConditionNotMet is returned when an ability carries an
	// activation restriction (CR 602.5 — "Activate only if you
	// control five or more lands") that the board does not satisfy.
	// Checked before any cost is validated or paid, so the source is
	// untouched. Added in the S32 mana-pipeline pass (#352).
	ErrConditionNotMet = errors.New("game: ability's activation condition is not met")

	// ErrDefender is returned by DeclareAttacker when the creature
	// has the defender keyword (CR 702.3). Added in S18 sub-PR 2.
	ErrDefender = errors.New("game: creature has defender and cannot attack")

	// ErrCantAttack is returned by DeclareAttacker when a continuous
	// effect says the creature can't attack (CR 508.1c) — Pacifism,
	// Arrest, Faith's Fetters. Distinct from ErrDefender because
	// defender is a printed keyword on the creature and this is an
	// effect from somewhere else, and the player who gets the toast
	// needs to know which one they are looking at. Added in S24 with
	// the restriction vocabulary (restrictions.go).
	ErrCantAttack = errors.New("game: an effect prevents this creature from attacking")

	// ErrCantActivate is returned by the activation paths when a
	// continuous effect says this permanent's activated abilities
	// can't be activated (CR 602.5) — Arrest, Faith's Fetters. The
	// mana-ability half is the same error: a player who cannot tap
	// an Arrested Birds of Paradise is being told the same thing.
	// Added in S24.
	ErrCantActivate = errors.New("game: an effect prevents activating this permanent's abilities")

	// ErrIllegalBlock is what DeclareBlocker's refusal wraps when an
	// evasion keyword on the attacker (flying, landwalk) or a CR
	// 509.1b restriction on either card ("~ can't block", "~ can't
	// be blocked") rules out the proposed blocker. The error actually
	// returned is a *BlockRefusedError carrying the reason; test with
	// errors.Is. Game.BlockPairRefusalLocked is the single source of
	// truth (ADR 0045 addendum). Post-S18 fix for the missing gate at
	// the DeclareBlocker call site; restrictions joined it in S24,
	// landwalk and the reason in #705.
	ErrIllegalBlock = errors.New("game: blocker cannot legally block this attacker")

	// ErrNoLegalAttackers is returned by DeclareAttackers when every
	// entry in a bulk declaration was skipped — all tapped, sick,
	// defenders, already declared, or aimed at a seat that can't be
	// attacked. Surfaced rather than swallowed so the room layer
	// leaves the undo stack and the snapshot sequence untouched for
	// what is, in the end, a no-op. Added in S31 for #318.
	ErrNoLegalAttackers = errors.New("game: no creature in the declaration is able to attack")

	// ErrIllegalAttackTarget is returned by DeclareAttacker when the
	// named target is not something this player's creature may attack
	// (CR 506.2, 508.1d): a seat that is not seated or is eliminated,
	// a permanent that is neither a planeswalker nor a battle, the
	// attacker's own controller, a planeswalker they control, or a
	// battle they protect. Distinct from ErrPlayerNotFound, which is
	// what the pre-S27 player-only path returned for all of these and
	// which is now a lie for every permanent target. Added in S27.
	ErrIllegalAttackTarget = errors.New("game: that is not a legal attack target")

	// ErrNotABattle is returned when an operation that only makes
	// sense for a battle — choosing its protector — names a permanent
	// that is not one. Added in S27.
	ErrNotABattle = errors.New("game: card is not a battle")
	// ErrInsufficientCrew is returned when the creatures named to pay
	// a Vehicle's crew cost do not add up to the crew number
	// (CR 702.122a) — including the case where none were named at
	// all. Distinct from ErrInvalidParam so the client can say "tap
	// more power" rather than "bad request". Added in S27.
	ErrInsufficientCrew = errors.New("game: crewing creatures' total power is below the crew number")
	// ErrInsufficientCounters is returned when the permanent named to
	// pay a "remove N counters" cost holds fewer than N counters of
	// the kind — including, for "remove a counter" of any kind, the
	// case where the named kind is not on it at all. Distinct from
	// ErrInvalidParam so the client can say what is missing. Added
	// for #625.
	ErrInsufficientCounters = errors.New("game: not enough counters to pay that cost")
	// ErrCantPayCounterCost is returned when a cost that PUTS a
	// counter on the source cannot be paid — CR 118.3's other
	// direction, where the trouble is not a shortage but a
	// prohibition. Devoted Druid cannot untap itself while something
	// stops it having counters put on it, and the refusal has to
	// happen before anything else is paid. Added for #789.
	ErrCantPayCounterCost = errors.New("game: that permanent can't have those counters put on it")
)

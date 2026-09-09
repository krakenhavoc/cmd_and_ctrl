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
	// (CR 702.79). Mana abilities and special actions are still
	// allowed. Added in S13.1.
	ErrSplitSecondActive = errors.New("game: split second is active")

	// ErrLoyaltyAlreadyActivated is returned by activate_loyalty when
	// the planeswalker's loyalty ability has already been activated
	// this turn (CR 606.5). The flag clears when the turn cursor
	// wraps to the next ActiveSeat. Added in S13.1.
	ErrLoyaltyAlreadyActivated = errors.New("game: planeswalker loyalty already activated this turn")

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

	// ErrNoPlayPermission is returned when a player tries to play a
	// card from exile without a live impulse-exile grant — the grant
	// belongs to someone else, has expired, was never made, or is
	// cast-only and the card is a land (CR 305.1: playing a land is
	// not casting). Added in S21 sub-PR 6.
	ErrNoPlayPermission = errors.New("game: no permission to play this card from exile")

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

	// ErrSummoningSick is returned when a creature that entered
	// the battlefield this turn is asked to attack or activate a
	// tap-cost ability without haste (CR 302.1, 702.10). Added in
	// S18 sub-PR 2.
	ErrSummoningSick = errors.New("game: creature has summoning sickness")

	// ErrDefender is returned by DeclareAttacker when the creature
	// has the defender keyword (CR 702.3). Added in S18 sub-PR 2.
	ErrDefender = errors.New("game: creature has defender and cannot attack")

	// ErrIllegalBlock is returned by DeclareBlocker when evasion
	// keywords on the attacker (flying, menace, fear, shadow, etc.)
	// rule out the proposed blocker. The CanBlock helper is the
	// single source of truth. Post-S18 fix for the missing gate at
	// the DeclareBlocker call site.
	ErrIllegalBlock = errors.New("game: blocker cannot legally block this attacker")
)

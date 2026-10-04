package game

import (
	"strings"

	"github.com/google/uuid"
)

// exile_permanent_cost.go — #1600: exiling permanents you control as
// part of a cost (CR 118.3, CR 602.2b, CR 605.3a, CR 701.13a).
//
//	The Soul Stone      "{6}{B}, {T}, Exile a creature you control:
//	                    Harness The Soul Stone."
//	Altar of Bhaal      "{2}{B}, {T}, Exile a creature you control:
//	                    Return target creature card from your graveyard
//	                    to the battlefield."
//	City of Shadows     "{T}, Exile a creature you control: Put a
//	                    storage counter on this land."
//	Food Chain          "Exile a creature you control: Add X mana of
//	                    any one color, where X is 1 plus the exiled
//	                    creature's mana value." (a MANA ability)
//
// The battlefield sibling of ExileCost (exile_cost.go), which exiles
// CARDS out of the activator's hand or graveyard, and of ExileSelf,
// which exiles the source. This one names OTHER permanents the
// activator controls, so it is a picking problem, and it is the one
// AbilityCost.SacrificeOther and ReturnToHandCost solve: the
// activator's own permanents only, matching the clause, exactly Count
// of them, each once — one candidate walk that the engine, the protocol
// view's picker and the legal-move enumerator all read (#544).
//
// It differs from both, observably:
//
//   - **Not SacrificeOther with a destination.** A sacrifice goes
//     through sacrificePermanentLocked: EventSacrifice, and a dies
//     trigger when the permanent is a creature. Exiling is a plain zone
//     change (CR 701.13a). Nothing dies (exile is not a graveyard), so
//     "whenever a creature dies" never sees it; nothing is sacrificed,
//     so "whenever you sacrifice" never sees it either. The permanent
//     leaves the battlefield through the one exit primitive, so
//     "whenever a permanent leaves the battlefield" always does.
//   - **Not ReturnToHandCost.** The destination is exile, and the
//     permanents it moved are recorded on PaidCost.Exiled, so "the
//     exiled creature" (Food Chain's mana value) can be answered.
//
// THE CLAUSE IS DATA, not a TargetSpec: a card type, read off the
// permanent's current characteristics (Card.HasCardType). Every printed
// clause the component covers is "a <card type> you control", and a
// TargetSpec would have brought its predicate closures into Game's
// reach — ADR 0041's closure ratchet (closure_fields_test.go) admits no
// new func-typed route, and asks that a new component be data. A clause
// a card type cannot say ("another nontoken artifact creature", Curie,
// Emergent Intelligence) grows a field here when its card is built.
//
// And what it shares with every cost component that moves cards: the
// move does not TARGET (CR 601.2h / 602.2b), so hexproof, shroud and
// protection never apply; it goes through routeCardToZoneLocked with
// MustSettleNow, because a cost is one indivisible step; and a
// commander named to it is asked CR 903.9 BEFORE anything is paid
// (#1397, cost_commander_choice.go), the answer riding the move.

// ExilePermanentsCost is the "exile N <permanents> you control"
// component of an ability's cost. Two owners, like ExileCost,
// SacrificeOther and TapOthers: AbilityCost.ExilePermanents (a CR 602
// ability — The Soul Stone) and ManaAbilityShape.ExilePermanents (a
// CR 605 mana ability — Food Chain).
//
// The zero value, and nil, demand nothing — so the activation paths can
// ask without a guard.
type ExilePermanentsCost struct {
	// Count is how many permanents must be exiled. 1 on every printed
	// card the component covers. A VARIABLE count ("one or more other
	// artifacts you control with total mana value X", Fabrication
	// Foundry) is an announcement this component does not model.
	Count int

	// CardType is the card type the permanent must have — "creature"
	// for "Exile a creature you control" — compared case-insensitively
	// against its CURRENT types, so an animated land is a creature and
	// a creature an effect turned into a noncreature artifact is not.
	// Empty is "a permanent you control". effects.Register refuses a
	// value that is not a permanent card type.
	CardType string

	// ExcludeSource is the printed word "another". No catalog card
	// prints it on this clause yet; when it is clear, the source pays
	// if it has the card type.
	ExcludeSource bool

	// Label is the clause as printed, without the verb — "a creature
	// you control" — shown above the client's picker.
	Label string
}

// Empty reports whether the cost demands nothing. Nil-safe.
func (c *ExilePermanentsCost) Empty() bool {
	return c == nil || c.Count < 1
}

// matches reports whether `card` has the clause's card type.
func (c *ExilePermanentsCost) matches(card *Card) bool {
	return c.CardType == "" || card.HasCardType(strings.ToLower(c.CardType))
}

// ExilePermanentsOptionsForEffect is the set of permanents that could
// pay `ec` right now for an ability on `sourceID` activated by
// `playerID`, in battlefield order — the ONE walk the protocol view
// stamps, the legal enumerator pays out of and the validator accepts
// from (#544). A hexproof permanent is on it: paying a cost does not
// target. Whether the options add up to Count is the payment's
// question; this never enumerates subsets.
//
// Caller must hold g.mu (read or write).
func (g *Game) ExilePermanentsOptionsForEffect(playerID, sourceID uuid.UUID, ec *ExilePermanentsCost) []uuid.UUID {
	if ec.Empty() {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != playerID || (ec.ExcludeSource && c.InstanceID == sourceID) || !ec.matches(c) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// ExilePermanentsPayable reports whether `ec` could be paid at all
// right now (CR 118.3).
//
// Caller must hold g.mu (read or write).
func (g *Game) ExilePermanentsPayable(playerID, sourceID uuid.UUID, ec *ExilePermanentsCost) bool {
	if ec.Empty() {
		return true
	}
	return len(g.ExilePermanentsOptionsForEffect(playerID, sourceID, ec)) >= ec.Count
}

// validateExilePermanentsCostLocked checks an ExilePermanents component
// without moving anything — ADR 0020 §3's "validate everything, then
// pay everything", so a refused activation never leaves a board half
// exiled. In the order the errors matter:
//
//   - IDs sent for an ability with no such component are refused
//     rather than ignored, as an unexpected sacrifice_ids or return_ids
//     is (ErrInvalidParam).
//   - Exactly Count permanents, each named once (ErrInvalidParam).
//   - None of them ALSO moved by another component of the same payment
//     (`alsoSpent`: the sacrifices, the returns, the source when the
//     cost exiles or sacrifices it). One permanent pays one component
//     (CR 118.3) (ErrInvalidParam).
//   - Never the source when the clause says "another" (ErrInvalidParam).
//   - On the battlefield (ErrCardNotFound), controlled by the activator
//     (ErrCardCallerMismatch), of the clause's card type
//     (ErrIllegalTarget, the code the sacrifice and return clauses use
//     for a permanent the clause does not admit).
//
// A TAPPED permanent is a legal pick: exiling is not tapping.
//
// Caller must hold g.mu.
func (g *Game) validateExilePermanentsCostLocked(playerID, sourceID uuid.UUID, ec *ExilePermanentsCost, ids, alsoSpent []uuid.UUID) error {
	if ec.Empty() {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if len(ids) != ec.Count {
		return ErrInvalidParam
	}
	spent := make(map[uuid.UUID]bool, len(alsoSpent))
	for _, id := range alsoSpent {
		spent[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] || spent[id] || (ec.ExcludeSource && id == sourceID) {
			return ErrInvalidParam
		}
		seen[id] = true
		c := findBattlefieldCard(g, id)
		if c == nil {
			return ErrCardNotFound
		}
		// "You control" is the cost's own clause (CR 118.3: you can
		// only pay with what is yours to pay with).
		if c.Controller != playerID {
			return ErrCardCallerMismatch
		}
		if !ec.matches(c) {
			return ErrIllegalTarget
		}
	}
	return nil
}

// movedSourceAlso is the `alsoSpent` list validateExilePermanentsCostLocked
// checks a payment against: every permanent another component of the
// same payment moves (`lists` — the sacrifices, which already hold the
// source when the cost sacrifices it, and the returns), plus the source
// when the cost exiles it.
func movedSourceAlso(sourceID uuid.UUID, exileSelf bool, lists ...[]uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	if exileSelf {
		out = append(out, sourceID)
	}
	for _, l := range lists {
		out = append(out, l...)
	}
	return out
}

// payExilePermanentsCostLocked exiles the validated permanents, one
// route each, so a leaves-the-battlefield watcher sees every one of
// them. Through the ONE exit door (routeCardToZoneLocked) with cause
// cost and MustSettleNow: a cost is one indivisible step (CR 601.2h /
// 602.2b, and CR 605.3b for a mana ability), so the move may not stop
// on a prompt. A commander named here was asked CR 903.9 before the
// payment began (askCostCommanderLocked, #1397); `answers` carries what
// its owner said, and a "yes" puts it in the command zone instead —
// the cost is still paid.
//
// Returns the instance IDs it moved, for PaidCost.Exiled. The battlefield
// exit has already written each permanent's last-known information
// (rememberDepartingPermanentLocked), so "the exiled creature's mana
// value" is read as it last existed there (LastKnownPermanentForEffect).
//
// Call only after validateExilePermanentsCostLocked has passed, and
// BEFORE a CR 602 ability's stack item is built, so the leaves-triggers
// this queues are drained by the closing state-check pass and sit ABOVE
// the ability (CR 603.3b). On a mana ability they are drained on the
// way out, with the mana already in the pool.
//
// A permanent that vanished between validation and payment is skipped,
// as payReturnToHandCostLocked skips one.
//
// Caller must hold g.mu.
func (g *Game) payExilePermanentsCostLocked(playerID, sourceID uuid.UUID, ids []uuid.UUID, answers map[uuid.UUID]bool) ([]uuid.UUID, error) {
	var moved []uuid.UUID
	for _, id := range ids {
		if findBattlefieldCard(g, id) == nil {
			continue
		}
		if _, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:        id,
			Dst:           ZoneExile,
			Actor:         playerID,
			Source:        sourceID,
			Cause:         MoveCause{Kind: MoveCauseCost, Controller: playerID},
			MustSettleNow: true,
			// #1397: the owner's CR 903.9 answer, asked before the
			// payment by askCostCommanderLocked.
			commanderAnswer: commanderAnswerFor(answers, id),
		}); err != nil {
			return moved, err
		}
		moved = append(moved, id)
	}
	return moved, nil
}

// PermanentCardTypes are the card types a permanent can have (CR 110.4)
// — the values ExilePermanentsCost.CardType may name, lowercase. Read by
// effects.Register's boot check.
var PermanentCardTypes = []string{"artifact", "battle", "creature", "enchantment", "land", "planeswalker"}

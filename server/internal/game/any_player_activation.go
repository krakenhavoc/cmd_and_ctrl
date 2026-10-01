package game

import "github.com/google/uuid"

// any_player_activation.go — "Any player may activate this ability."
// ADR 0106 §1, #1793.
//
// CR 602.2: "Only an object's controller (or its owner, if it doesn't
// have a controller) can activate its activated ability unless the
// object specifically says otherwise." CR 602.1b: text after the colon
// "may state which players can activate that ability … This text is
// not part of the ability's effect. It functions at all times."
//
// So the permission is part of the ABILITY (ActivatedAbilityShape.
// AnyPlayer), not a record about a player. A layer-6 grant or copy of
// the ability carries it, and an ability-removal effect removes it with
// the ability.
//
// Everything after the "may this player activate" question already
// reads the ACTIVATOR: the board-wide gate, the timing read, the
// Condition, every cost validator and payer, and the stack item's
// Controller (CR 602.1a, CR 113.8, CR 109.5). That is why the rule has
// one function here, mayActivate, and three callers: the activation
// path, internal/legal's enumerator and the view.

// ActivationPurpose is what activating an any-player ability does for
// a player who does NOT control its source — the card-side declaration
// a policy reads to know whether reaching across the table is worth it
// (ADR 0106 §1 decision 8, owner decision 2). ADR 0102's
// ControlPurpose, for an ability instead of a gift.
//
// The engine never reads it. It rides the wire as the row's `purpose`,
// and the heuristic bot activates another player's ability ONLY when
// the row declares one: Flailing Ogre's "{1}: +1/+1" has none, so a bot
// never pumps an opponent's Ogre.
//
// Every field is a printed amount, read from the oracle text by the
// card file, and the zero value is "no declared purpose".
type ActivationPurpose struct {
	// Draws is how many cards the ACTIVATOR draws ("you draw a card",
	// CR 109.5 — "you" is the player who activated it).
	Draws int
	// ControllerLosesLife is how much life the SOURCE'S CONTROLLER
	// loses ("Xantcha's controller loses 2 life").
	ControllerLosesLife int
}

// IsZero reports whether the row declares no purpose.
func (p ActivationPurpose) IsZero() bool {
	return p == ActivationPurpose{}
}

// MayActivate is the CR 602.2 "who may activate this" answer for one
// row of `source`, which is sitting in `zone`. The one predicate the
// activation path, the legal-move enumerator and the view share
// (ADR 0106 §1 decision 2), so a seat is never offered — or shown as
// live — an activation the engine refuses (#544).
//
// On the battlefield: the permanent's controller, or anyone when the
// row says "Any player may activate this ability". Off the battlefield
// the CR 108.4a owner rule is unchanged: no printed any-player ability
// functions from a hidden zone, and Register refuses one that declares
// a zone (checkAnyPlayerAbility in the effects package).
//
// CanActivateAbilities (Arrest) and the board-wide gate are NOT asked
// here. They restrict the object or the ability, not the activator, so
// they apply to every player and are asked by each caller beside this.
func MayActivate(player uuid.UUID, source Card, zone ZoneKind, ab ActivatedAbilityShape) bool {
	if zone == ZoneBattlefield {
		return source.Controller == player || ab.AnyPlayer
	}
	return source.Owner == player
}

// HasAnyPlayerAbility reports whether any activated ability `c` offers
// right now declares AnyPlayer. The enumerator's and the view's fast
// negative for a permanent the seat does not control: almost no
// permanent has one, and the answer is "skip it".
func HasAnyPlayerAbility(c Card) bool {
	for _, ab := range ActivatedAbilitiesForCard(c) {
		if ab.AnyPlayer {
			return true
		}
	}
	return false
}

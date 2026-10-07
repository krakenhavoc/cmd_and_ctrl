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

// What an any-player row buys an activator who does not control it
// is the row's Purpose (purpose.go): the card-side declaration a policy
// reads to know whether reaching across the table is worth it (ADR 0106
// §1 decision 8, owner decision 2). The heuristic bot activates another
// player's ability ONLY when the row declares one, and prices it by its
// Draws and ControllerLosesLife: Flailing Ogre's "{1}: +1/+1" has none,
// so a bot never pumps an opponent's Ogre.

// MayActivate is the CR 602.2 "who may activate this" answer for one
// row of `source`, which is sitting in `zone`. The one predicate the
// activation path, the legal-move enumerator and the view share
// (ADR 0106 §1 decision 2), so a seat is never offered — or shown as
// live — an activation the engine refuses (#544).
//
// On the battlefield: the permanent's controller, or anyone when the
// row says "Any player may activate this ability". Two more named
// activators ride the same predicate (ADR 0106 §1 amendment 2026-10-07,
// #1947): "Only your opponents may activate this ability" (Clergy of
// the Holy Nimbus) is every player BUT the controller, and "Only this
// creature's owner may activate this ability" (Personal Incarnation) is
// the owner alone, so the controller of a stolen one cannot. Off the
// battlefield
// the CR 108.4a owner rule is unchanged: no printed any-player ability
// functions from a hidden zone, and Register refuses one that declares
// a zone (checkAnyPlayerAbility in the effects package).
//
// CanActivateAbilities (Arrest) and the board-wide gate are NOT asked
// here. They restrict the object or the ability, not the activator, so
// they apply to every player and are asked by each caller beside this.
func MayActivate(player uuid.UUID, source Card, zone ZoneKind, ab ActivatedAbilityShape) bool {
	if zone == ZoneBattlefield {
		switch {
		case ab.AnyPlayer:
			return true
		case ab.OpponentsOnly:
			return source.Controller != player
		case ab.OwnerOnly:
			return source.Owner == player
		}
		return source.Controller == player
	}
	return source.Owner == player
}

// ReachesAcross reports whether the row names an activator other than
// the plain controller: any player, only the controller's opponents, or
// only the owner.
func (ab ActivatedAbilityShape) ReachesAcross() bool {
	return ab.AnyPlayer || ab.OpponentsOnly || ab.OwnerOnly
}

// HasAnyPlayerAbility reports whether any activated ability `c` offers
// right now names an activator other than its controller (AnyPlayer,
// OpponentsOnly or OwnerOnly). The enumerator's and the view's fast
// negative for a permanent the seat does not control: almost no
// permanent has one, and the answer is "skip it". The row actually
// named is still held to MayActivate.
func HasAnyPlayerAbility(c Card) bool {
	for _, ab := range ActivatedAbilitiesForCard(c) {
		if ab.ReachesAcross() {
			return true
		}
	}
	return false
}

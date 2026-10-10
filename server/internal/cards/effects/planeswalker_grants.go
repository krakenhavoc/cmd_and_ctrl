package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planeswalker_grants.go — #2797: abilities one permanent gives to a CLASS
// of planeswalkers. "Planeswalkers you control have '[−8]: …'" is printed
// by Kiora of Salt and Sand, Sanctum Lurker, Avatar of Burgeoning Echoes
// and the "Way of the …" cycle.
//
// The grant is an ordinary ADR 0093 layer-6 bundle whose ability row has a
// LoyaltyCost, handed to every planeswalker its controller controls. ADR
// 0109 §2 already made a granted loyalty row the HOST's: the walker pays
// the cost (CR 606.6), and the row shares the walker's one loyalty
// activation a turn with its own printed rows and with every other grant
// (CR 606.3), in either order. The class form adds nothing to that
// except the recipient set, which is read live every layer pass, so a
// planeswalker that enters after the grantor gets the row at once and one
// that leaves takes it with it.
//
// Two grantors (two Ways, a Way and Kiora) give a walker two rows and one
// activation between them, because the count is the permanent's.

// PlaneswalkersYouControl is the recipient predicate of "Planeswalkers you
// control have …": a planeswalker, controlled by the grantor's controller.
// The grantor itself is not one unless it is a planeswalker too. It is a
// StaticAbility.AppliesTo, so it can drive a Spec.Static beside an anthem
// as well as a grant.
func PlaneswalkersYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsPlaneswalker() && target.Controller == source.Controller
}

// GrantAbilitiesToYourPlaneswalkers is GrantAbilities over
// PlaneswalkersYouControl — "Planeswalkers you control have '[−8]: …'".
// Each key names a bundle some card declares in Spec.Grants.
func GrantAbilitiesToYourPlaneswalkers(keys ...string) game.StaticAbility {
	return GrantAbilities(PlaneswalkersYouControl, keys...)
}

// youActivatedALoyaltyAbilityThisTurn is "you've activated a loyalty
// ability this turn": the player has announced an activation whose cost
// had a loyalty symbol (CR 606.2), a printed row or a granted one, since
// the turn began. Read off the turn's event log, so it survives the
// walker leaving. The sandbox's manual loyalty verb (Game.ActivateLoyalty)
// announces no activation event and is not counted.
func youActivatedALoyaltyAbilityThisTurn(g *game.Game, you uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventActivateAbility && ev.Loyalty && ev.Actor == you {
			return true
		}
	}
	return false
}

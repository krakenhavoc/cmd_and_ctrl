package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b8_helpers.go — shared pieces for the eighth Star Trek batch
// (#4973): beam me up, printed and granted, and Marooned's "can't be
// beamed up". The engine half is game/beam_me_up.go.

// BeamMeUp is "Beam me up [cost] (You may cast this card from your
// graveyard for [cost] if you also return a creature you control to its
// owner's hand. Then exile this spell.)": flashback's cast from the
// graveyard and its exile as the spell leaves the stack (Flashback), with
// the return as the second half of the price. The returned creature
// rides alt_cost_ids, matched and never targeted. The card file still
// opens the graveyard:
//
//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
//	AlternativeCosts: []game.AlternativeCost{BeamMeUp("{2}{U}")},
func BeamMeUp(cost string) game.AlternativeCost {
	ac := Flashback(cost)
	ac.Key = game.AltCostKeyBeamMeUp
	ac.Label = "Beam me up " + cost
	ac.ReturnToHand = game.BeamMeUpReturnSpec(uuid.Nil)
	ac.PayLabel = game.BeamMeUpPayLabel
	return ac
}

// CardsInYourGraveyardHaveBeamMeUpDuringYourMainPhase is Scotty, Miracle
// Worker's "During your main phase, each card in your graveyard has beam
// me up. The beam me up cost is equal to that spell's mana cost. Scotty
// can't be beamed up to cast a spell this way." One standing permission
// over its controller's graveyard under game.AltCostKeyGrantedBeamMeUp,
// gated to that player's own main phases. The offer it synthesises
// (game.grantedBeamMeUpOffer) charges the card's mana cost, returns a
// creature other than the granting permanent and exiles the spell, and
// it is made only for an instant or a sorcery, as the reminder text says.
func CardsInYourGraveyardHaveBeamMeUpDuringYourMainPhase(label string) game.CastPermissionGate {
	return game.CastPermissionGate{
		Permission: game.CastPermission{
			Zone: game.ZoneGraveyard,
			// "Each card in your graveyard": the card as it is there (CR
			// 712.8a); only an instant or sorcery spell can be cast this way.
			Filter:            game.PermissionFilter{CardInZone: true, InstantOrSorceryOnly: true},
			AltCostKey:        game.AltCostKeyGrantedBeamMeUp,
			ExileOnResolution: true,
			CastOnly:          true,
			Label:             label,
		},
		Condition: trekB8DuringYourMainPhase,
	}
}

// trekB8DuringYourMainPhase is "during your main phase": either main
// phase of the controller's own turn (CR 505.1).
func trekB8DuringYourMainPhase(g *game.Game, controller, _ uuid.UUID) bool {
	return IsYourTurn(g, controller) &&
		(g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain)
}

// AttachedCantBeBeamedUp is Marooned's "Enchanted permanent ... can't be
// beamed up": a layer-6 static of the Aura that marks its host, so the
// host can't be returned to pay a beam me up cost for as long as the
// Aura is on it.
func AttachedCantBeBeamedUp() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Enchanted permanent can't be beamed up.",
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.CantBeBeamedUp = true
		},
	}
}

// trekB8DifferentlyNamedLands is "the number of differently named lands
// `controller` controls": each land counted once, however many share its
// name. Caller holds g.mu, as every *ForEffect read.
func trekB8DifferentlyNamedLands(g *game.Game, controller uuid.UUID) int {
	names := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsLand() {
			names[c.Name] = true
		}
	}
	return len(names)
}

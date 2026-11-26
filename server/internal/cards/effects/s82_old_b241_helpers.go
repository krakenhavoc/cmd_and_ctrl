package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b241_helpers.go — shared vocabulary of Hour of Devastation batch 241
// (S82 older sets, #5534).

// OldB241ControlsDesertOrHasOneInGraveyard is "you control a Desert or there is
// a Desert card in your graveyard" (Gilded Cerodon), read on effective
// subtypes for the battlefield and on the card's own for the graveyard.
func OldB241ControlsDesertOrHasOneInGraveyard(g *game.Game, you uuid.UUID) bool {
	if ControlsA("Desert")(g, you) {
		return true
	}
	p := g.PlayerByIDForEffect(you)
	if p == nil || p.Graveyard == nil {
		return false
	}
	for _, c := range p.Graveyard.Cards {
		if c.HasSubtype("Desert") {
			return true
		}
	}
	return false
}

// OldB241ControlsAPlaneswalkerOfSubtype is "you control a <subtype>
// planeswalker" (Avid Reclaimer's Nissa), on effective types and subtypes.
func OldB241ControlsAPlaneswalkerOfSubtype(g *game.Game, you uuid.UUID, subtype string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsPlaneswalker() && c.HasSubtype(subtype) {
			return true
		}
	}
	return false
}

// OldB241NoCountersOnPermanentsOfFourTypes is "Counters can't be put on
// artifacts, creatures, enchantments, or lands." (Solemnity): a world static
// that gives every such permanent the ObjectCantHaveCounters prohibition
// (CR 101.2). Planeswalkers and battles are left out, as printed.
func OldB241NoCountersOnPermanentsOfFourTypes() game.StaticAbility {
	return game.StaticAbility{
		Label: "Counters can't be put on artifacts, creatures, enchantments, or lands.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.HasCardType("artifact") || target.HasCardType("creature") ||
				target.HasCardType("enchantment") || target.HasCardType("land")
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.ExtraRestrictions |= game.ObjectCantHaveCounters
		},
	}
}

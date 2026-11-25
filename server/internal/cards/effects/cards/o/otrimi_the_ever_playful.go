package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Otrimi, the Ever-Playful — {3}{B}{G}{U} Legendary Creature — Nightmare Beast
// 6/6 (Commander 2020):
//
//	"Mutate {1}{B}{G}{U}
//	 Trample
//	 Whenever this creature deals combat damage to a player, return target
//	 creature card with mutate from your graveyard to your hand."
//
// Mutate is the shared alternative cost. The trigger is Electryte's "deals
// combat damage to a player"; its target is a creature card in the
// controller's graveyard that has a mutate cost among its catalogued
// alternative costs (the card's own, which is how the engine knows mutate).
// A card the catalog does not know has no mutate here.
func init() {
	t := effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.B36SelfDealtCombatDamageToPlayer(ev, source, g)
	}, "Otrimi — return target creature card with mutate from your graveyard to your hand",
		func(g *game.Game, item *game.StackItem) error {
			return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
		})
	t.Targets = effects.TargetCardInGraveyard("target creature card with mutate from your graveyard", effects.Creature(), effects.YouOwn(), otrimiHasMutate)
	effects.Register(effects.Spec{
		OracleID:         "c5054364-2a19-4b70-869b-94d48b7b12c0",
		Name:             "Otrimi, the Ever-Playful",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{1}{B}{G}{U}")},
		Triggered:        []game.TriggeredAbility{t},
	})
}

// otrimiHasMutate is "a card with mutate".
func otrimiHasMutate(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	for _, ac := range game.AlternativeCostsFor(c.OracleID) {
		if ac.Key == game.MutateKey {
			return true
		}
	}
	return false
}

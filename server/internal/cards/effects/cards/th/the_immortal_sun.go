package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Immortal Sun — Legendary Artifact {6}:
//
//	"Players can't activate planeswalkers' loyalty abilities.
//	 At the beginning of your draw step, draw an additional card.
//	 Spells you cast cost {1} less to cast.
//	 Creatures you control get +1/+1."
//
// Four statics and a trigger: a CR 602.5 activation restriction on
// every player's loyalty abilities, a draw-step trigger (CR 504.2),
// a generic cost reduction on your spells (CR 601.2f) and a layer-7c
// anthem.
//
// No simplification. Star Trek Commander (#4974).
func init() {
	anthem := effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller
	}, 1, 1)
	anthem.Label = "Creatures you control get +1/+1."
	effects.Register(effects.Spec{
		OracleID:     "52232aad-e4ca-4551-8c8d-c69a3b6b49c7",
		Name:         "The Immortal Sun",
		Completeness: effects.CompletenessFull,
		ActivationRestrictions: []game.ActivationRestriction{
			effects.TrekB48PlaneswalkersLoyaltyAbilitiesCantBeActivated(
				"The Immortal Sun — players can't activate planeswalkers' loyalty abilities."),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginDrawStep, effects.ByYou, "The Immortal Sun — draw an additional card", effects.Do(effects.DrawCards{N: 1})),
		},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "The Immortal Sun — spells you cast cost {1} less to cast.", effects.YourSpell()),
		},
		Static: []game.StaticAbility{anthem},
	})
}

package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tempting Witch — Creature — Human Warlock {2}{B}, 1/3 (Throne of Eldraine, #5516):
//
//	"When this creature enters, create a Food token.
//	 {2}, {T}, Sacrifice a Food: Target player loses 3 life."
//
// Sweettooth Witch with a {T} in the cost and 3 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82284fe4-ecb3-4b27-bb03-27290ad15b17",
		Name:         "Tempting Witch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tempting Witch — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice a Food: Target player loses 3 life.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))}),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.Do(effects.TargetPlayerLosesLife{N: 3}),
		}},
	})
}

package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warleader's Call — Enchantment, {1}{R}{W} (EDHREC rank 906):
//
//	"Creatures you control get +1/+1.
//	 Whenever a creature you control enters, this enchantment deals 1
//	 damage to each opponent."
//
// Glorious Anthem stapled to Impact Tremors. The anthem is the
// ordinary layer 7c static; the ping is the Tremors trigger with the
// enchantment as the damage source — a red source, so Torbran adds
// to it and a damage doubler doubles it, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a751c07b-fc21-4854-96e7-f71abf4e86c9",
		Name:         "Warleader's Call",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Creatures you control get +1/+1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.CreatureEnterPaying(effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl, "Warleader's Call — 1 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 1})),
				game.CreatureEnterPayoff{DamageEachOpponent: 1}),
		},
	})
}

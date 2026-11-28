package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Withering Wisps — Enchantment {1}{B}{B} (Ice Age, S82 OLD, #5636):
//
//	"At the beginning of the end step, if no creatures are on the battlefield, sacrifice
//	 this enchantment.
//	 {B}: This enchantment deals 1 damage to each creature and each player. Activate no
//	 more times each turn than the number of snow Swamps you control."
//
// Pestilence's body with a per-turn cap read live: the (n+1)th announcement is refused
// while fewer than n snow Swamps are controlled, and the cap uses the activation tally
// (CR 602.1b), so a count that falls mid-turn stops further activations.
//
// No simplification.
func init() {
	const label = "{B}: This enchantment deals 1 damage to each creature and each player. Activate no more times each turn than the number of snow Swamps you control."
	effects.Register(effects.Spec{
		OracleID:     "ce65c6ed-e5d4-48d8-b569-4e2774b2912c",
		Name:         "Withering Wisps",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
			Cost:    effects.ManaCost("{B}"),
			Condition: func(g *game.Game, controller, source uuid.UUID) bool {
				return g.ActivatedThisTurn(source, label) < effects.S82OldB841SnowSwampsYouControl(g, controller)
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B23DamageEachCreatureAndEachPlayer(effects.NewContext(g, item), 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B23NoCreaturesOnBattlefield(g)
			}, "Withering Wisps — no creatures: sacrifice it", effects.B23SacrificeSelfWhenNoCreatures),
		},
	})
}

package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Black Market Tycoon — Creature — Cat Rogue {R}{G}, 2/2 (Streets of New
// Capenna):
//
//	"At the beginning of your upkeep, this creature deals 2 damage to you
//	 for each Treasure you control.
//	 {T}: Create a Treasure token."
//
// The damage is worked out as the upkeep trigger resolves: two for each
// Treasure its controller has then, and none with no Treasures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "464968a8-865e-4ac8-927e-7c931e3dabb3",
		Name:         "Black Market Tycoon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Black Market Tycoon — deals 2 damage to you for each Treasure you control",
				func(g *game.Game, item *game.StackItem) error {
					n := 2 * effects.S82SncB11Treasures(g, item.Controller)
					return effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: n}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
		}},
	})
}

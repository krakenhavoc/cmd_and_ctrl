package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "464968a8-865e-4ac8-927e-7c931e3dabb3",
		Name:         "Black Market Tycoon",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Black Market Tycoon — deals 2 damage to you for each Treasure you control",
				func(g *game.Game, item *game.StackItem) error {
					n := 2 * s82SncB11Treasures(g, item.Controller)
					return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: n}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  Do(CreateToken{Template: TreasureToken(), N: 1}),
		}},
	})
}

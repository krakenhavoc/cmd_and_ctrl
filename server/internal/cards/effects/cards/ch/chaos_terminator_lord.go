package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chaos Terminator Lord — Creature — Astartes Warrior {3}{R}, 3/3 (Warhammer
// 40,000 Commander, S82 GAP batch 193, #5948):
//
//	"Lord of Chaos — At the beginning of combat on your turn, another target
//	 creature you control gains double strike until end of turn."
//
// The target is chosen as the trigger goes on the stack and re-checked as it
// resolves; "another" excludes the Lord itself. A pick gone by resolution
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Chaos Terminator Lord — another target creature you control gains double strike until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "e9dfab69-7618-4d90-8bb4-cf23581d1bad",
		Name:         "Chaos Terminator Lord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfYourCombat(label, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target := effects.FirstLegalBattlefieldTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					return effects.UntilEndOfTurn(ctx, target, nil, label, game.AddKeywordsMod("double strike"))
				}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}

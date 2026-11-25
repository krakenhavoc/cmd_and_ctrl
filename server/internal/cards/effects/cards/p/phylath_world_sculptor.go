package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phylath, World Sculptor — Legendary Creature — Elemental {4}{R}{G}, 5/5
// (Zendikar Rising, #5508):
//
//	"When Phylath enters, create a 0/1 green Plant creature token for each
//	 basic land you control.
//	 Landfall — Whenever a land you control enters, put four +1/+1 counters
//	 on target Plant you control."
//
// The Plants are counted as the trigger resolves (CR 608.2h), Avenger of
// Zendikar's shape over basic lands. The landfall target is chosen as it
// goes on the stack and re-checked as it resolves (CR 608.2b), and a Plant
// that is not there when it goes on the stack means no trigger at all.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d79cbe2-cf1b-4313-ab47-0118fcdbfc6c",
		Name:         "Phylath, World Sculptor",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Phylath, World Sculptor — create a 0/1 green Plant creature token for each basic land you control",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					n := effects.CountControlled(ctx.Game, ctx.Controller(), effects.IsBasicLand)
					return effects.CreateToken{Template: effects.TokenCard("0/1 green Plant"), N: n}.Apply(ctx)
				}))),
			effects.Targeting(
				effects.Landfall("Phylath, World Sculptor — landfall: put four +1/+1 counters on target Plant you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 4}.Apply(ctx)
					}),
				effects.TargetCreature("target Plant you control", effects.YouControl(), effects.Subtype("Plant"))),
		},
	})
}

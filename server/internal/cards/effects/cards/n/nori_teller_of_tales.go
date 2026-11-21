package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nori, Teller of Tales — Legendary Creature — Dwarf Bard {1}{R/W}, 2/2
// (The Hobbit, HOB):
//
//	"Whenever Nori attacks, target attacking creature gains first strike
//	 until end of turn."
//
// The target is any attacking creature, Nori herself included, chosen as
// the trigger goes on the stack; it is re-checked at resolution (CR 608.2b),
// so one that has left combat gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03253076-9add-4efb-9227-b0a062c86534",
		Name:         "Nori, Teller of Tales",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Nori, Teller of Tales — target attacking creature gains first strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"},
						Label: "Nori, Teller of Tales — first strike"}.Apply(ctx)
				}),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature())),
		},
	})
}

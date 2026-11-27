package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skittering Horror — Creature — Phyrexian Horror {2}{B}, 4/3 (Urza's Destiny,
// S82 #5621):
//
//	"When you cast a creature spell, sacrifice this creature."
//
// Fires on any creature spell its controller casts, whatever its source. It
// is a mandatory trigger: the creature spell being cast stays on the stack and
// resolves as normal; the Horror is sacrificed first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4344abeb-47ab-4638-b50f-dd502f14c1af",
		Name:         "Skittering Horror",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(effects.Creature()), "Skittering Horror — sacrifice it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

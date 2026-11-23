package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Festering Goblin — Creature — Zombie Goblin {B}, 1/1 (Dominaria
// Remastered):
//
//	"When this creature dies, target creature gets -1/-1 until end of turn."
//
// Festering Newt's dies trigger over any creature, not only an
// opponent's. The target is chosen as the trigger goes on the stack (it
// may be any creature, including one the Goblin's controller controls); a
// creature that is gone or no longer a legal target on resolution is left
// alone (CR 608.2b).
//
// No simplification.
func init() {
	t := effects.WhenThisDies("Festering Goblin — target creature gets -1/-1 until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -1,
				Toughness: -1,
				Label:     "Festering Goblin — -1/-1 until end of turn",
			}.Apply(effects.NewContext(g, item))
		})
	t.Targets = effects.TargetCreature("target creature")
	effects.Register(effects.Spec{
		OracleID:     "66fb4764-d309-4c30-a2a4-474f9030dc87",
		Name:         "Festering Goblin",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

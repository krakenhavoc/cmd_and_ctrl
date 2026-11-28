package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kjeldoran Dead — Creature — Skeleton {B}, 3/1 (Ice Age, S82 OLD batch 846, #5636):
//
//	"When this creature enters, sacrifice a creature.
//	 {B}: Regenerate this creature."
//
// Vindictive Mob's sacrifice: its controller chooses a creature as the trigger
// resolves, and the Dead may be the one (CR 701.21a); with none left they
// sacrifice nothing. The regeneration is the shared row, a shield on this
// permanent (CR 701.19a).
//
// No simplification.
func init() {
	const label = "Kjeldoran Dead — sacrifice a creature"
	effects.Register(effects.Spec{
		OracleID:     "61e00946-7df7-4298-91b2-c656169e6601",
		Name:         "Kjeldoran Dead",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller, effects.SacrificeSpec("a creature", effects.Creature()), label)
				return nil
			}),
		},
		Activated: []effects.ActivatedAbility{effects.RegenerateThisRow("{B}")},
	})
}

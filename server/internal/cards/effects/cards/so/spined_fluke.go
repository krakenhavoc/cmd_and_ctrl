package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spined Fluke — Creature — Worm Horror {2}{B}, 5/1 (Urza's Saga, S82 OLD
// batch 730, #5624):
//
//	"When this creature enters, sacrifice a creature.
//	 {B}: Regenerate this creature."
//
// The trigger is mandatory and untargeted: its controller sacrifices a creature
// of their choice as it resolves, and the Fluke itself is a legal choice (the
// only one when it enters alone). The second ability is the stock regeneration
// shield (CR 701.19a).
//
// No simplification.
func init() {
	const label = "Spined Fluke — sacrifice a creature"
	effects.Register(effects.Spec{
		OracleID:     "3c880ac2-0b22-4f7e-a55a-381de235e971",
		Name:         "Spined Fluke",
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

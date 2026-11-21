package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demonic Taskmaster — Creature — Demon {2}{B}, 4/3 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice a creature other than
//	 this creature."
//
// An edict on yourself over every creature but the Taskmaster: the
// choice is made as the trigger resolves and the Taskmaster itself is
// never offered. With nothing else to sacrifice, nothing happens
// (CR 701.21a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "162e45f5-edb8-4a13-acdc-8980ce5ba8ac",
		Name:            "Demonic Taskmaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Demonic Taskmaster — sacrifice a creature other than this creature",
				func(g *game.Game, item *game.StackItem) error {
					source := item.SourceCardID
					g.PlayerSacrificesForEffect(source, item.Controller,
						effects.SacrificeSpec("a creature other than this creature", effects.Creature(), effects.NotSelf(source)),
						"Demonic Taskmaster — sacrifice a creature other than this creature")
					return nil
				}),
		},
	})
}

package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gathering of Darkness — Sorcery {3}{B} (The Hobbit, HOB):
//
//	"Return up to one target creature card from your graveyard to your
//	 hand.
//	 Amass Goblins 3."
//
// With no target chosen the spell still resolves and amasses; with a target
// that has left the graveyard it does not resolve (CR 608.2b), as Fates'
// Reversal. The amass waits for any prompt the return asks (CR 608.2c, #4155).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85e1556d-c4da-4979-a5db-4ce3842c7797",
		Name:         "Gathering of Darkness",
		Purpose:      game.Purpose{Amass: 3},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("up to one target creature card from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 1),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand),
			effects.Amass{Subtype: "Goblin", N: 3},
		),
	})
}

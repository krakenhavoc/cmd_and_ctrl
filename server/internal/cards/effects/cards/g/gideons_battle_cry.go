package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gideon's Battle Cry — Sorcery {2}{W}{W} (War of the Spark):
//
//	"Put a +1/+1 counter on each creature you control. You may search your
//	 library and/or graveyard for a card named Gideon, the Oathsworn, reveal it,
//	 and put it into your hand. If you search your library this way, shuffle."
//
// The counters, then the optional search (Basri's Aegis' step): a library and/or
// graveyard search for the named card, shuffling only if the library was
// searched, and finding nothing is allowed (CR 701.19b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7528f47d-c758-42bc-bdfe-5aca805c4126",
		Name:         "Gideon's Battle Cry",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.B11PutCountersOnEachCreatureYouControl(ctx.Game, ctx.Item, 1)
			}),
			effects.Instruction(effects.S82OldB105MayFetchNamed("Gideon's Battle Cry", "Gideon, the Oathsworn")),
		),
	})
}

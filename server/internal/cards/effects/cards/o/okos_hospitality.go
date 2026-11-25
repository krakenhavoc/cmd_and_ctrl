package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oko's Hospitality — Instant {3}{G}{U} (Throne of Eldraine, #5516):
//
//	"Creatures you control have base power and toughness 3/3 until end of
//	 turn. You may search your library and/or graveyard for a card named
//	 Oko, the Trickster, reveal it, and put it into your hand. If you search
//	 your library this way, shuffle."
//
// The creatures you control as it resolves are the ones changed (CR
// 611.2c), a layer 7b set that ends at cleanup, so counters and pumps still
// apply on top. The search is the keyed answer Liliana's Scorn uses.
//
// No simplification.
func init() {
	const label = "Oko's Hospitality — base power and toughness 3/3 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "d804b0e5-785a-4e6e-9fe7-38b8add826fa",
		Name:         "Oko's Hospitality",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.Creature(), effects.YouControl()), label, game.SetBasePTMods(3, 3)...)
			}),
			effects.Instruction(effects.S82OldB105MayFetchNamed("Oko's Hospitality", "Oko, the Trickster")),
		),
	})
}

package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liberating Combustion — Sorcery {4}{R} (Kaladesh, #5539):
//
//	"Liberating Combustion deals 6 damage to target creature. You may
//	 search your library and/or graveyard for a card named Chandra,
//	 Pyrogenius, reveal it, and put it into your hand. If you search your
//	 library this way, shuffle."
//
// Basri's Aegis' shape: the damage goes to the target if it is still legal
// (CR 608.2b), and the search is its own sentence, the planeswalker-fetch
// cycle's keyed "you may" (zones named first, shuffle only after a library
// search).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3b40e34-edad-4838-a4d5-c167037d9ff7",
		Name:         "Liberating Combustion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SourceDealsDamageToEachLegalTarget(6)(ctx.Game, ctx.Item)
			}),
			effects.Instruction(effects.S82OldB105MayFetchNamed("Liberating Combustion", "Chandra, Pyrogenius")),
		),
	})
}

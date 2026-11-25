package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grasping Current — Sorcery {4}{U} (Ixalan):
//
//	"Return up to two target creatures to their owners' hands.
//	 Search your library and/or graveyard for a card named Jace, Ingenious
//	 Mind-Mage, reveal it, and put it into your hand. If you search your
//	 library this way, shuffle."
//
// Waterwhirl's simultaneous bounce, then Dovin's Dismissal's named search. Unlike
// Basri's Aegis the search is not a "you may": the zones are asked (CR 701.23b
// lets the searcher find nothing) and the shuffle follows only a library search.
// The search is made even with no target chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a10a4bb2-c6fa-482b-9c8b-e64cbb8424ba",
		Name:         "Grasping Current",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				ctx.Game.BounceCardsToHandForEffect(effects.LegalTargetCardIDs(ctx))
				return nil
			}),
			effects.Instruction(effects.S82OldB105MayFetchNamed("Grasping Current", "Jace, Ingenious Mind-Mage")),
		),
	})
}

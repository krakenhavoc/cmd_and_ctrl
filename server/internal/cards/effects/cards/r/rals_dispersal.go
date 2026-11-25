package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ral's Dispersal — Instant {3}{U}{U} (Guilds of Ravnica):
//
//	"Return target creature to its owner's hand. You may search your library
//	 and/or graveyard for a card named Ral, Caller of Storms, reveal it, and
//	 put it into your hand. If you search your library this way, shuffle."
//
// The bounce goes to the creature's owner if the target is still legal. The
// search is its own sentence, Basri's Aegis' keyed "you may" answer, so it
// runs even when the target was lost (CR 608.2b only voids a spell whose
// every target is illegal, in which case the whole spell does nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a81b7629-9db9-4351-8d0e-c45a7a6167d4",
		Name:         "Ral's Dispersal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, id := range effects.LegalTargetCardIDs(ctx) {
					if err := (effects.BounceToHand{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.Instruction(effects.S82OldB105MayFetchNamed("Ral's Dispersal", "Ral, Caller of Storms")),
		),
	})
}

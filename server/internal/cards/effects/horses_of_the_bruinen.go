package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horses of the Bruinen — Sorcery {3}{U}{U}:
//
//	"Return up to two target creatures to their owners' hands. Scry 1.
//	 The Ring tempts you."
//
// The creatures are returned together, then the scry, then the tempt,
// each in the one before's continuation: a returned commander may stop
// for its owner's command-zone answer (CR 903.9), and the scry stops
// for its own choice. With no target chosen the spell still scries and
// tempts; with every chosen target illegal it does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "100d1c9e-15aa-4bab-9fb3-76c2cd6abc2f",
		Name:         "Horses of the Bruinen",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ids := legalTargetIDs(ctx)
			tempt := ringTemptsYouNext(item)
			return ctx.Game.BounceCardsToHandThenForEffect(ids, func(g *game.Game, _ []uuid.UUID) error {
				return Scry{N: 1, Then: tempt}.Apply(NewContext(g, item))
			})
		},
	})
}

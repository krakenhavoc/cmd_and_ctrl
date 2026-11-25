package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Madcap Experiment — Sorcery {3}{R} (Kaladesh, #5539):
//
//	"Reveal cards from the top of your library until you reveal an artifact
//	 card. Put that card onto the battlefield and the rest on the bottom of
//	 your library in a random order. Madcap Experiment deals damage to you
//	 equal to the number of cards revealed this way."
//
// The reveal-until family's sentence (RevealUntilThenPutOntoBattlefield's
// body, with the revealed run kept for the count): a library with no
// artifact reveals itself whole, bottoms it all and deals that many
// damage. A token is revealed but is never the hit (CR 108.2). The damage
// is the whole run, the artifact included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "684e2237-b860-45bf-a053-a02cc7ed1527",
		Name:         "Madcap Experiment",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				player := ctx.Controller()
				run, hit := effects.RevealUntil(ctx, player, func(c game.Card) bool { return c.IsArtifact() },
					"Madcap Experiment — reveal until an artifact card")
				err := effects.PutFromLibraryOntoBattlefield{
					Player:   player,
					Cards:    run,
					Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
					All:      true,
					ThenTail: effects.PutRestOnBottomInRandomOrderTail,
				}.Apply(ctx)
				if err != nil {
					return err
				}
				return effects.DealDamage{Source: ctx.Source(), Target: player, Amount: len(run)}.Apply(ctx)
			}),
		),
	})
}

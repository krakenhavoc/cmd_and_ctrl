package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Telemin Performance — Sorcery {3}{U}{U} (Conflux):
//
//	"Target opponent reveals cards from the top of their library until they
//	 reveal a creature card. That player puts all noncreature cards revealed
//	 this way into their graveyard, then you put the creature card onto the
//	 battlefield under your control."
//
// Oath of Druids' reveal-until, run on the target opponent's library: the
// run is a real reveal, a token never stops it (CR 108.2), the noncreature
// cards go to their owner's graveyard first and the creature card then enters
// under the caster's control. A library with no creature card is revealed
// entirely and all of it goes to the graveyard. A target opponent who is no
// longer legal makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4014f665-08a6-4ae0-b4d7-5a69d043af6a",
		Name:         "Telemin Performance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim := effects.FirstLegalTargetPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			run, hit := effects.RevealUntil(ctx, victim, func(c game.Card) bool { return c.IsCreature() }, "Telemin Performance")
			var rest []uuid.UUID
			for _, id := range run {
				if id != hit {
					rest = append(rest, id)
				}
			}
			if err := effects.RestIntoGraveyard(ctx.Game, rest); err != nil {
				return err
			}
			if hit == uuid.Nil {
				return nil
			}
			_, err := ctx.Game.PutFromLibraryOntoBattlefieldForEffect(hit, game.LibraryEntryOptions{Controller: ctx.Controller()})
			return err
		},
	})
}

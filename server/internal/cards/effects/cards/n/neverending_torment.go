package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Neverending Torment — Sorcery {4}{B}{B} (Saviors of Kamigawa, #5601):
//
//	"Search target player's library for X cards, where X is the number of
//	 cards in your hand, and exile them. Then that player shuffles.
//	 Epic (For the rest of the game, you can't cast spells. At the
//	 beginning of each of your upkeeps, copy this spell except for its epic
//	 ability. You may choose a new target for the copy.)"
//
// Jester's Cap's search over the target's library, the caster choosing the
// cards (no quality, so exactly X, or the whole library if it is smaller,
// CR 701.23d); X is the caster's hand as the spell resolves (CR 608.2h), and
// the target's library is shuffled either way, even for X of zero. Epic is
// effects.Epic (CR 702.50a); each copy counts the hand again and may be
// aimed at a new player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f07d3d0-1f79-4d50-ab3e-f9dc5ef59482",
		Name:         "Neverending Torment",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: effects.Epic(func(_ *game.StackItem, ctx *effects.Context) error {
			victim, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			x := effects.SokB1CardsInYourHand(ctx)
			if x <= 0 {
				return effects.ShuffleLibrary{Player: victim}.Apply(ctx)
			}
			return effects.SearchLibrary{
				Player:       ctx.Controller(),
				LibraryOwner: victim,
				Dest:         game.ZoneExile,
				Limit:        x,
				Shuffle:      true,
				Reason:       "Neverending Torment — search target player's library for X cards and exile them",
				Source:       ctx.Source(),
			}.Apply(ctx)
		}),
	})
}

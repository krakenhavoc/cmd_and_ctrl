package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jester's Cap — Artifact {4} (Dominaria Remastered):
//
//	"{2}, {T}, Sacrifice this artifact: Search target player's library for
//	 three cards and exile them. Then that player shuffles."
//
// Bribery's search by the ability's controller of another player's
// library (SearchLibrary.LibraryOwner), here over any card, up to three,
// into exile, and the library's owner shuffles at the end. Any player may
// be the target, the controller included. "Three cards" is a ceiling when
// the library holds fewer (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8783a94a-d0b2-4291-a550-dbcf8e16bf4b",
		Name:         "Jester's Cap",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Search target player's library for three cards and exile them. Then that player shuffles.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.SearchLibrary{
					Player:       ctx.Controller(),
					LibraryOwner: item.Targets[0].ID,
					Dest:         game.ZoneExile,
					Limit:        3,
					Shuffle:      true,
					Reason:       "Jester's Cap — search target player's library for three cards and exile them",
				}.Apply(ctx)
			},
		}},
	})
}

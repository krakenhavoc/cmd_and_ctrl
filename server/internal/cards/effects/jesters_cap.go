package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8783a94a-d0b2-4291-a550-dbcf8e16bf4b",
		Name:         "Jester's Cap",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Search target player's library for three cards and exile them. Then that player shuffles.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				ctx := NewContext(g, item)
				return SearchLibrary{
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

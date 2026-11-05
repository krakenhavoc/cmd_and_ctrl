package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beastrider Vanguard — Creature — Human Knight {1}{G}, 2/2 (Aetherdrift,
// #4464):
//
//	"{4}{G}: Look at the top three cards of your library. You may reveal
//	 a permanent card from among them and put it into your hand. Put the
//	 rest on the bottom of your library in any order."
//
// Duskwatch Recruiter's dig with "permanent card" for "creature card":
// a look, so only the controller sees the three; the card taken is
// revealed; the rest go to the bottom in an order the controller picks
// (takeRevealRestBottomAny).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19a0d191-64c0-4e8a-8cb0-bf680cc681bb",
		Name:         "Beastrider Vanguard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}: Look at the top three cards of your library. You may reveal a permanent card from among them and put it into your hand. Put the rest on the bottom of your library in any order.",
			Cost:    ManaCost("{4}{G}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				player := ctx.Controller()
				return TakeFromLibraryToHand{
					Player:   player,
					Cards:    g.LookAtTopOfLibraryForEffect(player, 3),
					Match:    Permanent(),
					Max:      1,
					Optional: true,
					Label:    "Beastrider Vanguard — you may reveal a permanent card from among them and put it into your hand",
					Key:      takeRevealRestBottomAny,
				}.Apply(ctx)
			},
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spartan Veteran — Creature — Human Soldier {R}, 1/1 (Assassin's
// Creed, #5466):
//
//	"During your turn, this creature has first strike.
//	 {2}: This creature gets +1/+0 until end of turn."
//
// Inspiring Paladin's turn-gated first strike on itself, and Fire Nation
// Cadets' firebreathing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b255d1ec-49ac-49a2-9af8-21915bbcd224",
		Name:         "Spartan Veteran",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			}, "first strike"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Spartan Veteran — +1/+0 until end of turn"),
		}},
	})
}

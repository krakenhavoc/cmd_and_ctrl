package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shipwreck Sentry — Creature — Human Pirate {1}{U}, 3/3 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Defender
//	 As long as an artifact entered the battlefield under your control
//	 this turn, this creature can attack as though it didn't have
//	 defender."
//
// Platypus-Bear's waiver (avB18CanAttackAsThoughNoDefenderWhile), on only
// while the controller's entry tally shows an artifact this turn (Master's
// Guide-Mural's read, EnteredWithCardTypeThisTurn). The keyword stays, and
// the waiver is asked only as attackers are declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f05575f8-4961-4780-83fc-3584d9146a14",
		Name:            "Shipwreck Sentry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{
			avB18CanAttackAsThoughNoDefenderWhile(
				"As long as an artifact entered the battlefield under your control this turn, this creature can attack as though it didn't have defender.",
				func(g *game.Game, source *game.Card) bool {
					return g.EnteredWithCardTypeThisTurn(source.Controller, "artifact") > 0
				}),
		},
	})
}

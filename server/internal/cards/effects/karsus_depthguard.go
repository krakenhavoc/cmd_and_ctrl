package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karsus Depthguard — Creature — Lizard Warrior {2}{R}, 4/3 (March of
// the Machine):
//
//	"Defender
//	 As long as this creature's power is 5 or greater, it can attack as
//	 though it didn't have defender."
//
// Bristlepack Sentry's defender waiver, conditioned on this creature's
// own power read through the layered view.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dc8c6803-dd87-4f00-8a92-9613dbe90549",
		Name:            "Karsus Depthguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{avB18CanAttackAsThoughNoDefenderWhile(
			"As long as this creature's power is 5 or greater, it can attack as though it didn't have defender.",
			func(_ *game.Game, source *game.Card) bool {
				return source.CurrentPower() >= 5
			})},
	})
}

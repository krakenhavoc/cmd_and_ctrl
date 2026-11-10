package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bristlepack Sentry — Creature — Plant Wolf {1}{G}, 3/3 (Outlaws of Thunder Junction):
//
//	"Defender
//	 As long as you control a creature with power 4 or greater, this
//	 creature can attack as though it didn't have defender."
//
// The conditional defender waiver (CR 702.3b): a self-only layer-6 record
// that leaves the keyword on the creature, read live at declaration
// (avB18CanAttackAsThoughNoDefenderWhile, Platypus-Bear's shape). Current
// power counts, so a pump or a counter turns it on and off.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "05852d99-8f30-4318-913c-d0b1eaf8c222",
		Name:            "Bristlepack Sentry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{avB18CanAttackAsThoughNoDefenderWhile(
			"As long as you control a creature with power 4 or greater, this creature can attack as though it didn't have defender.",
			func(g *game.Game, source *game.Card) bool {
				return b24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
			})},
	})
}

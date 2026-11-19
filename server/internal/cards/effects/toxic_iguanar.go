package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Toxic Iguanar — Creature — Lizard {R}, 1/1 (Conflux, #5587):
//
//	"This creature has deathtouch as long as you control a green
//	 permanent. (Any amount of damage a creature with deathtouch deals to
//	 a creature is enough to destroy it.)"
//
// A layer-6 grant (CR 613.1f) read continuously through
// conB1WhileYouControlA: the colour is a word of the Iguanar's text (CR
// 612.2).
//
// No simplification.
func init() {
	deathtouch := KeywordGrant(conB1WhileYouControlA("G"), "deathtouch")
	deathtouch.Label = "This creature has deathtouch as long as you control a green permanent."
	Register(Spec{
		OracleID:     "87e148a0-4299-4771-b68a-6652a4c1c535",
		Name:         "Toxic Iguanar",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{deathtouch},
	})
}

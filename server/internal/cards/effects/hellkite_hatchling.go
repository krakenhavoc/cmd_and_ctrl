package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hellkite Hatchling — Creature — Dragon {2}{R}{G}, 2/2 (Conflux,
// #5587):
//
//	"Devour 1 (As this creature enters, you may sacrifice any number of
//	 creatures. It enters with that many +1/+1 counters on it.)
//	 This creature has flying and trample if it devoured a creature."
//
// Devour is the shared replacement (devour.go, CR 702.82a), which
// records the number devoured on the permanent (Card.Devoured). The two
// layer-6 grants (CR 613.1f) read that record, so a Hatchling that
// devoured nothing, or one that left and came back as a new object (CR
// 400.7), has neither keyword.
//
// No simplification.
func init() {
	devoured := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfCreature(target, g, source) && source.Devoured > 0
	}
	flying := KeywordGrant(devoured, "flying")
	flying.Label = "This creature has flying and trample if it devoured a creature."
	trample := KeywordGrant(devoured, "trample")
	trample.Label = "This creature has flying and trample if it devoured a creature."
	Register(Spec{
		OracleID:     "17b1094f-be67-465c-a91e-4010113bb352",
		Name:         "Hellkite Hatchling",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Devour("Hellkite Hatchling", 1)},
		Static:       []game.StaticAbility{flying, trample},
	})
}

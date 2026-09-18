package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squirrel Sovereign — Creature — Squirrel Noble {1}{G}, 2/2 (EDHREC
// rank 4297):
//
//	"Other Squirrels you control get +1/+1."
//
// The Squirrel lord, and a textbook one: "other", "you control", a
// flat +1/+1, nothing else. It is in the batch because the Squirrel
// deck is a real Commander archetype whose whole plan is a wide board
// of 1/1s, and an anthem that drops the "other" or the "you control"
// is the difference between a lord and a symmetrical enchantment.
//
// TribeFilter{Others, YoursOnly} is exactly those two words, so the
// Sovereign does not pump itself and does not pump the Squirrels an
// opponent's Chatterfang made. It reads EFFECTIVE subtypes, so a
// changeling you control is a Squirrel and gets the bonus.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f43d1ea5-8127-4b64-a87b-ee5cc9b9e9fa",
		Name:         "Squirrel Sovereign",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Squirrel"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}

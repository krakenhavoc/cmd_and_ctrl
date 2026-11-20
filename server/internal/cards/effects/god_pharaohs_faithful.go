package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// God-Pharaoh's Faithful — Creature — Human Wizard {W}, 0/4 (Hour of
// Devastation, #5534):
//
//	"Whenever you cast a blue, black, or red spell, you gain 1 life."
//
// A cast trigger over the spell's colours (Balefire Liege's shape): one
// spell of two of those colours is one trigger, not two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f86e6c6-fca5-43e9-88cc-4546cdd5f0ec",
		Name:         "God-Pharaoh's Faithful",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(OfColor("U"), OfColor("B"), OfColor("R")), "God-Pharaoh's Faithful — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}

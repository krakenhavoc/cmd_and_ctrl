package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Walking Skyscraper — Artifact Creature — Construct {8}, 8/8 (Kamigawa:
// Neon Dynasty):
//
//	"This spell costs {1} less to cast for each modified creature you
//	 control. (Equipment, Auras you control, and counters are
//	 modifications.)
//	 Trample
//	 This creature has hexproof as long as it's untapped."
//
// The reduction is a self cost modifier counting live; the hexproof is a
// layer 6 grant on itself while untapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "90c763b9-6173-46ff-b31c-7dcafc774170",
		Name:            "Walking Skyscraper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(neoB10ModifiedCount,
				"This spell costs {1} less to cast for each modified creature you control."),
		},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && !target.Tapped
			}, "hexproof"),
		},
	})
}

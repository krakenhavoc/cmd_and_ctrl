package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Behind the Scenes — Enchantment {2}{B} (Tarkir: Dragonstorm Commander):
//
//	"Creatures you control have skulk. (They can't be blocked by
//	 creatures with greater power.)
//	 {4}{W}: Creatures you control get +1/+1 until end of turn."
//
// The skulk is a layer-6 grant to every creature its controller has
// (the tribal grant with an empty tribe list); the block rule itself is
// the engine's (CR 702.118). The pump is Dawnwing Marshal's: it reaches
// the creatures you control as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e84837bd-4f94-43e1-876e-d82cae6ee247",
		Name:         "Behind the Scenes",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{TribalKeywordGrant(TribeFilter{YoursOnly: true}, "skulk")},
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}: Creatures you control get +1/+1 until end of turn.",
			Cost:    ManaCost("{4}{W}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Effect: Do(BoostUntilEOT{
				Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
				Label: "Behind the Scenes — creatures you control get +1/+1",
			}),
		}},
	})
}

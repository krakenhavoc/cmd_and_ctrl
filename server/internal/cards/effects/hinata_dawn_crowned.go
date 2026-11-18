package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hinata, Dawn-Crowned — {1}{U}{R}{W} Legendary Creature — Kirin Spirit 4/4
// (Kamigawa: Neon Dynasty, S82 NEO batch 20):
//
//	"Flying, trample
//	 Spells you cast cost {1} less to cast for each target.
//	 Spells your opponents cast cost {1} more to cast for each target."
//
// Two cost modifiers (CR 601.2f) that read the targets announced for the
// spell (CR 601.2c comes before cost determination), so each declares
// ReadsTargets and the engine, the auto-tap preview and the legal-move
// enumerator all price a cast per target set. Only spells are priced (an
// activated ability is not a spell). A spell with no target costs the
// same; the reduction takes generic mana only (CR 601.2f), and a target
// counts as many times as it is chosen as a target (CR 115.3 lets one
// object be chosen for different instances of "target").
//
// No simplification.
func init() {
	targetCount := func(q game.CostQuery) int { return game.RealTargetCount(q.Targets) }
	less := CostsLessEach(targetCount, "Spells you cast cost {1} less to cast for each target.", YourSpell())
	less.ReadsTargets = true
	more := CostsMoreEach(targetCount, "Spells your opponents cast cost {1} more to cast for each target.", OpponentsSpell())
	more.ReadsTargets = true
	Register(Spec{
		OracleID:        "aee473c7-c659-421c-96b5-05f57ca2ed97",
		Name:            "Hinata, Dawn-Crowned",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		CostModifiers:   []game.CostModifier{less, more},
	})
}

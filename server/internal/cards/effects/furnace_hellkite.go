package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Furnace Hellkite — Artifact Creature — Dragon {5}{R}{R}, 5/5 (Modern
// Horizons 3, #5467):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Flying
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Valkyrie Aerial Unit's affinity (CR 702.41a) on Shivan Dragon's body:
// the reduction counts artifacts you control as the spell is priced, and
// the firebreathing pumps the ability's own source. A Hellkite that left
// and came back before the ability resolved is a new object and gets
// nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "647980b2-088c-43a4-bdc4-5865f03a4fc4",
		Name:            "Furnace Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Furnace Hellkite — +1/+0"}.Apply(NewContext(g, item))
			},
		}},
	})
}

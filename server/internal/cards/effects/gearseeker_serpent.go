package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gearseeker Serpent — Creature — Serpent {5}{U}{U}, 5/6 (Aetherdrift,
// #4464):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 {5}{U}: This creature can't be blocked this turn."
//
// Memory Guardian's affinity (CR 702.41a; it takes {1} off for each
// artifact, so {U}{U} is never reduced) and Giant Koi's self-evasion for
// mana. The Serpent can't
// be blocked only while it is still the permanent that activated: one
// that left and came back in response is a new object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fdbd8a95-1dc8-4df2-bab0-a93d1941a405",
		Name:         "Gearseeker Serpent",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{5}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Gearseeker Serpent — can't be blocked"}.Apply(NewContext(g, item))
			},
		}},
	})
}

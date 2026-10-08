package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zahid, Djinn of the Lamp — Legendary Creature — Djinn {4}{U}{U}, 5/6:
//
//	"You may pay {3}{U} and tap an untapped artifact you control rather
//	 than pay this spell's mana cost.
//	 Flying"
//
// ADR 0135 §1 (#2030): the tap alternative cost with a mana half (CR
// 118.9). The artifact is named in alt_cost_ids, so the auto-tapper won't
// also tap it for the {3}{U} (CR 118.3); an artifact that is a mana source
// can be tapped for mana or for the cost, not both. Commander tax is added
// to the {3}{U} (CR 118.9d).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "89b07037-64df-4e66-acd8-87d97df61e3a",
		Name:            "Zahid, Djinn of the Lamp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{
			TapInsteadPaying("{3}{U}", 1, "an untapped artifact you control", Artifact()),
		},
	})
}

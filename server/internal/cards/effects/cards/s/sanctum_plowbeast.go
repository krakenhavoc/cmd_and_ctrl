package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctum Plowbeast — Artifact Creature — Beast {4}{W}{U}, 3/6 (Alara
// Reborn, #5586):
//
//	"Defender
//	 Plainscycling {2}, islandcycling {2} ({2}, Discard this card: Search
//	 your library for a Plains or Island card, reveal it, put it into your
//	 hand, then shuffle.)"
//
// Defender is the printed keyword. The two landcycling keywords are two
// cycling abilities (CR 702.29e), each the shared Typecycling row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ff3fae6e-d8a1-46b3-a6cc-fb4e8800ee68",
		Name:            "Sanctum Plowbeast",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Plainscycling", "{2}", "a Plains card", effects.IsLandWithSubtype("Plains")),
			effects.Typecycling("Islandcycling", "{2}", "an Island card", effects.IsLandWithSubtype("Island")),
		},
	})
}

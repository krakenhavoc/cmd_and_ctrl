package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pale Recluse — Creature — Spider {4}{G}{W}, 4/5 (Alara Reborn, #5586):
//
//	"Reach (This creature can block creatures with flying.)
//	 Forestcycling {2}, plainscycling {2} ({2}, Discard this card: Search
//	 your library for a Forest or Plains card, reveal it, put it into your
//	 hand, then shuffle.)"
//
// Reach is the printed keyword. The two landcycling keywords are two
// cycling abilities (CR 702.29e), each the shared Typecycling row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "488ca652-653b-4bb3-891c-874efdd418ff",
		Name:            "Pale Recluse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Forestcycling", "{2}", "a Forest card", effects.IsLandWithSubtype("Forest")),
			effects.Typecycling("Plainscycling", "{2}", "a Plains card", effects.IsLandWithSubtype("Plains")),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chorus of the Tides — Creature — Siren {3}{U}, 3/2 (Born of the Gods,
// #5559):
//
//	"Flying
//	 Heroic — Whenever you cast a spell that targets this creature, scry 1.
//	 (To scry 1, look at the top card of your library, then you may put
//	 that card on the bottom.)"
//
// The heroic seam (heroic) with a scry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "48925016-8a55-4efb-84c7-dc5001baa58e",
		Name:            "Chorus of the Tides",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			heroic("Chorus of the Tides — scry 1 (heroic)", Do(Scry{N: 1})),
		},
	})
}

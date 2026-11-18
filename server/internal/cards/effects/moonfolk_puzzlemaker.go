package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moonfolk Puzzlemaker — Artifact Creature — Moonfolk Wizard {2}{U}, 1/4
// (Neon Dynasty):
//
//	"Flying
//	 Whenever this creature becomes tapped, scry 1."
//
// Attentive Sunscribe's trigger: "becomes tapped" is the tap event and the
// attack that taps it, so a vigilance attack is not one (CR 701.26a). The
// scry is the footprinted primitive (CR 701.22a), its rider declared on
// the card's own Purpose.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8ee2371c-e2c3-412e-b9e9-1ec511c9d537",
		Name:            "Moonfolk Puzzlemaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Moonfolk Puzzlemaker — scry 1",
				Do(Scry{N: 1})),
		},
	})
}

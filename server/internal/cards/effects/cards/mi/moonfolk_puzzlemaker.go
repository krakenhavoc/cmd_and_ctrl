package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "8ee2371c-e2c3-412e-b9e9-1ec511c9d537",
		Name:            "Moonfolk Puzzlemaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Moonfolk Puzzlemaker — scry 1",
				effects.Do(effects.Scry{N: 1})),
		},
	})
}

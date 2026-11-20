package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flockchaser Phantom — Creature — Spirit {4}{W}{U}, 5/5 (March of the
// Machine Commander, #5647):
//
//	"Convoke
//	 Flying, vigilance
//	 Whenever this creature attacks, the next spell you cast this turn
//	 has convoke."
//
// Its own convoke is Spec.TapCost. The attack trigger is Wand of the
// Worldsoul's promise (NextSpellHasConvoke): the next spell its
// controller casts this turn may be paid for by tapping creatures, the
// Phantom among them, since vigilance left it untapped. A spell that
// already has convoke gains nothing more (CR 702.51d) and still spends
// the promise.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6471335e-e054-4e68-b02e-052d65a2fdfc",
		Name:            "Flockchaser Phantom",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		TapCost:         effects.Convoke(),
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Flockchaser Phantom — the next spell you cast this turn has convoke",
				effects.Do(effects.NextSpellHasConvoke("Flockchaser Phantom"))),
		},
	})
}

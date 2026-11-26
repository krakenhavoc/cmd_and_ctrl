package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ire Shaman — Creature — Orc Shaman {1}{R}, 2/1 (Dragons of Tarkir, S82 old
// batch 331):
//
//	"Menace
//	 Megamorph {R} (You may cast this card face down as a 2/2 creature for {3}.
//	 Turn it face up any time for its megamorph cost and put a +1/+1 counter on
//	 it.)
//	 When this creature is turned face up, exile the top card of your library.
//	 Until end of turn, you may play that card."
//
// Megamorph is the shared constructor. The face-up trigger is Abbot of Keral
// Keep's impulse draw (OutpostSiegeImpulse): the top card is exiled and its
// controller may play it until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "7f691e0d-dac5-4d7e-8727-dd2cab9a7863",
		Name:             "Ire Shaman",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"menace"},
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{R}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp("Ire Shaman — exile the top card of your library; until end of turn, you may play it",
				effects.OutpostSiegeImpulse),
		},
	})
}

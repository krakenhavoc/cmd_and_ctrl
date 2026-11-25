package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rowdy Crew — Creature — Human Pirate {2}{R}{R}, 3/3 (Ixalan, #5532):
//
//	"Trample
//	 When this creature enters, draw three cards, then discard two cards at
//	 random. If two cards that share a card type are discarded this way, put
//	 two +1/+1 counters on this creature."
//
// One trigger: the draw, the random discard of two, then the check on those
// two cards' types (a land and a land, a creature and a creature, an
// artifact creature and a plain creature, and so on). The check reads the
// cards where the discard left them; the counters go on the Crew only if it
// is still the same permanent (see s82_old_b227_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3d1d30e1-8934-4451-9b32-c6db2167fc2e",
		Name:            "Rowdy Crew",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rowdy Crew — draw three cards, then discard two cards at random; if they share a card type, put two +1/+1 counters on this creature", effects.OldB227RowdyCrewETB),
		},
	})
}

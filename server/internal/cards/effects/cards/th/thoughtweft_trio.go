package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thoughtweft Trio — Creature — Kithkin Soldier {2}{W}{W}, 5/5 (Lorwyn, S82 OLD,
// #5593):
//
//	"First strike, vigilance
//	 Champion a Kithkin (When this enters, sacrifice it unless you exile another
//	 Kithkin you control. When this leaves the battlefield, that card returns to
//	 the battlefield.)
//	 This creature can block any number of creatures."
//
// The shared champion helper (CR 702.72) over Kithkin and Palace Guard's block
// capacity on itself (CR 509.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "15fca8d6-8a86-4c24-9178-f0b6496518ab",
		Name:            "Thoughtweft Trio",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance"},
		Triggered:       []game.TriggeredAbility{thoughtweftTrioChampion.Trigger("Thoughtweft Trio")},
		Static:          []game.StaticAbility{effects.CanBlockAnyNumber(effects.SelfOnly)},
	})
}

// thoughtweftTrioChampion is the champion pick, keyed. An on-disk key: never
// renamed, never reused.
var thoughtweftTrioChampion = effects.NewChampion("Thoughtweft Trio", "Kithkin", "choose-cards/s82-old-b556-thoughtweft-trio-champion", game.TailRef{})

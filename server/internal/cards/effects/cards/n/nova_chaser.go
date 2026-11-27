package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nova Chaser — Creature — Elemental Warrior {3}{R}, 10/2 (Lorwyn, S82 OLD,
// #5593):
//
//	"Trample
//	 Champion an Elemental (When this enters, sacrifice it unless you exile
//	 another Elemental you control. When this leaves the battlefield, that card
//	 returns to the battlefield.)"
//
// Trample is printed; the shared champion helper (CR 702.72) over Elementals.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "997f5847-04c2-498a-a294-d3355fbb7424",
		Name:            "Nova Chaser",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{novaChaserChampion.Trigger("Nova Chaser")},
	})
}

// novaChaserChampion is the champion pick, keyed. An on-disk key: never renamed,
// never reused.
var novaChaserChampion = effects.NewChampion("Nova Chaser", "Elemental", "choose-cards/s82-old-b556-nova-chaser-champion", game.TailRef{})

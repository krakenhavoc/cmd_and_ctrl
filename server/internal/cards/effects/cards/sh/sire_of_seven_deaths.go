package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sire of Seven Deaths — Creature — Eldrazi {7}, 7/7 (Foundations,
// #4465):
//
//	"Reach, first strike
//	 Vigilance, menace
//	 Trample, lifelink
//	 Ward—Pay 7 life."
//
// Six printed keywords and Sedgemoor Witch's life ward at 7: a spell or
// ability an opponent controls that targets the Sire is countered
// unless its controller pays 7 life, and one who can't pay isn't asked
// (CR 119.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a15547d0-bbcb-41ee-a9ae-8325a9e7fe1f",
		Name:            "Sire of Seven Deaths",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "first strike", "vigilance", "menace", "trample", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardLife(7), "Sire of Seven Deaths — ward, pay 7 life"),
		},
	})
}

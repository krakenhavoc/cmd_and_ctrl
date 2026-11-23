package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Master Piandao — Legendary Creature — Human Warrior Ally {4}{W}, 4/4:
//
//	"First strike
//	 Whenever Master Piandao attacks, look at the top four cards of
//	 your library. You may reveal an Ally, Equipment, or Lesson card
//	 from among them and put it into your hand. Put the rest on the
//	 bottom of your library in a random order."
//
// The subtypes are read through the card's effective type line, so a
// changeling counts as an Ally. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e294d9d2-df91-4410-85d3-8ffb13fb2285",
		Name:            "Master Piandao",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Master Piandao — look at the top four cards; you may reveal an Ally, Equipment, or Lesson card and put it into your hand",
				effects.LookAtTopThenMayTakeToHand(4, effects.AnySubtype("Ally", "Equipment", "Lesson"), 1,
					"Master Piandao — reveal an Ally, Equipment, or Lesson card and put it into your hand")),
		},
	})
}

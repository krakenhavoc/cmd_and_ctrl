package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancestor's Chosen — Creature — Human Cleric {5}{W}{W}, 4/4 (Judgment,
// #5611):
//
//	"First strike
//	 When this creature enters, you gain 1 life for each card in your
//	 graveyard."
//
// The count is taken as the trigger resolves, so it sees what the graveyard
// holds then (the creature itself is on the battlefield, not in it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fc2ccab7-cab1-4463-b73d-898070136d74",
		Name:            "Ancestor's Chosen",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ancestor's Chosen — you gain 1 life for each card in your graveyard", effects.OldB623GainLifePerCardInYourGraveyard),
		},
	})
}

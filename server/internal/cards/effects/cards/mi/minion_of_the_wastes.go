package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minion of the Wastes — {3}{B}{B}{B} Creature — Minion, */* (Tempest, S82 OLD
// batch 769, #5628):
//
//	"Trample
//	 As this creature enters, pay any amount of life.
//	 Minion of the Wastes's power and toughness are each equal to the life paid
//	 as it entered."
//
// The payment is Phyrexian Processor's (PayAnyAmountOfLifeAsEnters): any amount
// from 0 to its controller's life total (CR 119.4), none when their life total
// can't change (CR 119.8), stored on the permanent (Card.ChosenNumber) for as
// long as this object lasts (CR 400.7). The power and toughness are a
// characteristic-defining ability (CR 604.3, layer 7a) reading that number
// live, so a Minion with nothing paid is 0/0 and dies to state-based actions
// (CR 704.5f). A bot pays half its life, keeping at least 10.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2593944c-1bde-4307-b3ca-f7ecbeb027c1",
		Name:            "Minion of the Wastes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		AsEnters:        effects.PayAnyAmountOfLifeAsEnters("Minion of the Wastes", game.PayAmountPower, effects.S82OldB769PayGoal),
		Static: []game.StaticAbility{
			effects.S82OldB769LifePaidPT("Minion of the Wastes's power and toughness are each equal to the life paid as it entered."),
		},
	})
}

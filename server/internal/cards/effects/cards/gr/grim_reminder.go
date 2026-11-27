package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Reminder — Instant {2}{B} (Mirrodin, #5606):
//
//	"Search your library for a nonland card and reveal it. Each opponent who
//	 cast a spell this turn with the same name as that card loses 6 life.
//	 Then shuffle.
//	 {B}{B}: Return this card from your graveyard to your hand. Activate only
//	 during your upkeep."
//
// The search only reveals: the card stays in the library (HoldFound), and the
// tail reads its name, makes each opponent who cast a spell this turn with that
// name lose 6 life and shuffles. The graveyard ability is Necrosavant's window
// (an activation from the graveyard, CR 113.6, in your own upkeep) with Unlawful
// Entry's return to hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12684e95-2652-4c3c-b2c7-021b62344d77",
		Name:         "Grim Reminder",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OldB605GrimReminderSearch,
		Activated: []effects.ActivatedAbility{{
			Label:     "{B}{B}: Return this card from your graveyard to your hand. Activate only during your upkeep.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{B}{B}"),
			Zones:     []game.ZoneKind{game.ZoneGraveyard},
			Condition: effects.DuringYourUpkeep(),
			Effect:    effects.ReturnThisFromYourGraveyardToHand,
		}},
	})
}

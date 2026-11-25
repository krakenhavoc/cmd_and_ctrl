package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra's Regulator — Legendary Artifact {1}{R} (Core Set 2020, #5518):
//
//	"Whenever you activate a loyalty ability of a Chandra planeswalker, you
//	 may pay {1}. If you do, copy that ability. You may choose new targets for
//	 the copy.
//	 {1}, {T}, Discard a Mountain card or a red card: Draw a card."
//
// The trigger is Battlemage's Bracers's: only a CR 602 activation emits the
// event, the {1} is a resolution-time optional payment, and on payment the
// activation still on the stack beneath the trigger is copied with the
// new-targets offer. A copy is not an activation, so it pays no loyalty cost
// and does not trigger this again. The second ability's cost discards a card
// in hand that is red or a Mountain card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8c95a86-2f05-44f9-8ed7-b1798eaed42f",
		Name:         "Chandra's Regulator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventActivateAbility, effects.YouActivateAChandraLoyaltyAbility,
				"Chandra's Regulator — you may pay {1} to copy that ability", effects.OldB154RegulatorOffer),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Discard a Mountain card or a red card: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost: effects.Plus(effects.ManaCost("{1}"), effects.TapCost(),
				effects.DiscardCardsMatching(1, "a Mountain card or a red card", effects.OldB154RedOrMountainCard)),
			Effect: effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}

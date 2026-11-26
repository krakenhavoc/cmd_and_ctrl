package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elite Arcanist — Creature — Human Wizard {3}{U}, 1/1 (Magic 2014, S82 old
// sets):
//
//	"When this creature enters, you may exile an instant card from your hand.
//	 {X}, {T}: Copy the exiled card. You may cast the copy without paying its
//	 mana cost. X is the mana value of the exiled card."
//
// Isochron Scepter's imprint, with no mana-value cap on the instant: the
// exile is LINKED to this Arcanist object (CR 607.2a), so a second Arcanist's
// card, or a card this one imprinted before it left and came back, is never the
// one copied. The activation's {X} is defined by the text (CR 107.3c, #6240):
// the engine prices it as generic mana equal to the mana value of the card
// exiled with this Arcanist, nothing asks for an X, and with nothing imprinted
// it costs only {T}. The copy is cast free with additional costs still owed;
// not casting it is the "may".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bc3c4e2-087d-4255-a576-0b17ed1e8c3f",
		Name:         "Elite Arcanist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Elite Arcanist — you may exile an instant card from your hand", effects.S82OldB400ArcanistImprint),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{X}, {T}: Copy the exiled card. You may cast the copy without paying its mana cost. X is the mana value of the exiled card.",
			Cost:  effects.Plus(effects.XIsTheManaValueOfTheExiledCard(), effects.TapCost()),
			// What it answers is whatever the imprinted instant does, so the
			// row declares the union of every answer an instant can give
			// (ADR 0142's ruling for a conditional row), as Isochron Scepter.
			Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerSurvive | game.AnswerPump |
				game.AnswerPrevent | game.AnswerRemove | game.AnswerSacOutlet | game.AnswerRestrict |
				game.AnswerGraveyard | game.AnswerCombatGrant | game.AnswerAnimate | game.AnswerMakesBlocker},
			Effect: effects.CopyExiledWithAndCast,
		}},
	})
}

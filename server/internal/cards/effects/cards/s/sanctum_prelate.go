package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sanctum Prelate — Creature — Human Cleric {1}{W}{W}, 2/2:
//
//	"As this creature enters, choose a number.
//	 Noncreature spells with mana value equal to the chosen number can't be
//	 cast."
//
// The number is Talion's as-enters prompt, but any number from 0 up with no
// ceiling, stored on the permanent (game.Card.ChosenNumber, a per-instance
// value that is not copiable). The ban is Gaddock Teeg's symmetrical "can't be
// cast": it binds every player, the Prelate's controller too (CR 101.2), reads
// the spell's mana value with the X the caster announced (CR 202.3e) and never
// refuses a creature spell or an ability. Zero is a number: it bars the
// noncreature spells with mana value 0.
//
// The as-enters family's declared window applies (game/choose_number.go): the
// Prelate is on the battlefield with a stored 0 while the prompt is open, and
// nothing can be cast while a prompt blocks the table.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e0cdba8-45e0-4fe5-b440-2d8fcf4d6b2b",
		Name:         "Sanctum Prelate",
		Completeness: effects.CompletenessFull,
		AsEnters:     sanctumPrelateChooseANumber,
		CastRestrictions: []game.CastRestriction{{
			Label: "Noncreature spells with mana value equal to the chosen number can't be cast.",
			Forbids: func(q game.CastQuery) bool {
				if q.Card.IsCreature() {
					return false
				}
				mv, ok := q.SpellManaValue()
				return ok && mv == q.Source.ChosenNumber
			},
		}},
	})
}

// sanctumPrelateChooseANumber is the as-enters prompt: any number, stored on
// the Prelate.
func sanctumPrelateChooseANumber(card *game.Card, ctx *effects.Context) error {
	return ctx.Game.QueueChooseNumberForEffect(game.ChooseNumber{
		Chooser:  card.Controller,
		Source:   card.InstanceID,
		Resource: game.PayResourceNone,
		Min:      0,
		NoMax:    true,
		Goal:     3,
		Question: "Sanctum Prelate — choose a number",
		Branch:   game.NumberBranch{Then: effects.StoreChosenNumber, Carry: []uuid.UUID{card.Controller}},
	})
}

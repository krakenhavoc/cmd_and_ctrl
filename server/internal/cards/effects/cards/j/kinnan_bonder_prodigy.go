package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kinnan, Bonder Prodigy — Legendary Creature — Human Druid {G}{U}, 2/2:
//
//	"Whenever you tap a nonland permanent for mana, add one mana of any
//	 type that permanent produced.
//	 {5}{G}{U}: Look at the top five cards of your library. You may put a
//	 non-Human creature card from among them onto the battlefield. Put the
//	 rest on the bottom of your library in a random order."
//
// The first ability is a triggered mana ability (CR 605.1b): it fires off
// a {T} mana ability of a nonland permanent you control, adds one mana of a
// type that permanent produced (one produced type is no choice, two is a
// pick), and never re-triggers. The additional mana is Kinnan's, so it
// carries none of the tapped permanent's restrictions (the printed ruling).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d11aa49-d4cd-48b1-aa0f-8548fa733416",
		Name:         "Kinnan, Bonder Prodigy",
		Completeness: effects.CompletenessFull,
		ManaTriggers: []game.ManaTrigger{
			effects.WheneverYouTapANonlandPermanentForMana("Kinnan, Bonder Prodigy — add one mana of any type that permanent produced",
				effects.AddsOneManaOfAnyTypeProduced()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}{U}: Look at the top five cards of your library. You may put a non-Human creature card from among them onto the battlefield. Put the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{5}{G}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.PutFromLibraryOntoBattlefield{
					Player:   item.Controller,
					Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 5),
					Match:    effects.And(effects.Creature(), effects.Not(effects.HasSubtype(ctx.TextSubtype("Human")))),
					Max:      1,
					Optional: true,
					Label:    "Kinnan, Bonder Prodigy — you may put a non-Human creature card onto the battlefield",
					ThenTail: effects.PutRestOnBottomInRandomOrderTail,
				}.Apply(ctx)
			},
		}},
	})
}

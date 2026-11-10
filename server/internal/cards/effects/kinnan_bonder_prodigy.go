package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8d11aa49-d4cd-48b1-aa0f-8548fa733416",
		Name:         "Kinnan, Bonder Prodigy",
		Completeness: CompletenessFull,
		ManaTriggers: []game.ManaTrigger{
			WheneverYouTapANonlandPermanentForMana("Kinnan, Bonder Prodigy — add one mana of any type that permanent produced",
				AddsOneManaOfAnyTypeProduced()),
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}{G}{U}: Look at the top five cards of your library. You may put a non-Human creature card from among them onto the battlefield. Put the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{5}{G}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return PutFromLibraryOntoBattlefield{
					Player:   item.Controller,
					Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 5),
					Match:    And(Creature(), Not(HasSubtype(ctx.TextSubtype("Human")))),
					Max:      1,
					Optional: true,
					Label:    "Kinnan, Bonder Prodigy — you may put a non-Human creature card onto the battlefield",
					ThenTail: PutRestOnBottomInRandomOrderTail,
				}.Apply(ctx)
			},
		}},
	})
}

package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stenn, Paranoid Partisan — Legendary Creature — Human Wizard {W}{U},
// 2/2 (Dominaria United, #5487):
//
//	"As Stenn enters, choose a card type other than creature or land.
//	 Spells you cast of the chosen type cost {1} less to cast.
//	 {1}{W}{U}: Exile Stenn. Return it to the battlefield under its
//	 owner's control at the beginning of the next end step."
//
// The as-enters choice is Cloud Key's ChooseOptionAsEnters over the card
// types a spell can have besides creature and land, and the discount is
// "Spells you cast of the chosen type cost {1} less" (Cloud Key's discount, read
// through Arachne's chosen-type predicate). The
// activated ability exiles only the Stenn it came from and schedules the
// return at the next end step (Djinn of the Fountain's helper); Stenn
// comes back as a new object and chooses again (CR 400.7).
//
// No simplification.
func init() {
	const name = "Stenn, Paranoid Partisan"
	effects.Register(effects.Spec{
		OracleID:     "03d2d7dd-54b6-469e-8b1c-c48db8cc0925",
		Name:         name,
		Completeness: effects.CompletenessFull,
		AsEnters: effects.ChooseOptionAsEnters(name,
			"artifact", "battle", "enchantment", "instant", "kindred", "planeswalker", "sorcery"),
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Spells you cast of the chosen type cost {1} less to cast.",
				effects.YourSpell(), effects.ArachneChosenTypeSpell()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}{U}: Exile Stenn. Return it to the battlefield under its owner's control at the beginning of the next end step.",
			Cost:    effects.ManaCost("{1}{W}{U}"),
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileThisThenScheduleReturn(effects.NewContext(g, item), item, name+" — return Stenn")
			},
		}},
	})
}

package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Springjack Pasture — Land (Commander 2013, S82 OLD b387):
//
//	"{T}: Add {C}.
//	 {4}, {T}: Create a 0/1 white Goat creature token.
//	 {T}, Sacrifice X Goats: Add X mana of any one color. You gain X life."
//
// The first and third abilities are mana abilities (CR 605.1a). The third reads
// the X the activator announces with it (#6125, CR 107.3a, 602.2b via 605.3a):
// exactly X Goats are sacrificed as the cost, X mana of the one colour the
// activator names are added, and X life is gained as it resolves. X = 0
// sacrifices nothing and adds nothing. The second is an ordinary activated
// ability (Woe Strider's Goat).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9eaadbbc-818b-4c21-9d4b-1bba48504d38",
		Name:         "Springjack Pasture",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:            effects.ManaAbilityCost{Tap: true, SacrificeOther: effects.SacrificeX("X Goats", effects.HasSubtype("Goat")).SacrificeOther},
				ProducedForPaid: effects.XManaOfAnyOneColor,
				GainLifeX:       true,
				Answers:         game.AnswerSacOutlet,
				Label:           "{T}, Sacrifice X Goats: Add X mana of any one color. You gain X life.",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Create a 0/1 white Goat creature token.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 white Goat"), N: 1}),
		}},
	})
}

package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Umori, the Collector — {2}{B/G}{B/G} Legendary Creature — Ooze, 4/5:
//
//	"Companion — Each nonland card in your starting deck shares a card type.
//	 As Umori enters, choose a card type.
//	 Spells you cast of the chosen type cost {1} less to cast."
//
// Companion is a deck-construction condition and a special action, neither
// a card ability: the condition is Umori's entry in the deck package's
// companion conditions (deck/companion_umori.go), and the engine offers the
// {3} action to any revealed companion (ADR 0172).
//
// The choice is Cloud Key's "As this enters, choose <type>" over the words
// of the card, with every card type offered (CR 205.2a); a land can be
// chosen and discounts nothing, since a land is never a spell. Until the
// controller answers, the choice reads as "" and matches nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "913383a1-9c0d-40a1-8c5a-b0a8a95c09a6",
		Name:         "Umori, the Collector",
		Completeness: effects.CompletenessFull,
		AsEnters: effects.ChooseOptionAsEnters("Umori, the Collector",
			"artifact", "battle", "creature", "enchantment", "instant", "kindred", "land", "planeswalker", "sorcery"),
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Spells you cast of the chosen type cost {1} less to cast.",
				effects.YourSpell(), func(q game.CostQuery) bool {
					return effects.RprB03IsOfChosenType(effects.ChosenOptionOf(q.Game, q.Source.InstanceID), q.Card)
				}),
		},
	})
}

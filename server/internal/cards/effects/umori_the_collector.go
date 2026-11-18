package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "913383a1-9c0d-40a1-8c5a-b0a8a95c09a6",
		Name:         "Umori, the Collector",
		Completeness: CompletenessFull,
		AsEnters: ChooseOptionAsEnters("Umori, the Collector",
			"artifact", "battle", "creature", "enchantment", "instant", "kindred", "land", "planeswalker", "sorcery"),
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Spells you cast of the chosen type cost {1} less to cast.",
				YourSpell(), func(q game.CostQuery) bool {
					return rprB03IsOfChosenType(ChosenOptionOf(q.Game, q.Source.InstanceID), q.Card)
				}),
		},
	})
}

package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bag End Banquet — Artifact {6} (The Hobbit Eternal):
//
//	"When this artifact enters, create three Food tokens.
//	 {T}: Add {C} for each Food you control."
//
// The trigger is three Foods (FoodToken, the same template Elanor Gardner
// makes one of). The mana ability is the Notary Hobbits' scaled shape,
// ProducedPerPermanent over the Foods you control: the count is read when the
// ability resolves into mana, so the Banquet's own three tokens and any other
// Food (a Food creature included, since the match reads the effective
// subtypes) are all counted, and an opponent's Food is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2522a9dd-bcf3-4b84-abff-52b9e6dbde34",
		Name:         "Bag End Banquet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bag End Banquet — create three Food tokens",
				effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 3})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedPerPermanent("C", func(c game.Card) bool { return c.HasSubtype("Food") }),
			Label:          "Add {C} for each Food you control",
		}},
	})
}

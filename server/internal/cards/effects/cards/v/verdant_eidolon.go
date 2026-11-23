package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Verdant Eidolon — Creature — Spirit {3}{G}, 2/2 (Dissension, #5598):
//
//	"{G}, Sacrifice this creature: Add three mana of any one color.
//	 Whenever you cast a multicolored spell, you may return this card from
//	 your graveyard to your hand."
//
// The first ability is a mana ability (CR 605.1a): Lotus Bloom's
// three-of-one-colour production behind a {G} and a sacrifice of itself.
// The return is disEidolonReturn, a graveyard trigger (InGraveyard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3acc388-4015-42e4-88ff-b35a91134d12",
		Name:         "Verdant Eidolon",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Mana: "{G}", Sacrifice: true},
			Produced: effects.OneColorOfAmount(3),
			Label:    "{G}, Sacrifice this creature: Add three mana of any one color.",
		}},
		Triggered: []game.TriggeredAbility{effects.DisEidolonReturn("Verdant Eidolon")},
	})
}

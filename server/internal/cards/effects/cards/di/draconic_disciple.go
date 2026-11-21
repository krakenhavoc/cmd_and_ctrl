package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Draconic Disciple — Creature — Human Shaman {1}{R}{G}, 2/2 (Commander
// Starter Decks):
//
//	"{T}: Add one mana of any color.
//	 {7}, {T}, Sacrifice this creature: Create a 5/5 red Dragon creature token
//	 with flying."
//
// Birds of Paradise's mana ability, and a token maker whose {T} and sacrifice
// are costs (CR 118.8): the Disciple is gone before the ability resolves, and
// the Dragon is created whether or not it was the same object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "781cee5f-f9cb-42ed-b408-543d297bf96d",
		Name:         "Draconic Disciple",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}, {T}, Sacrifice this creature: Create a 5/5 red Dragon creature token with flying.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{7}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 red Dragon with flying"), N: 1}),
		}},
	})
}

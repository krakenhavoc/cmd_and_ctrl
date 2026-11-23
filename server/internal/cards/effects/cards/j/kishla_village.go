package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kishla Village — Land (Tarkir: Dragonstorm, #4463):
//
//	"This land enters tapped unless you control an Island or a Swamp.
//	 {T}: Add {G}.
//	 {3}{G}, {T}: Surveil 2. (Look at the top two cards of your library,
//	 then put any number of them into your graveyard and the rest on top
//	 of your library in any order.)"
//
// The TDM village shape (Mistrise Village): the checkland entry over an
// Island or a Swamp and a plain {G}. The surveil is Lilypad Village's
// body with no activation condition.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4eb428ab-f5b0-46ca-98dd-b3466a91ef97",
		Name:         "Kishla Village",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedUnless(effects.YouControlLandTyped("island", "swamp"))},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}, {T}: Surveil 2.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{G}"), effects.TapCost()),
			Effect:  effects.Do(effects.Surveil{N: 2}),
		}},
	})
}

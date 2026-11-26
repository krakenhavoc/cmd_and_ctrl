package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ulvenwald Captive // Ulvenwald Abomination — {1}{G} Creature — Werewolf
// Horror 1/2 // Creature — Eldrazi Werewolf 4/6 (Eldritch Moon, S82 OLD,
// #5541):
//
//	Front: "Defender
//	        {T}: Add {G}.
//	        {5}{G}{G}: Transform this creature."
//	Back:  "{T}: Add {C}{C}."
//
// Defender is a printed keyword and the mana abilities are plain tap
// abilities. The transform is the permanent turning over in place at instant
// speed (Shrill Howler's shape).
//
// No simplification.
func init() {
	const oracle = "0272ca81-e727-4f4b-b06e-072d70bb5558"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Ulvenwald Captive",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}{G}: Transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{5}{G}{G}"),
			Effect:  effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Ulvenwald Abomination",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
	})
}

package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// tangleclawWerewolfOracle is shared by both faces: the front registers under
// it, the back under tangleclawWerewolfOracle + "#1".
const tangleclawWerewolfOracle = "280d5601-802c-4607-9655-75cc83248210"

// Tangleclaw Werewolf // Fibrous Entangler — a transforming creature (Eldritch
// Moon, S82 OLD, #5541):
//
//	Tangleclaw Werewolf {2}{G}{G}, Creature — Werewolf Horror, 2/4
//	  "This creature can block an additional creature each combat.
//	   {6}{G}: Transform this creature."
//
//	Fibrous Entangler, Creature — Eldrazi Werewolf, 4/6
//	  "Vigilance
//	   This creature must be blocked if able.
//	   This creature can block an additional creature each combat."
//
// Spike-Tailed Ceratops' extra block (CR 509.1a) on both faces; the back
// adds Riveteers Decoy's must-be-blocked static. The activation turns this
// creature over in place only if it is still the same face-up permanent
// (CR 701.27f, TransformThis).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     tangleclawWerewolfOracle,
		Name:         "Tangleclaw Werewolf",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.CanBlockAdditional(effects.SelfOnly, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}{G}: Transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{6}{G}"),
			Effect:  effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        tangleclawWerewolfOracle + "#1",
		Name:            "Fibrous Entangler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static:          []game.StaticAbility{effects.MustBeBlocked(), effects.CanBlockAdditional(effects.SelfOnly, 1)},
	})
}

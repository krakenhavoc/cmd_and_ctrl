package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// kessigProwlerOracle is shared by both faces: the front registers under it,
// the back under kessigProwlerOracle + "#1".
const kessigProwlerOracle = "d2aff0ea-5474-4329-ae6a-80ee4908ee8d"

// Kessig Prowler // Sinuous Predator — a transforming creature (Eldritch Moon,
// S82 OLD, #5541):
//
//	Kessig Prowler {G}, Creature — Werewolf Horror, 2/1
//	  "{4}{G}: Transform this creature."
//
//	Sinuous Predator, Creature — Eldrazi Werewolf, 4/4
//	  "This creature can't be blocked by more than one creature."
//
// Tangleclaw Werewolf's transform activation, and Battlefront Krushok's
// block-count rule (MaxBlockers, CR 509.1b) on the back face.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     kessigProwlerOracle,
		Name:         "Kessig Prowler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}: Transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{4}{G}"),
			Effect:  effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     kessigProwlerOracle + "#1",
		Name:         "Sinuous Predator",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.MaxBlockers(effects.OnSelf(), 1)},
	})
}

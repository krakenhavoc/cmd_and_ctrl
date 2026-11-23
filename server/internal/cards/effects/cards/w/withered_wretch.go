package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Withered Wretch — Creature — Zombie Cleric {B}{B}, 2/2 (Commander Starter
// Decks):
//
//	"{1}: Exile target card from a graveyard."
//
// Keen-Eyed Curator's activation without the exile link: the target is any
// card in any graveyard, exiled as the ability resolves, and nothing is done
// if it has left by then (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29e13ec1-a239-4217-a2bd-24ff90745ec0",
		Name:         "Withered Wretch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Exile target card from a graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.ExileTheFirstLegalTarget,
		}},
	})
}

package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Balduvian Conjurer — Creature — Human Wizard {1}{U}, 0/2 (Ice Age, S82 OLD #5636):
//
//	"{T}: Target snow land becomes a 2/2 creature until end of turn. It's still a
//	 land."
//
// Vastwood Animist's animation: a fixed 2/2 with the creature type added.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3eb21641-90e1-4865-b7be-a21d52b7b56b",
		Name:         "Balduvian Conjurer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target snow land becomes a 2/2 creature until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target snow land", effects.Land(), effects.Snow()),
			Effect:  effects.SnowLandAnimateB831,
		}},
	})
}

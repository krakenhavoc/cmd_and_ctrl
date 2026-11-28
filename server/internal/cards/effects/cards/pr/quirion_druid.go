package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quirion Druid — Creature — Elf Druid {2}{G} 1/2 (Visions, S82 OLD, #5631):
//
//	"{G}, {T}: Target land becomes a 2/2 green creature that's still a land.
//	 (This effect lasts indefinitely.)"
//
// The land is a creature with base 2/2 and the colour green, pinned to the
// land object for as long as it stays on the battlefield (Case of the
// Filched Falcon's indefinite change).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cdc0621a-5a88-46a0-90fa-87ba1e7e1eb2",
		Name:         "Quirion Druid",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: Target land becomes a 2/2 green creature that's still a land. (This effect lasts indefinitely.)",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.S82OldB788AnimateTargetLand,
		}},
	})
}

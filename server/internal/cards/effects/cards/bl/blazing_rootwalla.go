package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blazing Rootwalla — Creature — Lizard {R}, 1/1 (Mystery Booster 2):
//
//	"{R}: This creature gets +2/+0 until end of turn. Activate only once
//	 each turn.
//	 Madness {0} (If you discard this card, discard it into exile. When you
//	 do, cast it for its madness cost or put it into your graveyard.)"
//
// Basking Rootwalla's once-a-turn pump (OncePerTurnActivation, CR 602.5b) at
// {R} for +2/+0, and madness for a real {0} (CR 702.35a): a discarded
// Rootwalla is cast from exile for nothing.
//
// No simplification.
func init() {
	const label = "{R}: This creature gets +2/+0 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "d43f5977-0532-4653-bc57-ce29f9a531c9",
		Name:         "Blazing Rootwalla",
		Completeness: effects.CompletenessFull,
		Madness:      "{0}",
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{R}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 0, "Blazing Rootwalla — +2/+0 until end of turn"),
		}},
	})
}

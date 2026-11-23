package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbound Whelp — Artifact Creature — Dragon {3}{R}, 0/0 (Modern
// Horizons 2, #5501):
//
//	"Flying
//	 {R}: This creature gets +1/+0 until end of turn.
//	 Modular 2 (This creature enters with two +1/+1 counters on it. When
//	 it dies, you may put its +1/+1 counters on target artifact
//	 creature.)"
//
// Firefly's firebreathing (the shared thisGetsUntilEndOfTurn body at
// {R}, ending in the cleanup step, CR 514.2) on Arcbound Hybrid's
// modular 2 (game/modular.go, CR 702.43a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "36875030-8094-4e2c-860d-c0820825a24e",
		Name:            "Arcbound Whelp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "modular 2"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Arcbound Whelp — +1/+0 until end of turn"),
		}},
	})
}

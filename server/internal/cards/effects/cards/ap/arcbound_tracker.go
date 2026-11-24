package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbound Tracker — Artifact Creature — Dog {2}{R}, 0/0 (Modern Horizons
// 2, #5501):
//
//	"Menace
//	 Modular 2 (...)
//	 Whenever you cast a spell other than your first spell each turn, put a
//	 +1/+1 counter on this creature."
//
// The trigger reads the per-turn cast tally the way Cosmogrand Zenith's
// "second spell" does, with "two or more" for "other than your first". The
// counter goes on only if the Tracker is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ff86f385-dc3e-4bfe-83e7-9229afab4aa9",
		Name:            "Arcbound Tracker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace", "modular 2"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB37YouCastASpellOtherThanYourFirst,
				"Arcbound Tracker — put a +1/+1 counter on it (a spell other than your first this turn)",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

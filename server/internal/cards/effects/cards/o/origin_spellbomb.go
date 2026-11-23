package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Origin Spellbomb — Artifact {1} (Scars of Mirrodin, #5577):
//
//	"{1}, {T}, Sacrifice this artifact: Create a 1/1 colorless Myr artifact
//	 creature token.
//	 When this artifact is put into a graveyard from the battlefield, you
//	 may pay {W}. If you do, draw a card."
//
// A Spellbomb: the first ability makes the table's Myr token, the second
// is Nihil Spellbomb's draw (somB2SpellbombDraw).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3de0e14f-c917-4c00-8341-1170ac9b3cbc",
		Name:         "Origin Spellbomb",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Create a 1/1 colorless Myr artifact creature token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Myr artifact"), N: 1}),
		}},
		Triggered: []game.TriggeredAbility{effects.SomB2SpellbombDraw("Origin Spellbomb", "{W}")},
	})
}

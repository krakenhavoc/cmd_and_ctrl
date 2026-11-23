package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Panic Spellbomb — Artifact {1} (Scars of Mirrodin, #5577):
//
//	"{T}, Sacrifice this artifact: Target creature can't block this turn.
//	 When this artifact is put into a graveyard from the battlefield, you
//	 may pay {R}. If you do, draw a card."
//
// A Spellbomb: the first ability is Bola Warrior's
// (targetCreatureCantBlockThisTurn), the second Nihil Spellbomb's draw
// (somB2SpellbombDraw).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4429a30e-17f1-4bcc-92f0-55ab1315adff",
		Name:         "Panic Spellbomb",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Target creature can't block this turn.",
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCreatureCantBlockThisTurn("Panic Spellbomb — target creature can't block this turn"),
		}},
		Triggered: []game.TriggeredAbility{effects.SomB2SpellbombDraw("Panic Spellbomb", "{R}")},
	})
}

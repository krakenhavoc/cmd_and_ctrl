package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandstorm Eidolon — Creature — Spirit {3}{R}, 2/2 (Dissension, #5598):
//
//	"{R}, Sacrifice this creature: Target creature can't block this turn.
//	 Whenever you cast a multicolored spell, you may return this card from
//	 your graveyard to your hand."
//
// The restriction is Bola Warrior's targetCreatureCantBlockThisTurn.
// The return is disEidolonReturn, a graveyard trigger (InGraveyard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84fe1c59-3f7b-4c1c-9bc4-78a185cfcc0c",
		Name:         "Sandstorm Eidolon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: Target creature can't block this turn.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Effect:  effects.TargetCreatureCantBlockThisTurn("Sandstorm Eidolon — target creature can't block this turn"),
		}},
		Triggered: []game.TriggeredAbility{effects.DisEidolonReturn("Sandstorm Eidolon")},
	})
}

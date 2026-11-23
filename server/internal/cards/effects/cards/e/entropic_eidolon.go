package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Entropic Eidolon — Creature — Spirit {3}{B}, 2/2 (Dissension, #5598):
//
//	"{B}, Sacrifice this creature: Target player loses 1 life and you gain
//	 1 life.
//	 Whenever you cast a multicolored spell, you may return this card from
//	 your graveyard to your hand."
//
// The drain is Death Cultist's drainTargetOpponentOne.
// The return is disEidolonReturn, a graveyard trigger (InGraveyard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be5f788a-8a2e-447c-9e65-f9fc197ddfc9",
		Name:         "Entropic Eidolon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice this creature: Target player loses 1 life and you gain 1 life.",
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.DrainTargetOpponentOne,
		}},
		Triggered: []game.TriggeredAbility{effects.DisEidolonReturn("Entropic Eidolon")},
	})
}

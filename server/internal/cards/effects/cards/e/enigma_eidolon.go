package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enigma Eidolon — Creature — Spirit {3}{U}, 2/2 (Dissension, #5598):
//
//	"{U}, Sacrifice this creature: Target player mills three cards.
//	 Whenever you cast a multicolored spell, you may return this card from
//	 your graveyard to your hand."
//
// The mill is TargetPlayerMills.
// The return is disEidolonReturn, a graveyard trigger (InGraveyard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4489a526-0e07-4d37-9e83-628f7eb256c6",
		Name:         "Enigma Eidolon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, Sacrifice this creature: Target player mills three cards.",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.TargetPlayerMills(3),
		}},
		Triggered: []game.TriggeredAbility{effects.DisEidolonReturn("Enigma Eidolon")},
	})
}

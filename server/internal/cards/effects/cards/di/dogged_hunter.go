package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dogged Hunter — Creature — Human Nomad {2}{W}, 1/1 (Odyssey, #5613):
//
//	"{T}: Destroy target creature token."
//
// The target is a creature that is a token (CR 111.1); a nontoken
// creature is refused at announce and re-checked as the ability
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf5b3249-619f-48f5-8885-00fd41d24825",
		Name:         "Dogged Hunter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target creature token.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature token", effects.IsTokenPredicate()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}

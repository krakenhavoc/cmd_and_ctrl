package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alms — Enchantment {W} (Weatherlight):
//
//	"{1}, Exile the top card of your graveyard: Prevent the next 1 damage that
//	 would be dealt to target creature this turn."
//
// The cost exiles the topmost card of the activator's graveyard, whatever it is
// (CR 404.2, #6194), and the shield is one point on the target creature that
// lasts until it is used or the turn ends (CR 615.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efd46a6e-a901-4224-8a29-d53a4dc7750c",
		Name:         "Alms",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Exile the top card of your graveyard: Prevent the next 1 damage that would be dealt to target creature this turn.",
			Purpose: effects.PreventsDamage(),
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.ExileTopFromGraveyard(1, "the top card of your graveyard", nil)),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PreventNextOneDamageToTheTarget("Alms"),
		}},
	})
}

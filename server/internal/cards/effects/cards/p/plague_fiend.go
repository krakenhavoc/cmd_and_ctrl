package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plague Fiend — Creature — Insect {1}{B}, 1/1 (Prophecy, S82 OLD b670, #5617):
//
//	"Whenever this creature deals combat damage to a creature, destroy that
//	 creature unless its controller pays {2}."
//
// The shared damaged-creature pay-or-else trigger: a decline destroys the
// creature if it is still the object that was dealt the damage (the trigger
// carries its ID only then); one that already died is not destroyed twice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d287c11e-0684-42ed-abdb-758852ed0502",
		Name:         "Plague Fiend",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB670WheneverThisDealsCombatDamageToACreatureUnlessItsControllerPays(
				"Plague Fiend — destroy that creature unless its controller pays {2}", "{2}",
				"Plague Fiend — pay {2}, or your creature is destroyed?", plagueFiendDestroy),
		},
	})
}

// plagueFiendDestroy is the decline: destroy the damaged creature. Keyed, an
// on-disk identity.
var plagueFiendDestroy = effects.PayUnlessElse("pay-unless/s82-old-b670-plague-fiend-destroy",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		for _, id := range r.Carry {
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

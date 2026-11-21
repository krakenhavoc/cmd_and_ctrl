package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pack Leader — Creature — Dog {1}{W}, 2/2:
//
//	"Other Dogs you control get +1/+1.
//	 Whenever this creature attacks, prevent all combat damage that
//	 would be dealt this turn to Dogs you control."
//
// The anthem is the shared tribal lord over OTHER Dogs you control. The
// attack trigger is #2045's recipient set: permanents you control with
// the Dog subtype, read as the damage would be dealt (CR 611.2c), so a
// Dog that enters or that you gain control of later in the turn is
// protected, a changeling included (it is every creature type). You are
// not protected, and combat damage only.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3701ed34-a97c-4d7b-a15a-4faec02ef24b",
		Name:         "Pack Leader",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Dog"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Pack Leader — prevent all combat damage that would be dealt this turn to Dogs you control",
				effects.Do(effects.PreventDamageFromSource{Protect: effects.ShieldPermanentsYouControl(effects.QuerySubtype("Dog")), CombatOnly: true})),
		},
	})
}

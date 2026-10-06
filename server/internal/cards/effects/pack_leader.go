package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3701ed34-a97c-4d7b-a15a-4faec02ef24b",
		Name:         "Pack Leader",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Dog"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Pack Leader — prevent all combat damage that would be dealt this turn to Dogs you control",
				Do(PreventDamageFromSource{Protect: ShieldPermanentsYouControl(QuerySubtype("Dog")), CombatOnly: true})),
		},
	})
}

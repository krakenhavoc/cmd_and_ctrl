package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Graduation Day — Enchantment {W} (Secrets of Strixhaven, #4456):
//
//	"Repartee — Whenever you cast an instant or sorcery spell that targets a
//	 creature, put a +1/+1 counter on target creature you control."
//
// Repartee is an ability word (CR 207.2c); the ability is one trigger on your
// instant and sorcery casts that target a creature, read from the targets the
// spell was cast with (informedInkwrightCast, the heroic reading): a target
// that is a creature on the battlefield as the spell is cast. Its own target
// is a creature you control, chosen as the trigger goes on the stack: with no
// creature of yours it is not put on the stack at all. It resolves before the
// spell, and the counter lands even if that spell is countered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40e0665d-57b8-4c19-9fb8-b3155d7bbb94",
		Name:         "Graduation Day",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventCast, informedInkwrightCast,
				"Graduation Day — Repartee: a +1/+1 counter on target creature you control",
				PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

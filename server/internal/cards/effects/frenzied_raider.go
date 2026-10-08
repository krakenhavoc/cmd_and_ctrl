package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frenzied Raider — Creature — Demon Berserker {1}{R}, 2/2:
//
//	"Whenever you activate a boast ability, put a +1/+1 counter on this creature."
//
// Watches the activation event (game.Event.Boast, stamped at the
// announcement) rather than the ability list, so it triggers off ANY
// boast ability you activate, another creature's included, and the
// trigger goes on the stack above the ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1207881c-4c03-48e1-9ca7-a2330e8053d1",
		Name:         "Frenzied Raider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventActivateAbility, AllOf(ByYou, ABoastAbility),
				"Frenzied Raider — put a +1/+1 counter on this creature", plusOneCountersOnThis(1)),
		},
	})
}

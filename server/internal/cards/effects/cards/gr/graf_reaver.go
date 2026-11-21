package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Graf Reaver — Creature — Zombie Warrior {1}{B}, 3/3:
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, destroy target
//	 planeswalker.
//	 At the beginning of your upkeep, this creature deals 1 damage to
//	 you."
//
// Exploit is the shared constructor (exploit.go); the payoff is a
// reflexive trigger whose target planeswalker is chosen once a creature
// has really been sacrificed, and with no planeswalker to target it is
// not put on the stack (CR 603.3d). The upkeep trigger is damage the
// Reaver deals, not life loss, so it can be prevented.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8bd41805-5903-4f3f-927e-2391f0ab9478",
		Name:         "Graf Reaver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Graf Reaver", grafReaverDestroyBody),
			effects.AtYourUpkeep("Graf Reaver — this creature deals 1 damage to you", effects.Do(effects.DamageController{N: 1})),
		},
	})
}

// grafReaverDestroyBody is "When this creature exploits a creature,
// destroy target planeswalker." An on-disk key: never renamed, never
// reused.
var grafReaverDestroyBody = game.ReflexiveBody("graf-reaver/destroy", effects.SimpleBody(effects.DestroyChosenPermanent),
	effects.ConstTargets(func() *game.TargetSpec { return effects.TargetPermanent("target planeswalker", effects.Planeswalker()) }))

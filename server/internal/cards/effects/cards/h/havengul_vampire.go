package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Havengul Vampire — Creature — Vampire {3}{R}, 2/2 (Conspiracy: Take the
// Crown, S82 older sets batch 272, #5540):
//
//	"Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it.
//	 Whenever another creature dies, put a +1/+1 counter on this creature."
//
// Two triggers, one each: Abyssal Specter's combat-damage-to-a-player
// watcher, and Orca's "another creature died" (any controller's, the
// Vampire itself excluded and read as it last existed, CR 603.10a). The
// counter is put only while the Vampire is still on the battlefield, so a
// Vampire that died alongside the other creature gets none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "616b3ee0-4678-4562-a21a-5f2d31b9964c",
		Name:         "Havengul Vampire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Havengul Vampire — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B15AnotherCreatureDied(ev, source, g)
				},
				Key:    "Havengul Vampire — another creature died: put a +1/+1 counter on this creature",
				Effect: effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
			},
		},
	})
}

package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scytheclaw — Artifact — Equipment {5}:
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +1/+1.
//	 Whenever equipped creature deals combat damage to a player, that
//	 player loses half their life, rounded up.
//	 Equip {3}"
//
// Batterskull's living weapon and Buster Sword's condition for combat
// damage to a player; the halving is Unstoppable Slasher's, aimed at the
// player the damage event names ("that player").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a75ddd9c-2574-46e9-ac85-58082789ff3e",
		Name:         "Scytheclaw",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.LivingWeapon("Scytheclaw"),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
				},
				Key:    "Scytheclaw — that player loses half their life, rounded up",
				Effect: effects.TriggeringTargetLosesHalfTheirLife,
			},
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}

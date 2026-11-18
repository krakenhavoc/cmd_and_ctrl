package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a75ddd9c-2574-46e9-ac85-58082789ff3e",
		Name:         "Scytheclaw",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			LivingWeapon("Scytheclaw"),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
				},
				Key:    "Scytheclaw — that player loses half their life, rounded up",
				Effect: triggeringTargetLosesHalfTheirLife,
			},
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

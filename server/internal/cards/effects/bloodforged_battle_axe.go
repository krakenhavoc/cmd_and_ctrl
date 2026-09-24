package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodforged Battle-Axe — Artifact — Equipment {1}:
//
//	"Equipped creature gets +2/+0.
//	 Whenever equipped creature deals combat damage to a player,
//	 create a token that's a copy of this Equipment.
//	 Equip {2}"
//
// "A copy of THIS Equipment" is CreateTokenCopy with `Copy` pointed at
// item.SourceCardID (the Axe's own instance) rather than at anything
// captured in Build. The token comes out unattached — the printed
// text says nothing about attaching it — and carries the Axe's own
// oracle ID, so it grows its own combat-damage trigger the moment it
// is equipped to something.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81a7f559-5edd-47ec-91f7-51e5ee998ed0",
		Name:         "Bloodforged Battle-Axe",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Bloodforged Battle-Axe — create a token copy of this Equipment",
					func(g *game.Game, item *game.StackItem) error {
						return CreateTokenCopy{
							Controller: item.Controller,
							Copy:       item.SourceCardID,
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

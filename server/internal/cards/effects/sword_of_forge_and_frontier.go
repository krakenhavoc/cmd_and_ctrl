package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sword of Forge and Frontier — Artifact — Equipment {3} (Phyrexia: All Will
// Be One):
//
//	"Equipped creature gets +2/+2 and has protection from red and from
//	 green.
//	 Whenever equipped creature deals combat damage to a player, exile the
//	 top two cards of your library. You may play those cards this turn. You
//	 may play an additional land this turn.
//	 Equip {2}"
//
// #5479, S82. The Sword frame (Sword of Hearth and Home's pump, layer-6
// protection and combat-damage-to-a-player condition) over Escape to the
// Wilds' two clauses: the impulse exile with the default this-turn
// permission, then Explore's one-shot extra land drop. The extra drop is
// granted whether or not the exiled cards include a land, and whether or not
// the library had two cards to exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6331ce3-21f7-4784-80fe-df1f541a4c46",
		Name:         "Sword of Forge and Frontier",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{ExtraLandDrops: 1},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("protection from red", "protection from green"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Sword of Forge and Frontier — exile the top two cards; you may play them this turn and an additional land",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if _, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, 2, game.CastPermission{}); err != nil {
					return err
				}
				g.GrantAdditionalLandPlayForEffect(item.Controller, 1)
				return nil
			},
		}},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

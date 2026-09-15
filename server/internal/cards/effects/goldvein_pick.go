package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goldvein Pick — Artifact — Equipment for {2} (EDHREC rank 2007):
//
//	"Equipped creature gets +1/+1.
//	 Whenever equipped creature deals combat damage to a player,
//	 create a Treasure token.
//	 Equip {1}"
//
// The cheap ramp Equipment: a one-mana equip that turns any evasive
// body into a mana source, and the Treasure keeps coming as long as
// the creature keeps connecting.
//
// Both halves reuse machinery that is already load-bearing elsewhere:
// attachedCreatureDealtCombatDamageToPlayer is the Swords' and Mask
// of Memory's condition, and TreasureToken is the same template
// fifteen other cards in the catalog create. The only thing this card
// adds is the pairing.
//
// The trigger fires once per combat damage event, so a double-striking
// carrier makes two Treasures — which is the printed behaviour, not a
// leak: first-strike and regular damage are separate damage events.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c1624d10-8838-4af8-aea1-a96c0fe6fd6b",
		Name:         "Goldvein Pick",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Goldvein Pick — create a Treasure", Do(CreateToken{
				Template: TreasureToken(),
				N:        1,
			})),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}

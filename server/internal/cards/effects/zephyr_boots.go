package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zephyr Boots — Artifact — Equipment {1} (EDHREC rank 4495):
//
//	"Equipped creature has flying.
//	 Whenever equipped creature deals combat damage to a player, draw
//	 a card, then discard a card.
//	 Equip {2} ({2}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// One mana to cast, two to equip, and the creature it goes on both
// gets through and loots. The evasion is the point: flying on a
// commander in a deck with no other way to connect turns the Boots
// into a repeatable Faithless Looting, and a graveyard deck wants
// the discard as much as the draw.
//
// Three abilities, all real:
//
//   - The flying is GrantToAttached, a layer 6 grant scoped to
//     whatever the Boots are on right now. It moves with the Boots
//     and falls off when they come off.
//   - The loot is the Swords' trigger condition, narrowed to nothing:
//     equipped creature, combat damage, to a PLAYER. Damage to a
//     planeswalker or another creature does not fire it.
//   - Equip {2} is the standard sorcery-speed attach.
//
// "Draw a card, THEN discard a card" is an ordered pair, not a
// simultaneous swap: the drawn card is in hand and is a legal
// discard. A controller with an empty library draws on an empty
// library and loses at the next state-based check, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02f3b9c3-f611-4f05-ab3e-b296916efdad",
		Name:         "Zephyr Boots",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("flying")},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Zephyr Boots — draw a card, then discard a card", func(g *game.Game, item *game.StackItem) error {
				return b16DrawThenDiscard(g, item, 1, 1)
			}),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

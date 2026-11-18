package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mask of Riddles — Artifact — Equipment {U}{B}:
//
//	"Equipped creature has fear. (It can't be blocked except by artifact
//	 creatures and/or black creatures.)
//	 Whenever equipped creature deals combat damage to a player, you may
//	 draw a card.
//	 Equip {2}"
//
// Rogue's Gloves' draw (the same attachment condition and the "may"
// prompt) with Killer's Mask's keyword grant: fear is read through the
// layered view, so the evasion is the block rule's own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a50c55a7-4e7e-40f6-b76c-1e32598dc9b6",
		Name:         "Mask of Riddles",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("fear")},
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Mask of Riddles — draw a card", Do(DrawCards{N: 1})), "Mask of Riddles — draw a card?", 1),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

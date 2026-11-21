package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a50c55a7-4e7e-40f6-b76c-1e32598dc9b6",
		Name:         "Mask of Riddles",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("fear")},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Mask of Riddles — draw a card", effects.Do(effects.DrawCards{N: 1})), "Mask of Riddles — draw a card?", 1),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

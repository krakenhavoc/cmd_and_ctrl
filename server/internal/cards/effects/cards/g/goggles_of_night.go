package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goggles of Night — Artifact — Equipment {1}{U} (Commander Legends: Battle
// for Baldur's Gate, #5948):
//
//	"Whenever equipped creature deals combat damage to a player, scry 1, then
//	 draw a card.
//	 Equip {2}"
//
// Scry 1 then draw one, in that order and waiting on the scry's prompt: the
// same sequence as True Identity's trigger (LembasScryThenDraw). The trigger
// is the equipped creature's combat damage to a player (CR 510.2), once for
// each player damaged.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af27761d-88df-4d4f-a366-d55ed62d1d33",
		Name:         "Goggles of Night",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1, Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Goggles of Night — scry 1, then draw a card", effects.LembasScryThenDraw),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

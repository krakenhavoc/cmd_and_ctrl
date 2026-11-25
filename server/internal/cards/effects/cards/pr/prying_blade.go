package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prying Blade — Artifact — Equipment {1} (Ixalan, S82 older sets):
//
//	"Equipped creature gets +1/+0.
//	 Whenever equipped creature deals combat damage to a player, create a
//	 Treasure token.
//	 Equip {2}"
//
// Goldvein Pick's trigger with a +1/+0 attached pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5fe86083-1009-4de2-b186-f817ad1cc4e4",
		Name:         "Prying Blade",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Prying Blade — create a Treasure", effects.Do(effects.CreateToken{
				Template: effects.TreasureToken(),
				N:        1,
			})),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}

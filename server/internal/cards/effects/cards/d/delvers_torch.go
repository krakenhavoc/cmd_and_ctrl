package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Delver's Torch — Artifact — Equipment {1}{W}:
//
//	"Equipped creature gets +1/+1.
//	 Whenever equipped creature attacks, venture into the dungeon.
//	 Equip {3}"
//
// The Equipment's controller ventures (the trigger is the Equipment's).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1c8b833-8c74-4f95-8a47-1dd41ef99361",
		Name:         "Delver's Torch",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source)
			}, effects.VentureLabel("Delver's Torch"), effects.Venture),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}

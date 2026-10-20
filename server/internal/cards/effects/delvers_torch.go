package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a1c8b833-8c74-4f95-8a47-1dd41ef99361",
		Name:         "Delver's Torch",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			}, ventureLabel("Delver's Torch"), Venture),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

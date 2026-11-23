package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mace of the Valiant — Artifact — Equipment {2}{W}:
//
//	"Equipped creature gets +1/+1 for each charge counter on this Equipment
//	 and has vigilance.
//	 Whenever a creature you control enters, put a charge counter on this
//	 Equipment.
//	 Equip {3}"
//
// The counter count is read on every layer pass (layer 7c) off the Equipment,
// and vigilance is a layer 6 grant. The trigger fires for every creature that
// enters under its controller's control, token or card, and puts the counter
// through AddCounter, so counter doublers apply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "881c3dd4-49df-48c3-b514-509990332e25",
		Name:         "Mace of the Valiant",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int {
				return source.Counters["charge"]
			}),
			effects.GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl, "Mace of the Valiant — put a charge counter on this Equipment",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
						return nil
					}
					return effects.AddCounter{Target: ctx.Item.SourceCardID, Kind: "charge", N: 1}.Apply(ctx)
				}))),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{3}"),
		},
	})
}

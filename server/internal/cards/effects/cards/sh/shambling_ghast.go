package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shambling Ghast — Creature — Zombie {B}, 1/1 (AFR):
//
//	"When this creature dies, choose one —
//	 • Target creature an opponent controls gets -1/-1 until end of turn.
//	 • Create a Treasure token."
func init() {
	ghast := effects.WhenThisDies("Shambling Ghast — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	ghast.Modes = effects.ChooseOne(
		effects.ModeDoing("Target creature an opponent controls gets -1/-1 until end of turn.",
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: -1, Toughness: -1,
					Label: "Shambling Ghast — -1/-1 until end of turn"}.Apply(ctx)
			}),
		effects.ModeWithPurpose(effects.ModeDoing("Create a Treasure token.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 1}.Apply(ctx)
			}), game.Purpose{Tokens: 1}),
	)
	effects.Register(effects.Spec{
		OracleID:     "e5e6229a-3422-4afd-9a87-86daa2d31d3a",
		Name:         "Shambling Ghast",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{ghast},
	})
}

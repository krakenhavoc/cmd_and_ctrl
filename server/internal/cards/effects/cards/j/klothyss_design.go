package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Klothys's Design — Sorcery {5}{G} (Theros Beyond Death, #5514):
//
//	"Creatures you control get +X/+X until end of turn, where X is your devotion to
//	 green."
//
// X is your devotion to green counted once as the spell resolves (CR 700.5, CR
// 608.2h), and the creatures that get it are the ones you control then (CR
// 611.2c): a creature that enters later is not pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85c20c39-7b9c-4661-9372-e659b321773f",
		Name:         "Klothys's Design",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := effects.DevotionTo(ctx.Game, item.Controller, ctx.TextColor("G"))
			if x <= 0 {
				return nil
			}
			return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: x, Toughness: x, Label: "Klothys's Design — +X/+X"}.Apply(ctx)
		},
	})
}

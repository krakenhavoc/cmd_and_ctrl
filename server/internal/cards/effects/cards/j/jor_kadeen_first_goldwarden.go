package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jor Kadeen, First Goldwarden — Legendary Creature — Human Rebel {R}{W}, 2/2
// (March of the Machine: Phyrexia: All Will Be One Commander):
//
//	"Trample
//	 Whenever Jor Kadeen attacks, it gets +X/+X until end of turn, where X is
//	 the number of equipped creatures you control. Then if Jor Kadeen's power
//	 is 4 or greater, draw a card."
//
// X is counted as the ability resolves (CR 608.2h), over creatures its
// controller controls that have an Equipment attached (isEquippedCreature,
// CR 301.5); Jor Kadeen counts if he is equipped. "Then if" is a check inside
// the resolution and not an intervening "if" (CR 603.4): his power is read
// through the layered view after the pump has applied, so a power of 4 reached
// by the pump draws. A Jor Kadeen that has left the battlefield is neither
// pumped nor drawn for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "967897db-a34b-493c-bd48-04e2650902d2",
		Name:            "Jor Kadeen, First Goldwarden",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Jor Kadeen, First Goldwarden — +X/+X for each equipped creature you control, then draw a card if his power is 4 or greater",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
							return nil
						}
						n := effects.EquippedCreaturesYouControl(ctx.Game, ctx.Item.Controller, false)
						return effects.BoostUntilEOT{
							Target: ctx.Item.SourceCardID, Power: n, Toughness: n,
							Label: "Jor Kadeen, First Goldwarden — +X/+X",
						}.Apply(ctx)
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
							return nil
						}
						if effects.B31CurrentPowerOf(ctx.Game, ctx.Item.SourceCardID) < 4 {
							return nil
						}
						return effects.DrawCards{N: 1}.Apply(ctx)
					}),
				)),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "967897db-a34b-493c-bd48-04e2650902d2",
		Name:            "Jor Kadeen, First Goldwarden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Jor Kadeen, First Goldwarden — +X/+X for each equipped creature you control, then draw a card if his power is 4 or greater",
				Do(
					Instruction(func(ctx *Context) error {
						if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
							return nil
						}
						n := equippedCreaturesYouControl(ctx.Game, ctx.Item.Controller, false)
						return BoostUntilEOT{
							Target: ctx.Item.SourceCardID, Power: n, Toughness: n,
							Label: "Jor Kadeen, First Goldwarden — +X/+X",
						}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
							return nil
						}
						if b31CurrentPowerOf(ctx.Game, ctx.Item.SourceCardID) < 4 {
							return nil
						}
						return DrawCards{N: 1}.Apply(ctx)
					}),
				)),
		},
	})
}

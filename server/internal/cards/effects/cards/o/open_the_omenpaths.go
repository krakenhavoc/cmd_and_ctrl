package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Open the Omenpaths — Instant {2}{R}:
//
//	"Choose one —
//	 • Add two mana of any one color and two mana of any other color. Spend
//	   this mana only to cast creature or enchantment spells.
//	 • Creatures you control get +1/+0 until end of turn."
//
// The first bullet asks for the first colour, adds two mana of it, then asks
// for a second colour that may not be the first (Seal of the Guildpact's
// "choose two colors") and adds two of that, every mana carrying the
// restriction: it pays for a cast of a creature or an enchantment spell only,
// never an ability. The second bullet pumps the creatures its controller
// controls as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ac01998-6e26-4ba5-b66c-614ec831caef",
		Name:         "Open the Omenpaths",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Add two mana of any one color and two mana of any other color. Spend this mana only to cast creature or enchantment spells."),
			effects.Mode("Creatures you control get +1/+0 until end of turn."),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				effects.ChooseColorTail(game.ColorForMana, ctx.Game, item.Controller, item.SourceCardID,
					"Open the Omenpaths — choose the first color for two mana", ctx.Tail(effects.OldB070OmenpathsFirst))
				return nil
			}
			return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1,
				Label: "Open the Omenpaths — creatures you control get +1/+0"}.Apply(ctx)
		},
	})
}

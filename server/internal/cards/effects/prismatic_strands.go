package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prismatic Strands — Instant {2}{W}:
//
//	"Prevent all damage that sources of the color of your choice would
//	 deal this turn.
//	 Flashback—Tap an untapped white creature you control. (You may cast
//	 this card from your graveyard for its flashback cost. Then exile
//	 it.)"
//
// The colour is chosen as the spell resolves, then the shield is ADR 0108
// §7's property shield with no named source (Ethereal Haze's shape) and
// that colour as its property: every source of the colour, to anything,
// for the rest of the turn, rechecked as it would deal the damage (CR
// 615.9), so a source that stops being that colour gets through.
//
// The flashback (CR 702.34a) is ADR 0135 §1's tap alternative cost
// (#2030): no mana, one untapped white creature you control tapped with
// the spell already on the stack. It may have arrived this turn (CR
// 302.6). The card is exiled however it leaves the stack.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:      "8e414ad6-ba19-44a1-a291-2e420734a6dd",
		Name:          "Prismatic Strands",
		Completeness:  CompletenessFull,
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			FlashbackTap(1, "an untapped white creature you control", OfColor("W"), Creature()),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseColorThen(game.ColorForProtection, ctx.Game, item.Controller, item.SourceCardID,
				"Prismatic Strands — choose a color of sources to prevent damage from",
				func(g *game.Game, color string) error {
					if color == "" {
						return nil
					}
					return PreventDamageFromSource{Protect: ShieldAnything, Queries: []game.PermanentQuery{QueryColors(color)}}.Apply(NewContext(g, item))
				})
			return nil
		},
	})
}

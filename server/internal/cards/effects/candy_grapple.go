package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Candy Grapple — Instant {1}{B} (Wilds of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Target creature gets -3/-3 until end of turn. If this spell was
//	 bargained, that creature gets -5/-5 until end of turn instead."
//
// Bargain is the shared optional cost (g2Bargain), paid with the cast and
// read back from the payment record at resolution, as Torch the Tower
// reads it. The shrink is Last Gasp's one until-end-of-turn modify; a
// bargained cast applies -5/-5 in its place, never both. The declared
// purpose is the unbargained -3/-3, the least the spell does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "9588a7fd-bbe5-4a81-9a7e-d8e8f6c0537d",
		Name:          "Candy Grapple",
		Completeness:  CompletenessFull,
		Purpose:       ForTargets(HarmUntilEOT(0, -3, -3)),
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			amount := -3
			if g2WasBargained(ctx) {
				amount = -5
			}
			return BoostUntilEOT{
				Target:    id,
				Power:     amount,
				Toughness: amount,
				Label:     "Candy Grapple — shrink until end of turn",
			}.Apply(ctx)
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_5dn_b2_helpers.go — the shape Fifth Dawn batch 2 shares (#5604).

// dn5B2BoostTheTargetThenScry2 is "Target creature gets +P/+T until end of
// turn. Scry 2." (Ferocious Charge, Stand Firm, Lose Hope): Titan's
// Strength's order, the boost first and the scry last, so the scry's prompt
// has nothing printed after it to hold. A spell that resolves always scries;
// one whose only target became illegal doesn't resolve (CR 608.2b).
func dn5B2BoostTheTargetThenScry2(power, toughness int, label string) func(item *game.StackItem, ctx *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		if id, ok := b16FirstLegalTargetCard(ctx); ok {
			if err := (BoostUntilEOT{Target: id, Power: power, Toughness: toughness, Label: label}).Apply(ctx); err != nil {
				return err
			}
		}
		return Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
	}
}

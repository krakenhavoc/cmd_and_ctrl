package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// team_pump_until_eot.go — the whole OnResolve of "Creatures you
// control get +P/+T until end of turn." (Warrior's Honor; #5630:
// Warrior's Charge, Steadfastness). The creatures you control as it
// resolves are the ones pumped (CR 611.2c), so one that enters later
// this turn is not; a layer 7c change (CR 613.4c) that ends at cleanup
// (CR 514.2). `label` names the effect.
func creaturesYouControlGetUntilEOT(power, toughness int, label string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return BoostUntilEOT{
			Match: And(Creature(), YouControl()), Power: power, Toughness: toughness,
			Label: label,
		}.Apply(ctx)
	}
}

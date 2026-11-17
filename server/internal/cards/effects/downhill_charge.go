package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Downhill Charge — Instant {2}{R}
// (Nemesis, #5618):
//
//	"You may sacrifice a Mountain rather than pay this spell's mana cost.
//	 Target creature gets +X/+0 until end of turn, where X is the number of Mountains you control."
//
// Crash's alternative cost (SacrificeInstead, a Mountain), paid with
// the spell on the stack (CR 601.2h). X is counted as the spell
// resolves, so a Mountain sacrificed for the cost is not counted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8018419c-ea19-499b-a33a-7c20b35bce7b",
		Name:         "Downhill Charge",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0)),
		AlternativeCosts: []game.AlternativeCost{
			SacrificeInstead(1, "a Mountain", 0, HasSubtype("Mountain")),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			x := b08LandsWithSubtypeControlled(ctx.Game, ctx.Controller(), ctx.TextSubtype("Mountain"))
			return BoostUntilEOT{Target: item.Targets[0].ID, Power: x, Label: "Downhill Charge — +X/+0"}.Apply(ctx)
		},
	})
}

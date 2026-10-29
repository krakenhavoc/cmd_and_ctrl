package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Depower — {2}{U} Instant:
//
//	"This spell costs {2} less to cast if it targets an attacking creature.
//	 Target creature gets -4/-0 until end of turn.
//	 Draw a card."
//
// The discount is a self cost modifier that reads the announced target
// (CostsLessIfItTargets, as Bitter Downfall's does); a creature is attacking
// from the declare attackers step until it leaves combat (CR 506.4).
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	Register(Spec{
		OracleID:     "33f89437-885b-4850-88b7-239f5dfcc8b1",
		Name:         "Depower",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(HarmUntilEOT(0, -4, 0)),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets an attacking creature.",
				And(Creature(), attackingCreature())),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: -4, Label: "Depower — -4/-0"}).Apply(ctx); err != nil {
					return err
				}
				break
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// March of Swirling Mist — Instant {X}{U}:
//
//	"As an additional cost to cast this spell, you may exile any number
//	 of blue cards from your hand. This spell costs {2} less to cast for
//	 each card exiled this way.
//	 Up to X target creatures phase out. (While they're phased out,
//	 they're treated as though they don't exist. Each one phases in
//	 before its controller untaps during their next untap step.)"
//
// The phase-out is PhaseOut over every still-legal target in one call,
// so the creatures and everything attached to them leave together
// (CR 702.26a, 702.26g) — Clever Concealment's shape.
//
// The additional cost (#3175) is ExileAnyNumberFromHandCost: any number
// of blue cards from the caster's hand, none included, named at announce
// (CR 601.2b) and exiled with the spell on the stack (CR 601.2h). The
// count is CostQuery.ExilingFromHand, and CostsLessPerExiledFromHand takes
// {2} of generic mana off per card at CR 601.2f — the announced X is
// generic mana a reduction may eat (#2701), the {U} never. The X on the
// stack stays the announced X, so exiling two cards casts an X = 4 March
// for {U}.
//
// "Up to X targets" is CountFromX with UpToX: the announced X is the
// ceiling and fewer targets are legal (#1738).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "debc69ea-372a-4720-838a-16856cd50b07",
		Name:           "March of Swirling Mist",
		Completeness:   CompletenessFull,
		AdditionalCost: ExileAnyNumberFromHandCost("blue cards", "U"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessPerExiledFromHand("{2}", "This spell costs {2} less to cast for each card exiled this way"),
		},
		Targets: marchOfSwirlingMistTargets(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PhaseOut{Targets: legalTargetCards(item, ctx.Game)}.Apply(ctx)
		},
	})
}

// marchOfSwirlingMistTargets is "up to X target creatures".
func marchOfSwirlingMistTargets() *game.TargetSpec {
	spec := TargetCreature("up to X target creatures")
	spec.CountFromX = true
	spec.UpToX = true
	return spec
}

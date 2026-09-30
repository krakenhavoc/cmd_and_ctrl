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
// One declared simplification, weaker than printed:
//
//   - The optional exile-blue-cards discount is not offered. An
//     additional cost that exiles a variable number of cards from hand
//     and reduces the price per card has no shape (no Spec slot pairs
//     an exile-from-hand cost with a per-card reduction), so the spell
//     always costs its full {X}{U}.
//   - "Up to X targets" is CountFromX with UpToX: the announced X is the
//     ceiling and fewer targets are legal (#1738).
func init() {
	Register(Spec{
		OracleID:     "debc69ea-372a-4720-838a-16856cd50b07",
		Name:         "March of Swirling Mist",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You can't exile blue cards from your hand to make it cheaper, so it always costs its full price.",
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

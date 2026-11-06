package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Behind the Mask — Instant {U} (Murders at Karlov Manor, #5193):
//
//	"As an additional cost to cast this spell, you may collect evidence 6.
//	 (Exile cards with total mana value 6 or greater from your graveyard.)
//	 Until end of turn, target artifact or creature becomes an artifact
//	 creature with base power and toughness 4/3. If evidence was collected,
//	 it has base power and toughness 1/1 until end of turn instead."
//
// The animation is Living Brain's (BecomeCreatureUntilEOT: artifact and
// creature added at layer 4, the base power and toughness set at layer 7b,
// pinned to the object), so counters and pumps still apply on top. With
// evidence collected (ctx.EvidenceCollected, the announcement, CR 607.2i)
// the base is 1/1 instead. A creature that is already an artifact creature
// only has its base set.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "85ac3efc-5e19-4661-b52f-116f57e39581",
		Name:          "Behind the Mask",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{OptionalCollectEvidence(6)},
		Targets:       TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			power, toughness, label := 4, 3, "Behind the Mask — a 4/3 artifact creature until end of turn"
			if ctx.EvidenceCollected() {
				power, toughness, label = 1, 1, "Behind the Mask — a 1/1 artifact creature until end of turn"
			}
			return BecomeCreatureUntilEOT{Target: t.ID, SetPower: power, SetToughness: toughness, Label: label}.Apply(ctx)
		},
	})
}

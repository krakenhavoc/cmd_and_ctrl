package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Filigree Vector — Artifact Creature — Phyrexian Construct {3}{W}, 1/1
// (March of the Machine Commander):
//
//	"When this creature enters, put a +1/+1 counter on each of any number
//	 of target creatures and a charge counter on each of any number of
//	 target artifacts.
//	 {1}, {T}, Sacrifice another artifact: Proliferate."
//
// Two target clauses, each "any number" (zero allowed), chosen as the
// trigger goes on the stack and re-checked as it resolves (CR 608.2b): a
// target that left or stopped being a creature or artifact gets nothing,
// and the other clause is unaffected. A permanent that is both an artifact
// and a creature may be chosen in both clauses and gets both counters. The
// activated ability needs the Vector to have been under its controller's
// control since the turn began (CR 302.6) and another artifact to
// sacrifice as a cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "44b90948-e85a-4f69-a9e2-aba63775ba40",
		Name:         "Filigree Vector",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Filigree Vector — a +1/+1 counter on each of any number of target creatures and a charge counter on each of any number of target artifacts",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.ClauseTargets(0) {
							if !ctx.IsTargetLegal(t) {
								continue
							}
							if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						for _, t := range ctx.ClauseTargets(1) {
							if !ctx.IsTargetLegal(t) {
								continue
							}
							if err := (AddCounter{Target: t.ID, Kind: game.CounterCharge, N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}),
				Clauses(
					TargetCreature("any number of target creatures").WithCount(0, 0),
					TargetPermanent("any number of target artifacts", Artifact()).WithCount(0, 0),
				)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice another artifact: Proliferate.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump},
			Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeAnotherN(1, "another artifact", Artifact())),
			Effect:  func(g *game.Game, item *game.StackItem) error { return Proliferate{}.Apply(NewContext(g, item)) },
		}},
	})
}

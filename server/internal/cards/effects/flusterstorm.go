package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flusterstorm — Instant {U} (edhrec_rank 307):
//
//	"Counter target instant or sorcery spell unless its controller
//	 pays {1}.
//	 Storm (When you cast this spell, copy it for each spell cast
//	 before it this turn. You may choose new targets for the copies.)"
//
// The card #1238 was filed against, from roadmap batch 02 (#295):
// half of it — CounterUnlessPaid — has existed since #951, and storm
// was the only reason it could not be written down. It is now the
// two-line Spec that issue predicted.
//
// Why it is the storm card that matters most at a Commander table:
// each copy is its OWN "unless its controller pays {1}", so a storm
// count of three is four separate taxes on (possibly) four different
// spells. Paying for one does not pay for the next, and the copies
// resolve top-down with the victim answering each in turn — which is
// exactly what the engine's per-copy re-target prompt plus
// CounterUnlessPaid's own halt (#951: the guarded spell may not
// resolve while its tax is unanswered) produce, with nothing card-
// specific.
//
// The counter half is a PAY-UNLESS, not a counter, so a spell printed
// "can't be countered" is still a legal target: the prompt is offered
// and the counter simply does nothing (CR 701.6a), which is
// observably different from Flusterstorm fizzling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86bf58f2-7f25-4e10-b797-25e0e8e67769",
		Name:         "Flusterstorm",
		Completeness: CompletenessFull,
		Targets:      instantOrSorcerySpell("target instant or sorcery spell"),
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{1}",
				Question: "Flusterstorm — pay {1} or your spell is countered",
			}.Apply(ctx)
		},
	})
}

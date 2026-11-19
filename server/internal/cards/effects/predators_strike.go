package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Predator's Strike — Instant {1}{G} (Foundations Jumpstart):
//
//	"Target creature gets +3/+3 and gains trample until end of turn."
//
// The +3/+3 is a layer 7c modification and the trample a layer 6 grant,
// both until the cleanup step (CR 514.2) and both on the one target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f62e9a9f-9989-4a62-b063-cca32e21fe8b",
		Name:         "Predator's Strike",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(BenefitUntilEOT(0, 3, 3, "trample"))},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Predator's Strike — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Predator's Strike — trample"}.Apply(ctx)
		},
	})
}

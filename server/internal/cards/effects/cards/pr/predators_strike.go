package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Predator's Strike — Instant {1}{G} (Foundations Jumpstart):
//
//	"Target creature gets +3/+3 and gains trample until end of turn."
//
// The +3/+3 is a layer 7c modification and the trample a layer 6 grant,
// both until the cleanup step (CR 514.2) and both on the one target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f62e9a9f-9989-4a62-b063-cca32e21fe8b",
		Name:         "Predator's Strike",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "trample"))},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Predator's Strike — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Predator's Strike — trample"}.Apply(ctx)
		},
	})
}

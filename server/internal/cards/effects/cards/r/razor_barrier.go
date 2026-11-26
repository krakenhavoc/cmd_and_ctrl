package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Razor Barrier — Instant {1}{W} (Duel Decks: Elspeth vs. Tezzeret, #5578):
//
//	"Target permanent you control gains protection from artifacts or from the
//	 color of your choice until end of turn."
//
// One target, "you control" judged at announce and at resolution. As it
// resolves its controller picks one of six answers, artifacts or one of the
// five colours, in a single option prompt, and the permanent gains that
// protection until end of turn (CR 702.16, CR 611.2). A target that left in
// response gets no prompt (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "669701a5-fd7a-43f8-93f4-da7a8c4dd3dc",
		Name:         "Razor Barrier",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent you control", effects.YouControl()),
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			options := make([]game.ChoiceOption, 0, len(razorBarrierProtections))
			for _, p := range razorBarrierProtections {
				options = append(options, game.ChoiceOption{Label: p.label})
			}
			return effects.PickOption{
				Question: "Razor Barrier — protection from what until end of turn?",
				Options:  options,
				ThenKey:  razorBarrierAnswer,
				Carry:    []uuid.UUID{target.ID},
			}.Apply(ctx)
		},
	})
}

// razorBarrierProtections are the answers in the order the prompt lists them.
var razorBarrierProtections = []struct{ label, keyword string }{
	{"Artifacts", "protection from artifacts"},
	{"White", game.ProtectionFromColor("W")},
	{"Blue", game.ProtectionFromColor("U")},
	{"Black", game.ProtectionFromColor("B")},
	{"Red", game.ProtectionFromColor("R")},
	{"Green", game.ProtectionFromColor("G")},
}

// razorBarrierAnswer grants the chosen protection. An on-disk key: never
// renamed, never reused.
var razorBarrierAnswer = effects.OptionPickThen("option-pick/razor-barrier-protection", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(razorBarrierProtections) || len(r.Carry) == 0 {
		return nil
	}
	p := razorBarrierProtections[r.Index]
	return effects.GrantKeywordUntilEOT{Target: r.Carry[0], Keywords: []string{p.keyword}, Label: "Razor Barrier — " + p.keyword + " until end of turn"}.Apply(ctx)
})

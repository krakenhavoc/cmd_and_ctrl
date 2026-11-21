package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Toils of Night and Day — Instant — Arcane {2}{U} (Betrayers of
// Kamigawa, #5602):
//
//	"You may tap or untap target permanent, then you may tap or untap
//	 another target permanent."
//
// Hidden Strings without cipher: two different target permanents (one
// clause of count two, so the same permanent can't be both), in announce
// order. For each still-legal one in turn (CR 608.2b) the caster picks
// tap it, untap it or leave it, and the second question is asked only
// once the first is answered ("then"), through Hidden Strings' shared
// body (mayTapOrUntapEachInTurn) under this card's own keyed answer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a50be01d-821f-4fdc-b12a-4a1a76432a21",
		Name:         "Toils of Night and Day",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent, then another target permanent").WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			return effects.MayTapOrUntapEachInTurn(ctx, "Toils of Night and Day", toilsOfNightAndDayAnswer, ids)
		},
	})
}

// toilsOfNightAndDayAnswer is the answer about the first target, then the
// question about the next, keyed (#4253): the targets still to be asked
// about ride the prompt as Carry. Its key is an on-disk identity: never
// renamed, never reused. Assigned in init for Hidden Strings' reason (the
// continuation names its own key).
var toilsOfNightAndDayAnswer game.OptionPickThen

func init() {
	toilsOfNightAndDayAnswer = effects.OptionPickThen("option-pick/toils-of-night-and-day-tap-or-untap",
		func(ctx *effects.Context, r game.OptionPicked) error {
			return effects.TapOrUntapEachAnswered(ctx, r, "Toils of Night and Day", toilsOfNightAndDayAnswer)
		})
}

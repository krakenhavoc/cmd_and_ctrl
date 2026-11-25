package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spell Shrivel — Instant {2}{U} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Counter target spell unless its controller pays {4}. If that spell is
//	 countered this way, exile it instead of putting it into its owner's
//	 graveyard."
//
// Reject's pay-unless prompt with its exile destination, on any spell and for
// {4}: the destination applies only when the spell really is countered, so a
// spell whose controller pays resolves and one that can't be countered is left
// alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bfaf376f-90f9-45d9-bfbe-dd84a2a4688b",
		Name:            "Spell Shrivel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		Targets:         effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  targets[0].ID,
				Cost:     "{4}",
				Question: "Spell Shrivel — pay {4} or your spell is countered and exiled",
				Dest:     game.ZoneRef{Kind: game.ZoneExile},
			}.Apply(ctx)
		},
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abstruse Appropriation — Instant {2}{W}{B} (Modern Horizons 3, #5948):
//
//	"Devoid (This card has no color.)
//	 Exile target nonland permanent. You may cast that card for as long as
//	 it remains exiled, and you may spend colorless mana as though it were
//	 mana of any color to cast that spell."
//
// Hurl Through Hell's exile-and-grant in one call, so a token or a
// commander that went elsewhere gets nothing, with the duration of
// the WhileInZone window. The spend clause is narrower than "mana of
// any color": only colorless mana stands in for a colored symbol
// (CastPermission.AnyColorFrom = "C"), other mana pays as printed. Devoid
// is the engine's colorless-by-rule keyword (CR 702.114).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "61a23402-3244-4b5e-9134-98e1f82eda65",
		Name:            "Abstruse Appropriation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets:         TargetPermanent("target nonland permanent", Not(Land())),
		Purpose:         game.Purpose{Answers: game.AnswerRemove}, // exiles a permanent
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return exileAllWithPermission(ctx.Game, []uuid.UUID{item.Targets[0].ID}, game.CastPermission{
				Player:       item.Controller,
				CastOnly:     true,
				AnyColorFrom: "C",
				Duration:     game.WhileInZoneDuration(),
				Label:        "Abstruse Appropriation — cast it for as long as it remains exiled",
			})
		},
	})
}

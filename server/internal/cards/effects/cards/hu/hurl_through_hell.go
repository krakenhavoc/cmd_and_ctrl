package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hurl Through Hell — {2}{B}{R} Instant (Forgotten Realms Commander):
//
//	"Exile target creature. Until the end of your next turn, you may cast
//	 that card and you may spend mana as though it were mana of any color
//	 to cast that spell."
//
// The exile and the grant are one call (ExileCardWithPermissionForEffect, the
// Cryptic Pursuit shape): the permission lands on the card only once it
// reached exile, so a commander that went to the command zone instead (CR
// 903.9) or a token (which ceases to exist) gets nothing. The caster, not the
// owner, holds it; "cast" is CastOnly, so an exiled creature card is never
// "played" as a land; the window is Bruse Tarl's seat-turn one, which ends with
// the CASTER's next turn (ADR 0063).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4254c048-7207-4ca0-bd20-ebb418d9ed1a",
		Name:         "Hurl Through Hell",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // exiles a creature
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ctx.Game.ExileCardWithPermissionForEffect(item.Targets[0].ID, game.CastPermission{
				Player:   item.Controller,
				CastOnly: true,
				AnyColor: true,
				Duration: ctx.Game.UntilEndOfYourNextTurnDuration(item.Controller),
				Label:    "Hurl Through Hell — cast it until the end of your next turn",
			})
		},
	})
}

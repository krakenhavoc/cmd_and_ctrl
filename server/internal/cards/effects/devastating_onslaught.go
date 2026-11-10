package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Devastating Onslaught — Sorcery {X}{X}{R} (Edge of Eternities):
//
//	"Create X tokens that are copies of target artifact or creature you
//	 control. Those tokens gain haste until end of turn. Sacrifice them at
//	 the beginning of the next end step."
//
// Saheeli, the Sun's Brilliance's shape with X copies: one creation makes
// all X tokens at once (CR 608.2f), so each sees the others enter and
// "enters" abilities of the original trigger for every one of them; the
// tail then gives each token haste until end of turn (a duration, not part
// of the copiable values, so a copy of a token does not have it) and
// schedules ONE delayed trigger for the next end step that sacrifices
// exactly those tokens. A token that is already gone is skipped (CR
// 603.7c). X is announced as the spell is cast; X = 0 makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e17cd4c-ebf0-40d0-b52d-b4d84bbcf7b3",
		Name:         "Devastating Onslaught",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetPermanent("target artifact or creature you control", Or(Artifact(), Creature()), YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			original := FirstLegalBattlefieldTarget(ctx)
			if original == uuid.Nil || ctx.X() <= 0 {
				return nil
			}
			return CreateTokenCopiesTogether{
				Controller: item.Controller,
				Copies:     []uuid.UUID{original},
				N:          ctx.X(),
				Then:       ctx.ItemTail(devastatingOnslaughtRest),
			}.Apply(ctx)
		},
	})
}

// devastatingOnslaughtRest is haste until end of turn for each token and
// the sacrifice at the beginning of the next end step. An on-disk key:
// never renamed, never reused.
var devastatingOnslaughtRest = CardTail("devastating-onslaught/haste-and-sacrifice-copies", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
			Duration: ctx.Game.UntilEndOfTurnDuration(), Label: "Devastating Onslaught — the token gains haste until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Item.Controller,
		Label:      "Devastating Onslaught — sacrifice the tokens",
		Cards:      o.IDs,
		Body:       sacrificeListedCardsBody,
	}.Apply(ctx)
})

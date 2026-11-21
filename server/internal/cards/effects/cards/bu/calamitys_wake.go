package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Calamity's Wake — Instant {1}{W} (The Brothers' War, #5482):
//
//	"Exile all graveyards. Players can't cast noncreature spells this
//	 turn. Exile Calamity's Wake."
//
// Three instructions in printed order (ResolveSteps, as All Suns' Dawn
// exiles itself last). Exile all graveyards is every seat's, yours too
// (Bazaar of Wonders' b02ExileAllGraveyards). The ban is a cast ban on every
// seat that has the CR 611.2 duration "this turn" and a noncreature filter,
// so a creature spell and a land play are untouched and the ban ends in the
// cleanup step. The spell then exiles itself instead of going to the
// graveyard: it is a card on the stack as it resolves, and a copy of it
// (not a card) exiles nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6436330-f35a-4d31-9230-03e0c62f9e7e",
		Name:         "Calamity's Wake",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.B02ExileAllGraveyards(ctx.Game, ctx.Item) }),
			effects.Instruction(func(ctx *effects.Context) error {
				for _, p := range ctx.Game.Seats {
					if p == nil {
						continue
					}
					if err := (effects.RestrictCasting{
						Player:   p.ID,
						Rule:     game.CastBanRule{Kind: game.CastBanOutright, Filter: game.PermissionFilter{NoncreatureOnly: true}},
						Label:    "Calamity's Wake — can't cast noncreature spells this turn",
						Duration: effects.DurationUntilEndOfTurn(ctx),
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.ExileTarget{Target: ctx.Source()}.Apply(ctx) }),
		),
	})
}

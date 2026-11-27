package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sentry Oak — Creature — Treefolk Warrior {4}{W}, 3/5 (Lorwyn, #5593):
//
//	"Defender
//	 At the beginning of combat on your turn, you may clash with an
//	 opponent. If you win, this creature gets +2/+0 and loses defender
//	 until end of turn. (Each clashing player reveals the top card of their
//	 library, then puts that card on their choice of the top or bottom. A
//	 player wins if their card had a greater mana value.)"
//
// Defender is a printed keyword. The beginning-of-combat trigger is an
// optional one, asked as it resolves. On a win the Oak gets +2/+0 and loses
// defender (a layer 6 removal with its own timestamp, as Canopy Claws'
// flying) until end of turn, while it is still the same permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1b2e0a60-844e-4db7-91fd-323497ab3cbf",
		Name:            "Sentry Oak",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtBeginningOfYourCombat("Sentry Oak — clash with an opponent; if you win, +2/+0 and loses defender until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ClashWithAnOpponent{Then: ctx.ItemTail(sentryOakClashed)}.Apply(ctx)
				}), "Sentry Oak — clash with an opponent?"),
		},
	})
}

// sentryOakClashed is "If you win, this creature gets +2/+0 and loses
// defender until end of turn". An on-disk key.
var sentryOakClashed = effects.CardTail("then/sentry-oak-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.WonTheClash(o) || !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	src := ctx.Source()
	if err := (effects.BoostUntilEOT{Target: src, Power: 2, Label: "Sentry Oak — +2/+0"}).Apply(ctx); err != nil {
		return err
	}
	return effects.UntilEndOfTurn(ctx, src, nil, "Sentry Oak — loses defender until end of turn", game.RemoveKeywordsMod("defender"))
})

package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Springjack Knight — Creature — Kithkin Knight {2}{W}, 2/1 (Lorwyn, #5593):
//
//	"Whenever this creature attacks, clash with an opponent. If you win,
//	 target creature gains double strike until end of turn. (Each clashing
//	 player reveals the top card of their library, then puts that card on
//	 their choice of the top or bottom. A player wins if their card had a
//	 greater mana value.)"
//
// The ability's one target is chosen as it goes on the stack, before the
// clash (CR 603.3d); on a win it is checked again (CR 608.2b) and gains
// double strike until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97b0b719-d948-4671-97ba-1b2be8cc55e4",
		Name:         "Springjack Knight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Springjack Knight — clash with an opponent; if you win, target creature gains double strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ClashWithAnOpponent{Then: ctx.ItemTail(springjackKnightClashed)}.Apply(ctx)
				}), effects.TargetCreature("target creature")),
		},
	})
}

// springjackKnightClashed is "If you win, target creature gains double
// strike until end of turn". An on-disk key.
var springjackKnightClashed = effects.CardTail("then/springjack-knight-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.WonTheClash(o) {
		return nil
	}
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Springjack Knight — double strike"}.Apply(ctx)
})

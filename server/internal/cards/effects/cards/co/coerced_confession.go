package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coerced Confession — Sorcery {4}{U/B} (Gatecrash, #5565):
//
//	"Target player mills four cards. You draw a card for each creature card
//	 put into their graveyard this way."
//
// The mill is the keyed form whose tail is told the cards that LANDED in
// the graveyard (CR 400.7): a card a replacement sent elsewhere, or left in
// the library, is not counted. The creature cards among them are counted
// as the tail runs and the caster (not the milled player) draws that many.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49b84c50-ca0a-4947-8225-ae41aa08e1f1",
		Name:         "Coerced Confession",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				return ctx.Game.MillToZoneKeyedForEffect(t.ID, 4, game.ZoneGraveyard, game.MillUntil{}, ctx.ItemTail(coercedConfessionDraw))
			}
			return nil
		},
	})
}

// coercedConfessionDraw draws a card for each creature card the mill put
// into the graveyard. An on-disk key: never renamed, never reused.
var coercedConfessionDraw = effects.CardTail("mill-then/coerced-confession-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	n := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return effects.DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
})

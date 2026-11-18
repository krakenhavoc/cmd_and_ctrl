package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b19_helpers.go — shared bodies for the S82 reprint pass,
// batch 19 (#5908).

// rprB19OpponentDrewNotFirstInDrawStep is Xyris's trigger condition:
// "whenever an opponent draws a card except the first one they draw in
// each of their draw steps" — the draw half of Orcish Bowmasters'
// condition. Only the first completed draw of an opponent's own draw
// step is exempt.
func rprB19OpponentDrewNotFirstInDrawStep(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventDrawCard && ByAnOpponent(ev, source, ch, g) &&
		!g.FirstCompletedDrawInOwnDrawStepForEffect(ev.Actor)
}

// rprB19XHydraKey is the step record Zaxara's trigger keeps between
// creating the Hydra and putting the counters on it.
const rprB19XHydraKey = "zaxara-the-exemplary/hydra"

// rprB19HydraWithXCountersSteps is "create a 0/0 green Hydra creature
// token, then put X +1/+1 counters on it" for a cast trigger: X is the
// announced value of the spell that triggered it, read from the spell on
// the stack or its last-known information (CR 608.2h); the counters go on
// the Hydra the creation step made (Fractal Anomaly's two steps).
func rprB19HydraWithXCountersSteps() []Applier {
	return []Applier{
		Instruction(func(ctx *Context) error {
			x, _ := ctx.Game.SpellXValueForEffect(ctx.Trigger().Event.CardID)
			ctx.RecordStep(rprB19XHydraKey, nil, int(b25LastEventSeq(ctx.Game)), x)
			_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("0/0 green Hydra")), 1, game.TokenEntryOptions{})
			return err
		}),
		Instruction(func(ctx *Context) error {
			rec := ctx.StepRecord(rprB19XHydraKey)
			if len(rec.Ns) < 2 || rec.Ns[1] <= 0 {
				return nil
			}
			for _, id := range b27TokensCreatedByAfter(ctx.Game, ctx.Controller(), uint64(rec.Ns[0])) {
				c, ok := ctx.Game.LookupCardForEffect(id)
				if !ok || !onBattlefield(ctx.Game, id) || !c.IsToken() || c.Name != "Hydra" {
					continue
				}
				if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: rec.Ns[1]}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			return nil
		}),
	}
}

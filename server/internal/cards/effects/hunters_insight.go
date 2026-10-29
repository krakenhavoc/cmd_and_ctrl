package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunter's Insight — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "cf8a0782-f545-445e-abac-d9063dad696b", Name: "Hunter's Insight", Completeness: CompletenessFull,
		Targets: TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			ref, ok := ctx.Game.PermanentRefForEffect(id)
			if !ok {
				return nil
			}
			return DelayedOnEvent{Label: "Hunter's Insight — draw that many cards", On: []game.EventKind{game.EventDealDamage}, Condition: huntersInsightHit, CondParams: game.EffectParams{Object: ref}, Body: huntersInsightDraw, Repeats: true}.Apply(ctx)
		},
	})
}

var huntersInsightHit = game.DelayedCondition("combat/hunters-insight-creature-hit-player-or-planeswalker", func(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Source != p.Object.ID {
		return false
	}
	ref, ok := g.PermanentRefForEffect(ev.Source)
	if !ok || ref.Epoch != p.Object.Epoch {
		return false
	}
	if g.PlayerByIDForEffect(ev.Target) != nil {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && c.IsPlaneswalker()
})
var huntersInsightDraw = game.SimpleDelayedBody("hunters-insight/draw-that-much", func(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return DrawCards{N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
})

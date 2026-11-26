package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aurelia's Fury — Instant {X}{R}{W} (Gatecrash, S82 OLD #5565):
//
//	"Aurelia's Fury deals X damage divided as you choose among any number of
//	 targets. Tap each creature dealt damage this way. Players dealt damage
//	 this way can't cast noncreature spells this turn."
//
// Rolling Thunder's divided damage (CR 601.2d), then the two riders read what
// the damage actually did: only damage that landed counts, so a target that
// prevented it all is neither tapped nor barred. Each creature still on the
// battlefield that took damage is tapped; each player who took damage can't
// cast noncreature spells until end of turn (Permission Denied's ban).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ceaa0e02-2dec-412d-86aa-c1129fd0fb08",
		Name:         "Aurelia's Fury",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetAny().WithCount(0, 0).Dividing(effects.DivideX()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			cursor := effects.B25LastEventSeq(ctx.Game)
			if err := effects.DealDividedDamage(ctx); err != nil {
				return err
			}
			var hit []game.Event
			for _, ev := range ctx.Game.Events {
				if ev.Seq > cursor && ev.Kind == game.EventDealDamage && ev.Source == ctx.Source() && ev.Amount > 0 {
					hit = append(hit, ev)
				}
			}
			for _, ev := range hit {
				if p := ctx.Game.PlayerByIDForEffect(ev.Target); p != nil {
					if err := (effects.RestrictCasting{
						Player:   p.ID,
						Rule:     game.CastBanRule{Kind: game.CastBanOutright, Filter: game.PermissionFilter{NoncreatureOnly: true}},
						Label:    "Aurelia's Fury — can't cast noncreature spells this turn",
						Duration: effects.DurationUntilEndOfTurn(ctx),
					}).Apply(ctx); err != nil {
						return err
					}
					continue
				}
				if c, ok := ctx.Game.LookupCardForEffect(ev.Target); ok && c.IsCreature() {
					if err := (effects.TapTarget{Target: ev.Target}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}

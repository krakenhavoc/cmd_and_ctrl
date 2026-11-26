package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S82OldB398PingTheBlocker is "this creature deals N damage to that creature"
// for a "becomes blocked by a creature" trigger (CR 509.3d: one per blocker):
// the damage goes to that blocker only if it is still on the battlefield as the
// ability resolves.
func S82OldB398PingTheBlocker(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil || item.Trigger.Object == nil {
			return nil
		}
		o := item.Trigger.Object
		if _, ok := g.PermanentForEffect(game.ObjectRef{ID: o.ID, Epoch: o.Epoch}); !ok {
			return nil
		}
		return DealDamage{Source: item.SourceCardID, Target: o.ID, Amount: n}.Apply(NewContext(g, item))
	}
}

// S82OldB398ThreatenTarget is "Gain control of target creature until end of
// turn. Untap that creature. It gains haste until end of turn." for the
// card `name`: the target's control, untap and haste, in printed order, if the
// target is still legal (CR 608.2b).
func S82OldB398ThreatenTarget(name string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (GainControl{
				Target:   t.ID,
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    name + " — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   t.ID,
				Keywords: []string{"haste"},
				Label:    name + " — haste until end of turn",
			}.Apply(ctx)
		}
		return nil
	}
}

// S82OldB398ThisDealtAtLeastToAnOpponent is "this creature deals N or more
// damage to an opponent": one damage event from the source to an opponent
// of its controller, of at least n (CR 603.2).
func S82OldB398ThisDealtAtLeastToAnOpponent(n int) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Amount >= n && ThisDealtDamageToAnOpponent(ev, source, game.Characteristic{}, g)
	}
}

// S82OldB398DamageEachClauseTarget is "this deals n damage to <the first
// target> and n damage to <the second target>" as one damage instance: each
// target clause is rechecked on its own (CR 608.2b), so a gone target leaves the
// other one hit.
func S82OldB398DamageEachClauseTarget(ctx *Context, n int) error {
	return ctx.Game.DamageInstanceForEffect(func() error {
		for slot := 0; slot < 2; slot++ {
			if t, ok := ctx.ClauseTarget(slot); ok {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}).Apply(ctx); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// S82OldB398DestroyLegalTarget is "destroy target <permanent>" for the one
// target of a triggered ability: it is destroyed if it is still legal as the
// ability resolves (CR 608.2b).
func S82OldB398DestroyLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	legal := ctx.LegalTargets()
	for i := range legal {
		if legal[i].Kind == game.TargetCard {
			return DestroyTarget{Target: legal[i].ID}.Apply(ctx)
		}
	}
	return nil
}

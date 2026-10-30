package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc13_helpers.go — shared bodies for Marvel Super Heroes Commander,
// batch 13 (#4455).

// msc13CreaturesWithPowerAtLeast counts the creatures `controller` controls
// with power n or greater, as the effect resolves (layers recomputed so a
// counter just put on counts).
func msc13CreaturesWithPowerAtLeast(g *game.Game, controller uuid.UUID, n int) int {
	g.RecomputeLayersIfStaleLocked()
	count := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.CurrentPower() >= n {
			count++
		}
	}
	return count
}

// msc13TapTheLegalTargets taps each target of the item that is still legal
// ("tap up to one target creature"), nothing for none.
func msc13TapTheLegalTargets(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// msc13ThreatenAndPaint is "gain control of target creature until end of
// turn. Untap that creature. Until end of turn, it gains haste and becomes
// a <subtype> in addition to its other types" (Coercive Recruiter, Loki's
// Scepter): the steal, the untap, haste, and a layer-4 type add pinned to
// the creature's own entry (CR 400.7, CR 611.2c). `subtype` is the word
// read off the card's own text (ctx.TextSubtype).
func msc13ThreatenAndPaint(ctx *Context, target uuid.UUID, name, subtype string) error {
	if err := (GainControl{
		Target:   target,
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    name + " — gain control until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	if err := (GrantKeywordUntilEOT{
		Target:   target,
		Keywords: []string{"haste"},
		Label:    name + " — haste until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	return untilEndOfTurn(ctx, target, nil,
		name+" — becomes a "+subtype+" in addition to its other types",
		game.AddSubtypesMod(subtype))
}

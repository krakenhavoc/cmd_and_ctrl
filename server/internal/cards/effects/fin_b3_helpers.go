package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b3_helpers.go — shared bodies for Final Fantasy batches 3 and 4
// (#4462).

// finB3AttachAndGrantUntilEOT is "attach it to target creature you control.
// That creature gains <keyword> until end of turn" as an enters trigger's
// body (Coral Sword). The keyword goes to the still-legal target whether or
// not the attach happened: with the Equipment gone in response the attach
// does nothing (CR 701.3b) and "that creature" is still the target.
func finB3AttachAndGrantUntilEOT(name, keyword string) func(*game.Game, *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := AttachSourceToTarget(g, item); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{keyword},
			Label:    name + " — " + keyword + " until end of turn",
		}.Apply(NewContext(g, item))
	}
}

// finB3LoreCounterOnTheModesTarget is "put a lore counter on" (delta 1) or
// "remove a lore counter from" (delta -1) "target Saga you control" as a
// modal bullet's body: the occurrence's still-legal pick (CR 608.2b).
func finB3LoreCounterOnTheModesTarget(delta int) func(*game.StackItem, *Context, int) error {
	return func(_ *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok || t.Kind != game.TargetCard {
			return nil
		}
		return AddCounter{Target: t.ID, Kind: game.CounterLore, N: delta}.Apply(ctx)
	}
}

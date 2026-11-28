package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b710_helpers.go — shared helpers for OLD batch 710 (Portal Three
// Kingdoms, #5622).

// OldB710SelfWhileAnOpponentControlsA is a self-only static condition: "this
// creature gets … as long as an opponent controls a <pred> permanent" (Wu
// Admiral: an Island). It holds while any opponent still in the game
// controls a permanent that matches.
func OldB710SelfWhileAnOpponentControlsA(pred CardPredicate) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		if target.InstanceID != source.InstanceID {
			return false
		}
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == source.Controller {
				continue
			}
			if len(PermanentsControlledByMatching(g, p.ID, pred)) > 0 {
				return true
			}
		}
		return false
	}
}

// OldB710TargetGetsAndGainsIndefinitely is "target creature gets +P/+T and
// gains <keyword>" with no duration printed (Riding the Dilu Horse): one
// layer-6 and layer-7c effect that lasts indefinitely (CR 611.2a), pinned to
// the object it was applied to so it ends if that object leaves the
// battlefield (CR 400.7). The target must still be legal on resolution.
func OldB710TargetGetsAndGainsIndefinitely(ctx *Context, power, toughness int, keyword, label string) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	if !OnBattlefield(ctx.Game, t.ID) {
		return nil
	}
	id := t.ID
	return ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.ModifyPTMod(power, toughness), game.AddKeywordsMod(keyword)},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
		Label:    fmt.Sprintf("%s — %+d/%+d and %s", label, power, toughness, keyword),
	}.Apply(ctx)
}

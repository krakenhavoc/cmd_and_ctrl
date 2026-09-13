package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heliod's Intervention — Instant {X}{W}{W} (EDHREC rank 1499):
//
//	"Choose one —
//	 • Destroy X target artifacts and/or enchantments.
//	 • Target player gains twice X life."
//
// White's scalable artifact-and-enchantment sweep, with a lifegain
// mode for when there is nothing to sweep. The first mode's target
// count IS X — Crackle with Power's CountFromX, on a mode's clause
// rather than the card's, so the announce path pins the count to the
// X announced and re-checks each slot at resolution; a target that
// left in response is skipped and the rest are still destroyed (CR
// 608.2b). Each target is destroyed through the single-permanent
// verb rather than the batched sweep, so an indestructible target
// survives as printed (#446: the batched path does not check). The
// second mode reads X back at resolution and doubles it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7564d66-767c-4cd9-a5f0-0f2488a4a74b",
		Name:         "Heliod's Intervention",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Destroy X target artifacts and/or enchantments.", b13XTargetArtifactsOrEnchantments()),
			Mode("Target player gains twice X life.", TargetPlayer("target player")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
			case ctx.HasMode(1):
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				return GainLife{Player: item.Targets[0].ID, Amount: 2 * ctx.X()}.Apply(ctx)
			}
			return nil
		},
	})
}

// b13XTargetArtifactsOrEnchantments is "X target artifacts and/or
// enchantments" — the count bound to the announced X.
func b13XTargetArtifactsOrEnchantments() *game.TargetSpec {
	spec := TargetPermanent("X target artifacts and/or enchantments", Or(Artifact(), Enchantment()))
	spec.CountFromX = true
	return spec
}

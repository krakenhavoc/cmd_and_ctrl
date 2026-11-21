package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin War Strike — Sorcery {R} (Scourge):
//
//	"Goblin War Strike deals damage to target player or planeswalker equal
//	 to the number of Goblins you control."
//
// Volley Veteran's count (every Goblin permanent you control, read as the
// spell resolves; a changeling counts, CR 702.73a) dealt to Keldon
// Champion's player-or-planeswalker target. With no Goblin it deals no
// damage, and a target that left in response takes none (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5944d5c7-65ab-425f-a453-ed5871ca3d73",
		Name:         "Goblin War Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			goblin := ctx.TextSubtype("Goblin")
			n := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtype(goblin) })
			targets := ctx.LegalTargets()
			if n <= 0 || len(targets) == 0 {
				return nil
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: n}.Apply(ctx)
		},
	})
}

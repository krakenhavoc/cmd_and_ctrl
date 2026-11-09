package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akoum — Plane — Zendikar (#4870):
//
//	"Players may cast enchantment spells as though they had flash.
//	 Whenever chaos ensues, destroy target creature that isn't enchanted."
//
// The first ability is a cast-timing statement about every player's
// enchantment spells (CastKindAsThoughFlash with TimingAffectsEachPlayer); the
// cast gate reads it from the face-up planar card (game.forEachPlanarSourceLocked).
// The chaos target must not be enchanted when chosen, and is checked again as
// the ability resolves: a creature enchanted by then makes the target illegal
// and nothing happens (the ruling).
//
// No simplification.
func init() {
	flash := CastKindAsThoughFlash(game.PermissionFilter{CardType: "Enchantment"}, "Players may cast enchantment spells as though they had flash.")
	flash.Affects = game.TimingAffectsEachPlayer
	Register(Spec{
		OracleID:     "25650a32-4014-4065-ad01-7357c3ad3995",
		Name:         "Akoum",
		Completeness: CompletenessFull,
		CastTimings:  []game.CastTimingRule{flash},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Akoum — destroy target creature that isn't enchanted",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return DestroyTarget{Target: id}.Apply(ctx)
				}),
				TargetCreature("target creature that isn't enchanted", b15NotEnchanted())),
		},
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b828_helpers.go — shared helpers for OLD batch 828 (Ice Age Auras,
// #5636).

// OldB828OpponentControlsA reports whether an opponent of the source's
// controller still in the game controls a permanent that matches pred.
func OldB828OpponentControlsA(g *game.Game, source *game.Card, pred CardPredicate) bool {
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

// OldB828PumpAttachedWhileOpponentControls is "Enchanted creature gets
// +P/+T as long as an opponent controls a <pred> permanent" (Black Scarab,
// Red Scarab): a layer 7c modify checked as the layers are recomputed. The
// opponent is the Aura controller's.
func OldB828PumpAttachedWhileOpponentControls(pred CardPredicate, power, toughness int, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:    label,
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return AttachedToSource(target, g, source) && OldB828OpponentControlsA(g, source, pred)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// OldB828SnowLandsControlledBy counts the snow lands a player controls.
func OldB828SnowLandsControlledBy(g *game.Game, player uuid.UUID) int {
	return len(PermanentsControlledByMatching(g, player, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsLand() && c.HasSupertype("Snow")
	}))
}

// OldB828SnowblindStatic is Snowblind's "Enchanted creature gets -X/-Y. If
// that creature is attacking, X is the number of snow lands defending player
// controls. Otherwise, X is the number of snow lands its controller controls.
// Y is equal to X or to enchanted creature's toughness minus 1, whichever is
// smaller." One layer 7c modify; the toughness read is the one the earlier
// effects of the layer have left.
func OldB828SnowblindStatic() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Enchanted creature gets -X/-Y for the snow lands.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: AttachedToSource,
		// X reads whether the creature is attacking (and whom), so the
		// layers are recomputed when combat starts or ends.
		DependsOnAttackingStatus: true,
		Apply: func(c *game.Characteristic, host *game.Card, g *game.Game, _ *game.Card) {
			owner := host.Controller
			if host.AttackingTarget != uuid.Nil {
				if d := g.DefendingPlayerForAttackerForEffect(host.InstanceID); d != uuid.Nil {
					owner = d
				}
			}
			x := OldB828SnowLandsControlledBy(g, owner)
			y := min(x, c.Toughness-1)
			c.Power -= x
			c.Toughness -= y
		},
	}
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadly Rollick — Instant {3}{B}:
//
//	"If you control a commander, you may cast this spell without
//	paying its mana cost.
//	 Exile target creature."
//
// One of the five "free spells" from Commander Legends. The exile
// half is ordinary removal; the free-cast half is what makes the card
// a staple.
//
// The shared commander alternative-cost gate checks the caster controls
// a commander when the spell is announced. Either printed cost is legal.
func init() {
	Register(Spec{
		OracleID:         "0456ec64-2c81-4763-a352-8ff64a4c3d6b",
		Name:             "Deadly Rollick",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{FreeIfYouControlCommander("Cast without paying its mana cost")},
		Targets:          TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}

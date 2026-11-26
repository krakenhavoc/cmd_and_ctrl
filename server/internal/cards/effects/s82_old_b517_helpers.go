package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b517_helpers.go — shared shapes of Shards of Alara batch 517 (S82
// OLD, #5588).

// S82OldB517CantAttackUnlessAnotherArtifact is "This creature can't attack
// unless you control another artifact." (Steelclad Serpent): a layer 6
// static on the creature itself that sets CantAttack while its controller
// controls no OTHER artifact (it is an artifact itself and does not count).
// Types are read after the layers; restrictions are checked as attackers are
// declared (CR 508.1c).
func S82OldB517CantAttackUnlessAnotherArtifact() game.StaticAbility {
	return game.StaticAbility{
		Label: "This creature can't attack unless you control another artifact.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			if !SelfOnly(target, g, source) {
				return false
			}
			for _, id := range PermanentsControlledByMatching(g, source.Controller, Artifact()) {
				if id != source.InstanceID {
					return false
				}
			}
			return true
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= game.CantAttack
		},
	}
}

// S82OldB517CondYouControlPower5 is "if you control a creature with power 5
// or greater" (Sunseed Nurturer's intervening "if", CR 603.4): read through
// the layers as the end step begins and again as the ability resolves.
type S82OldB517CondYouControlPower5 struct{}

func (S82OldB517CondYouControlPower5) Holds(ctx *Context) bool {
	return B24ControlsCreatureWithPowerAtLeast(ctx.Game, ctx.Controller(), 5)
}

func (S82OldB517CondYouControlPower5) Condition() (game.FootprintStep, bool) {
	return CondStep(game.FootprintCondBoard, 0), true
}

// S82OldB517YouControlPower5 is the same test as a trigger's When.
func S82OldB517YouControlPower5(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return B24ControlsCreatureWithPowerAtLeast(g, source.Controller, 5)
}

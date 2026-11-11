package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2_b19_helpers.go — shared bodies for Commander wave 2, batches 19
// and 20 (#5368).

// cmdr2B19AttachedCantAttackYou is "Enchanted creature … can't attack you
// or planeswalkers you control" (Vow of Duty, Vow of Lightning, CR 508.1c):
// a layer 6 restriction on the attached creature naming the Aura's
// controller, read when the layers run, so a stolen Aura protects its new
// controller. A battle that player protects is not covered.
func cmdr2B19AttachedCantAttackYou() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Enchanted creature can't attack you or planeswalkers you control.",
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if source == nil {
				return
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
				Source:     source.InstanceID,
				SourceName: source.Name,
				NotPlayer:  source.Controller,
			})
		},
	}
}

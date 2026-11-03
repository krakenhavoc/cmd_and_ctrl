package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// AttachedCantHaveCounters is "Enchanted creature can't have counters put
// on it" (Blossombind, #4692): a layer-6 static of the Aura that sets
// Characteristic.CantHaveCounters on its host. Read live off the
// attachment, so it ends the moment the Aura leaves, and it is the Aura's
// effect, so the creature losing all abilities does not lift it. The
// engine does the rest (game/cant_have_counters.go): every placement,
// counter cost, blight choice and proliferate choice leaves the creature
// out (CR 101.2, 118.3).
func AttachedCantHaveCounters() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Enchanted creature can't have counters put on it.",
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.CantHaveCounters = true
		},
	}
}

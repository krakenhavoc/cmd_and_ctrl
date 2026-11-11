package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// soc_b5_helpers.go — shared bodies for the fifth Secrets of Strixhaven
// Commander batch (#4834).

// socB5EnchantedByYourAuraCantAttackYou is Eriette of the Charmed Apple's
// "Each creature that's enchanted by an Aura you control can't attack you
// or planeswalkers you control" (CR 508.1c): a layer 6 static that writes
// a NotPlayer restriction, naming the source's controller, onto each
// creature with an Aura that player controls attached. "You" and the
// Aura's controller are both read when the layer pass runs, so a stolen
// Eriette protects its new controller, and an Aura that changes hands or
// falls off changes which creatures are bound.
func socB5EnchantedByYourAuraCantAttackYou() game.StaticAbility {
	return game.StaticAbility{
		Label: "Each creature that's enchanted by an Aura you control can't attack you or planeswalkers you control.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			if source == nil || !target.IsCreature() || g == nil || g.Battlefield == nil {
				return false
			}
			for i := range g.Battlefield.Cards {
				a := &g.Battlefield.Cards[i]
				if a.Controller == source.Controller && a.IsAttachedTo(target.InstanceID) && a.IsAura() {
					return true
				}
			}
			return false
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			cmdr2B19BarAttackingSourcesController(c, source)
		},
	}
}

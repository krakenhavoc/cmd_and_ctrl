package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// HealOtherDamageWhenDealtDamage is the replacement "If damage would be
// dealt to <this creature>, instead that damage is dealt, but all other
// damage already dealt to it is healed" (CR 614.1a, CR 701.69a): Wolverine,
// Fierce Fighter (#4614).
//
// Put it on the creature's Spec.Replacements. The new damage is dealt in
// full; the damage marked before it is removed, so state-based actions
// judge lethal damage on the new damage alone. Damage dealt at the same
// time (the blockers of one combat damage step, CR 510.2, or one
// instruction's) is not "already dealt" and is kept. It is not a
// prevention effect, so "can't be prevented" does not touch it. The rule,
// and what it leaves alone (wither and infect counters, deathtouch), is
// game/heal_damage.go.
func HealOtherDamageWhenDealtDamage() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			// Nothing to heal: leave the CR 616 ordering prompt out of it.
			return src != nil && ev.Kind == game.RepEventDamage && ev.DamageTarget == src.InstanceID &&
				g.HealableDamageForEffect(ev) > 0
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
			g.HealOtherDamageForEffect(ev)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Heal other damage",
	}
}

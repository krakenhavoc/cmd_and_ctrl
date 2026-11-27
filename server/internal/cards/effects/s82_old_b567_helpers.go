package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b567_helpers.go — shared pieces of Coldsnap batch 567 (#5597, S82).

// S82OldB567SnowPermanentsAtLeast is "Activate only if you control four or more
// snow permanents" (Heidar, Rimewind Cryomancer): the activator's own count,
// read through the layers as the ability is announced (CR 602.5b).
func S82OldB567SnowPermanentsAtLeast(n int) ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		return len(PermanentsControlledByMatching(g, controller, Snow())) >= n
	}
}

// S82OldB567PreventDamageDealtBy is a standing CR 615 prevention of damage
// dealt by the permanent that has it: "prevent all damage that this creature
// would deal to <those>". `to` narrows the recipient, read as the damage would
// be dealt.
func S82OldB567PreventDamageDealtBy(label string, to func(recipient game.Card) bool) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource != src.InstanceID {
				return false
			}
			recipient, ok := g.LookupCardForEffect(ev.DamageTarget)
			return ok && g.Battlefield.Contains(ev.DamageTarget) && to(recipient)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      label,
	}
}

package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skoa, Embermage — Legendary Creature — Goblin Wizard {4}{R}{R}, 4/4
// (Modern Horizons 3):
//
//	"When Skoa enters, it deals 4 damage to any target.
//	 Grandeur — Discard another card named Skoa, Embermage, Sacrifice two
//	 Mountains: Skoa deals 4 damage to any target."
//
// The entry is a targeted trigger (Case of the Burning Masks' shape).
// Grandeur is an ordinary activated ability, Oriss, Samite Guardian's
// discard-by-name cost plus a sacrifice of two Mountains (effective land
// subtype, so a land that is a Mountain through another effect counts);
// "another" is any other copy in hand, since this one is on the
// battlefield.
//
// No simplification.
func init() {
	const name = "Skoa, Embermage"
	enter := effects.Targeting(effects.WhenThisEnters("Skoa, Embermage — it deals 4 damage to any target", effects.DamageToFirstTarget(4)), effects.TargetAny())
	enter.Purpose = effects.ForTargets(effects.DamageToTarget(0, 4))
	effects.Register(effects.Spec{
		OracleID:     "0e49ee63-a293-4b33-98cc-ebdcc2c64e82",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enter},
		Activated: []effects.ActivatedAbility{{
			Label: "Grandeur — Discard another card named Skoa, Embermage, Sacrifice two Mountains: Skoa deals 4 damage to any target.",
			Cost: effects.Plus(
				effects.DiscardCardsMatching(1, "another card named Skoa, Embermage", func(c game.Card) bool { return c.Name == name }),
				effects.SacrificeN(2, "two Mountains", effects.Subtype("Mountain"))),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
			Targets: effects.TargetAny(),
			Effect:  effects.DamageToFirstTarget(4),
		}},
	})
}

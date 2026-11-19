package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	enter := Targeting(WhenThisEnters("Skoa, Embermage — it deals 4 damage to any target", DamageToFirstTarget(4)), TargetAny())
	enter.Purpose = ForTargets(DamageToTarget(0, 4))
	Register(Spec{
		OracleID:     "0e49ee63-a293-4b33-98cc-ebdcc2c64e82",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enter},
		Activated: []ActivatedAbility{{
			Label: "Grandeur — Discard another card named Skoa, Embermage, Sacrifice two Mountains: Skoa deals 4 damage to any target.",
			Cost: Plus(
				DiscardCardsMatching(1, "another card named Skoa, Embermage", func(c game.Card) bool { return c.Name == name }),
				SacrificeN(2, "two Mountains", Subtype("Mountain"))),
			Purpose: ForTargets(DamageToTarget(0, 4)),
			Targets: TargetAny(),
			Effect:  DamageToFirstTarget(4),
		}},
	})
}

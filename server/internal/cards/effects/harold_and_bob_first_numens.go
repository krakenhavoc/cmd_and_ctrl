package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harold and Bob, First Numens — Legendary Creature — Treefolk Mutant
// {2}{G}, 3/3 (#2809):
//
//	"Reach, vigilance
//	 When Harold and Bob dies, if it was a creature, return it to the
//	 battlefield. It's an Aura enchantment with enchant Forest you
//	 control and 'Enchanted Forest has "{T}: Add three mana of any one
//	 color. You get two rad counters."' Harold and Bob loses all other
//	 abilities."
//
// The seam is "returns as an Aura" (returned_aura.go, ADR 0154's
// 2026-10-23 amendment), shared with any card that prints the same
// shape:
//
//   - The dies trigger is WhenThisDiesReturnItAsAnAura. "If it was a
//     creature" is read from the last-known information (CR 603.4,
//     CR 603.10a), so the Aura that dies in its turn stays in the
//     graveyard.
//   - The return is an Aura entering other than by resolving as an Aura
//     spell, so its controller chooses a Forest they control as it
//     enters (CR 303.4f), and with none the card stays in the graveyard
//     (CR 303.4g). The clause is Spec.ReturnedAura. It is held to it
//     afterwards (CR 704.5m): a Forest that changes hands or stops being
//     a Forest takes the Aura with it.
//   - ReturnedAuraForm gives the permanent its Aura form: an Aura
//     enchantment and no creature (CR 205.1a), still legendary
//     (CR 205.4b), with every ability but the Aura's own gone (CR 613.1f).
//   - The grant is the ordinary layer-6 bundle of ADR 0093, to the
//     enchanted Forest: a mana ability whose rider gives its controller
//     two rad counters. The rider runs inside the mana ability's
//     resolution (CR 605.3b) and it is the Forest's controller who gets
//     them, as "you" in a granted ability is. Rad counters are given
//     through the CR 614 player-counter window (playerGetsRadCounters).
//
// "Three mana of any one color" is one pick minting three of that
// colour, not three picks (OneColorOfAmount).
//
// No simplification.
func init() {
	const grant = "harold-and-bob-first-numens/three-mana-and-rad"
	Register(Spec{
		OracleID:        "d0ecea04-9768-4b09-8718-dd59135251e3",
		Name:            "Harold and Bob, First Numens",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		ReturnedAura:    EnchantLand(Subtype("Forest"), YouControl()),
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add three mana of any one color. You get two rad counters.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: OneColorOfAmount(3),
				Label:    "Add three mana of any one color. You get two rad counters",
				Rider: func(g *game.Game, controller, _ uuid.UUID) error {
					return playerGetsRadCounters(g, controller, controller, 2)
				},
			}},
		}},
		Static: ReturnedAuraForm(grant),
		Triggered: []game.TriggeredAbility{
			WhenThisDiesReturnItAsAnAura("Harold and Bob"),
		},
	})
}

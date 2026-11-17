package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slobad, Iron Goblin — Legendary Creature — Phyrexian Goblin Artificer
// {2}{R}, 3/3 (ONE):
//
//	"{T}, Sacrifice an artifact: Add an amount of {R} equal to the
//	 sacrificed artifact's mana value. Spend this mana only to cast
//	 artifact spells or activate abilities of artifacts."
//
// A mana ability (CR 605.1a) whose cost taps it and sacrifices an
// artifact. The amount is read from the payment, the artifact as it last
// existed on the battlefield (CR 608.2h, ProducedForPaid over
// PaidCost.SacrificedObjects, as Food Chain reads its exiled creature).
// A mana value of 0 adds nothing. The restriction is Cargo Ship's: an
// artifact spell, or an ability whose source is an artifact (CR 106.6).
// Summoning sickness applies to the {T} (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72062fb6-0329-424d-936a-882b638d6d57",
		Name:         "Slobad, Iron Goblin",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:            true,
				SacrificeOther: SacrificeN(1, "an artifact", Artifact()).SacrificeOther,
			},
			ProducedForPaid: slobadProduced,
			Label:           "{T}, Sacrifice an artifact: Add an amount of {R} equal to the sacrificed artifact's mana value. Spend this mana only to cast artifact spells or activate abilities of artifacts.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
			)},
		}},
	})
}

// slobadProduced is "an amount of {R} equal to the sacrificed artifact's
// mana value". With nothing sacrificed yet (CR 106.7's "could produce"
// readers ask before any payment) it answers one {R}: the colour is what
// matters there, and the amount is the payment's business.
func slobadProduced(g *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	if len(paid.SacrificedObjects) == 0 {
		return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"R"}, N: 1}}
	}
	mv := 0
	if info, ok := g.PermanentForEffect(paid.SacrificedObjects[0]); ok {
		mv = info.ManaValue
	}
	return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"R"}, N: mv}}
}

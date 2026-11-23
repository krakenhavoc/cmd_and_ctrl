package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tsp_b2_helpers.go — shapes shared by the Time Spiral batch 2 cards
// (#5596, S82). Each is a package-level func or a constructor whose
// closure captures only constants, so nothing here holds a game.

// tspB2StorageLand is the whole Spec of a Time Spiral storage land:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Put a storage counter on this land.
//	 {1}, Remove X storage counters from this land: Add X mana in any
//	 combination of {A} and/or {B}."
//
// Mage-Ring Network's shape with a {1} in the payout's cost and no {T}:
// the counter is a plain activated ability that uses the stack (CR
// 605.1a), and the payout is a mana ability whose cost pays {1} and
// removes the X counters the activator announces (X may be 0). Each of
// the X mana is either of the land's two colours, picked one by one.
// `a` and `b` are colour letters ("W", "U").
func TspB2StorageLand(oracleID, name, a, b string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost: ManaAbilityCost{
					Mana:           "{1}",
					RemoveCounters: RemoveCountersXFromThis(game.CounterStorage, 0).RemoveCounters,
				},
				ProducedForPaid: ProducedPerCounterRemovedInAnyCombinationOf(a, b),
				Label:           "{1}, Remove X storage counters: Add X mana in any combination of {" + a + "} and/or {" + b + "}",
			},
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Put a storage counter on this land.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect:  PutStorageCounterOnThis,
		}},
	}
}

// tspB2OtherWhiteCreatures scopes Celestial Crusader's anthem: every
// creature on the battlefield that is white as it is now, whoever controls
// it, except the Crusader itself. The colour is a word of the card's text
// (CR 612.2), read through TextColor.
func TspB2OtherWhiteCreatures(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.InstanceID != source.InstanceID &&
		slices.Contains(target.EffectiveColors(), source.TextColor("W"))
}

// tspB2DestroyArtifactsCreaturesEnchantments is Magus of the Disk's
// "Destroy all artifacts, creatures, and enchantments.", the Magus
// included if it is still there.
func TspB2DestroyArtifactsCreaturesEnchantments(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	match := Or(Artifact(), Creature(), Enchantment())
	return DestroyAllMatching{Match: match}.Apply(ctx)
}

// tspB2MainPhaseAmount is n, or `main` instead if the spell's controller
// cast it during their main phase (Might of Old Krosa, Haunting Hymn,
// Sulfurous Blast, Careful Consideration). It is read as the spell
// resolves, which is the same answer: the stack has to empty for a step
// to end, so a spell resolves in the step it was cast in.
func TspB2MainPhaseAmount(ctx *Context, n, main int) int {
	if B14InYourMainPhase(ctx.Game, ctx.Controller()) {
		return main
	}
	return n
}

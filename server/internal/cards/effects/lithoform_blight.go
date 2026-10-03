package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lithoform Blight — Enchantment — Aura {1}{B}:
//
//	"Enchant land
//	 When this Aura enters, draw a card.
//	 Enchanted land loses all land types and abilities and has
//	 '{T}: Add {C}' and '{T}, Pay 1 life: Add one mana of any color.'"
//
// ADR 0109 §2 (#1604): the static form of "loses all land types"
// (LosesAllLandTypes, layer 4) with "loses all abilities and has …" in
// layer 6 (LosesAllAbilitiesAndHas). The enchanted land keeps its card
// types, supertypes and any other subtype (CR 205.1a: a Dryad Arbor is
// still a Dryad), has no intrinsic mana ability left (CR 305.6), and taps
// for the two granted abilities only. The 1 life is a cost, paid as the
// land taps (CR 118.3, ManaAbilityCost.Life), so it can't be activated at
// 0 life; like Mana Confluence, the auto-tapper never spends life on its
// own.
//
// No simplification.
const (
	lithoformBlightColorless = "lithoform-blight/colorless"
	lithoformBlightAnyColor  = "lithoform-blight/any-color-for-life"
)

func init() {
	Register(Spec{
		OracleID:     "b256c311-e37d-4b3c-889c-85443e3d5e7e",
		Name:         "Lithoform Blight",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Grants: []AbilityGrant{
			TapForManaGrant(lithoformBlightColorless, "{C}", "Add {C}", "{T}: Add {C}."),
			{
				Key: lithoformBlightAnyColor,
				Mana: []ManaAbility{{
					Cost:     ManaAbilityCost{Tap: true, Life: 1},
					Produced: "{W|U|B|R|G}",
					Label:    "Pay 1 life: Add one mana of any color",
				}},
				Text: "{T}, Pay 1 life: Add one mana of any color.",
			},
		},
		Static: []game.StaticAbility{
			LosesAllLandTypes(AttachedToSource),
			LosesAllAbilitiesAndHas(AttachedToSource, lithoformBlightColorless, lithoformBlightAnyColor),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lithoform Blight — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

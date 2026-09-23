package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weaver of Harmony — Enchantment Creature — Snake Druid {1}{G}, 2/2
// (EDHREC rank 2090):
//
//	"Other enchantment creatures you control get +1/+1.
//	 {G}, {T}: Copy target activated or triggered ability you control
//	 from an enchantment source. You may choose new targets for the
//	 copy. (Mana abilities can't be targeted.)"
//
// The enchantress deck's lord. The anthem is a layer 7c modify over
// other enchantment creatures the controller controls — post-layer
// types, so a creature made an enchantment by another effect counts.
//
// The copy ability shipped with #1223, which built both halves it was
// waiting on: a target clause that can name an ability ITEM
// (AbilityOnStack, game.TargetSpec.Abilities) and
// game.CopyAbilityForEffect. "From an enchantment source" is
// AbilityFromSource(Enchantment()), read off the source card as it is
// now — a permanent that has stopped being an enchantment has stopped
// being a legal source. The reminder text's "mana abilities can't be
// targeted" needs no clause of its own: a mana ability never uses the
// stack (CR 605.3b), so there is no item to offer.
func init() {
	Register(Spec{
		OracleID:     "494b31b2-27ef-4ca1-ac72-c2fcfc8a23a1",
		Name:         "Weaver of Harmony",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{G}, {T}: Copy target activated or triggered ability you control from an enchantment source",
			Cost:  Plus(ManaCost("{G}"), TapCost()),
			Targets: AbilityOnStack("target activated or triggered ability you control from an enchantment source",
				AnAbilityYouControl(), AbilityFromSource(Enchantment())),
			Effect: copyTargetedAbility,
		}},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
					target.IsCreature() && target.IsEnchantment()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
	})
}

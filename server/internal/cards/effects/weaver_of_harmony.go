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
// Sandbox simplification, declared (Multani's posture, one whole
// ability omitted): the copy ability is not implemented. Copying
// an ABILITY on the stack has no shape — CopySpellForEffect resolves
// its target through the stack's spell cards, and a triggered or
// activated ability is a StackMeta item with no card, so neither the
// target clause ("target activated or triggered ability") nor the
// copy itself exists yet. Weaker than printed, never stronger: the
// Weaver is a 2/2 lord that taps for nothing.
func init() {
	Register(Spec{
		OracleID:     "494b31b2-27ef-4ca1-ac72-c2fcfc8a23a1",
		Name:         "Weaver of Harmony",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The ability-copying activation isn't implemented — the Weaver only gives your other enchantment creatures +1/+1."},
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

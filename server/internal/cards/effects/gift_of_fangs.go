package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gift of Fangs — Enchantment — Aura {B} (Innistrad: Crimson Vow, #5495):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 as long as it's a Vampire. Otherwise,
//	 it gets -2/-2."
//
// Two halves of one condition, both a layer 7c modification (CR 613.4c),
// read live on every pass (Burden of Proof's shape): the enchanted creature
// is a Vampire, so +2/+2, or it is not, so -2/-2. "Vampire" is a word of
// the Aura's text, read through it (CR 612.2), and through the layers, so
// a changeling gets the bonus and a creature that stops being a Vampire
// swings to the penalty at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "901f0a8f-7462-4970-8d21-265c8c0188e6",
		Name:         "Gift of Fangs",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			{
				Label:    "Enchanted creature gets +2/+2 as long as it's a Vampire.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source.IsAttachedTo(target.InstanceID) && target.HasSubtype(source.TextSubtype("Vampire"))
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			{
				Label:    "Otherwise, it gets -2/-2.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source.IsAttachedTo(target.InstanceID) && !target.HasSubtype(source.TextSubtype("Vampire"))
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power -= 2
					c.Toughness -= 2
				},
			},
		},
	})
}

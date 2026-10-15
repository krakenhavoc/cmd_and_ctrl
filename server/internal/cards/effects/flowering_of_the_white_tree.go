package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flowering of the White Tree — Legendary Enchantment {W}{W} (EDHREC
// top 1,000, #3654):
//
//	"Legendary creatures you control get +2/+1 and have ward {1}.
//	 Nonlegendary creatures you control get +1/+1."
//
// Two layer 7c statics split on the legendary supertype, plus the
// granted ward as Chapel Shieldgeist grants it (WardGranted): the ward
// trigger reads the warded creature's supertypes when it is targeted.
// The two pumps never overlap, so the +2/+1 and +1/+1 are independent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3cda958a-f777-433d-850e-a6c8f581db4e",
		Name:         "Flowering of the White Tree",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				Label:    "Legendary creatures you control get +2/+1.",
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller && target.IsLegendary()
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness++
				},
			},
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				Label:    "Nonlegendary creatures you control get +1/+1.",
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller && !target.IsLegendary()
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{1}"), "Flowering of the White Tree — legendary creatures you control have ward {1}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller && target.IsLegendary()
				}),
		},
	})
}

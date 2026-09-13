package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elesh Norn, Grand Cenobite — Legendary Creature — Phyrexian Praetor
// {5}{W}{W}, 4/7 (EDHREC rank 859):
//
//	"Vigilance
//	 Other creatures you control get +2/+2.
//	 Creatures your opponents control get -2/-2."
//
// The one-sided Wrath that stays. Two Layer 7c statics — Glorious
// Anthem's shape with a 2, and the same again with the sign flipped
// for the other side of the table. The layer engine feeds the
// state-based actions, so an opponent's 2/2 dies to the shrink at
// the next check, and a creature they cast later dies as it lands,
// as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "958d71ff-c9f7-46f0-96ca-79e7f4d65a16",
		Name:            "Elesh Norn, Grand Cenobite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller &&
						target.InstanceID != source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller != source.Controller
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power -= 2
					c.Toughness -= 2
				},
			},
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Intangible Virtue — Enchantment {1}{W} (EDHREC rank 662):
//
//	"Creature tokens you control get +1/+1 and have vigilance."
//
// The token deck's anthem. Two static abilities in two layers, the
// way Glorious Anthem and Lord of Atlantis split the same sentence:
// the +1/+1 is layer 7c and the vigilance grant layer 6. Both apply
// to creature TOKENS the controller controls — IsToken reads the
// "Token" on the type line the templates stamp — and re-evaluate on
// every recompute, so a token made after the Virtue is pumped the
// moment it enters (the layer listener treats a token's creation as
// an entry since batch 01).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d21c3c8f-d105-4ba9-bf69-e5f26f0f8ec5",
		Name:         "Intangible Virtue",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: b05CreatureTokenYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
			{
				Layer:     game.Layer6Ability,
				AppliesTo: b05CreatureTokenYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !eotHasAbility(c.Abilities, "vigilance") {
						c.Abilities = append(c.Abilities, "vigilance")
					}
				},
			},
		},
	})
}

// b05CreatureTokenYouControl is Intangible Virtue's applies-to: a
// creature token under the source's controller.
func b05CreatureTokenYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && IsToken(*target) && target.Controller == source.Controller
}

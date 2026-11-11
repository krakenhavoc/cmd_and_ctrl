package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2_b17_helpers.go — card-side helpers of cmdr2 batch 17
// (Squirrels, Rats and Saprolings, #5368).

// cmdr2B17SquirrelMobPT is "This creature gets +1/+1 for each other
// Squirrel on the battlefield" (Squirrel Mob).
func cmdr2B17SquirrelMobPT() game.StaticAbility {
	return game.StaticAbility{
		Label:    "Gets +1/+1 for each other Squirrel on the battlefield.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			for _, o := range g.Battlefield.Cards {
				if o.InstanceID != source.InstanceID && o.IsCreature() && o.HasSubtype("Squirrel") {
					n++
				}
			}
			c.Power += n
			c.Toughness += n
		},
	}
}

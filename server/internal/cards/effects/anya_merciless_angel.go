package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anya, Merciless Angel — Legendary Creature — Angel {3}{R}{W}, 4/4:
// "Flying. Anya gets +3/+3 for each opponent whose life total is less than
// half their starting life total. As long as an opponent's life total is
// less than half their starting life total, Anya has indestructible."
// Both statics read life totals, so both set DependsOnLifeTotal (the layer
// cache would otherwise answer one life change late).
func init() {
	self := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID
	}
	pump := game.StaticAbility{
		Label:              "Anya gets +3/+3 for each opponent whose life total is less than half their starting life total.",
		Layer:              game.Layer7PT,
		SubLayer:           game.SubLayer7C_Modify,
		DependsOnLifeTotal: true,
		AppliesTo:          self,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := cmdr2S52OpponentsBelowHalfStarting(g, source)
			c.Power += 3 * n
			c.Toughness += 3 * n
		},
	}
	indestructible := LifeGatedKeyword(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && cmdr2S52OpponentsBelowHalfStarting(g, source) > 0
	}, "indestructible")
	indestructible.Label = "As long as an opponent's life total is less than half their starting life total, Anya has indestructible."
	Register(Spec{
		OracleID:        "0665daf5-e12a-4b59-81fc-63fec2807571",
		Name:            "Anya, Merciless Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{pump, indestructible},
	})
}

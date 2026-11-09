package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// planes_b9_helpers.go holds the shapes the ninth and tenth batches of plane
// cards (ADR 0216 slice 4, #4870) share. Append-only.

// b9AllCreaturesHave is "All creatures have <keyword>.": a layer-6 grant to
// every creature on the battlefield, whoever controls it (Concordant
// Crossroads' shape), read from the command zone while the plane is face up.
func b9AllCreaturesHave(keyword string) game.StaticAbility {
	return game.StaticAbility{
		Label: "All creatures have " + keyword + ".",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsCreature()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Abilities = game.AppendKeywordAbility(c.Abilities, keyword)
		},
	}
}

// b9AllCreaturesGetPT is "<match> creatures get +N/+M" as a layer-7c static.
func b9AllCreaturesGetPT(label string, applies func(target *game.Card, g *game.Game, source *game.Card) bool, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool { return target.IsCreature() && applies(target, g, source) },
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// b9DestroyEachNonlandPermanent destroys the nonland permanents `keep`
// does not spare, simultaneously.
func b9DestroyEachNonlandPermanent(ctx *Context, spared CardPredicate) error {
	match := Nonland()
	if spared != nil {
		match = And(Nonland(), Not(spared))
	}
	return DestroyAllMatching{Match: match}.Apply(ctx)
}

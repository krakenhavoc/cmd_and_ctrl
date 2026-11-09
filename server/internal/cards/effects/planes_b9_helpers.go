package effects

import (
	"slices"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
		Label:    label,
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && applies(target, g, source)
		},
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

// b9PerOpponentGraveyardClauses is "for each opponent, <verb> up to one
// target <noun> card from that player's graveyard": one "up to one" clause
// per possible opponent, the n-th bound to the n-th seat after the planar
// controller's. A seat that is not at the table, or is eliminated, offers
// nothing (opponent_clauses.go's reasoning).
func b9PerOpponentGraveyardClauses(noun string, what ...CardPredicate) *game.TargetSpec {
	var out []*game.TargetSpec
	for n := 1; n <= game.MaxPlayers-1; n++ {
		out = append(out, TargetCardInGraveyard(
			"up to one target "+noun+" card from the "+nthOpponentWords[n]+" opponent after you's graveyard",
			append([]CardPredicate{trekB36OwnedByNthOpponent(n)}, what...)...).WithCount(0, 1))
	}
	return Clauses(out[0], out[1:]...)
}

// b9IsBlack reports whether a card on the battlefield is black.
func b9IsBlack(c game.Card) bool {
	return slices.ContainsFunc(c.Effective().Colors, func(col string) bool { return strings.EqualFold(col, "B") })
}

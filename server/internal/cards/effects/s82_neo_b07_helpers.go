package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b07_helpers.go — S82 NEO batch 7 (#5493): the channel cards.
// Append-only.

// neoB07TargetGetsUntilEOT is "Target creature gets +P/+T and gains <keywords> until
// end of turn." (Ironhoof Boar's and Favor of Jukai's channel).
func neoB07TargetGetsUntilEOT(label string, power, toughness int, keywords ...string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		if power != 0 || toughness != 0 {
			if err := (BoostUntilEOT{Target: target, Power: power, Toughness: toughness, Label: label + " — " + fmt.Sprintf("%+d/%+d", power, toughness)}).Apply(ctx); err != nil {
				return err
			}
		}
		if len(keywords) == 0 {
			return nil
		}
		return GrantKeywordUntilEOT{Target: target, Keywords: keywords, Label: label + " — " + keywords[0]}.Apply(ctx)
	}
}

// neoB07DrawsCards is "Draw N cards." as an ability's whole effect.
func neoB07DrawsCards(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return DrawCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
	}
}

// neoB07AnimateLand is the tail of Roaring Earth's channel: once the counters
// are on, the land becomes a 0/0 green Spirit creature with haste, still a
// land, for good (CR 611.2a) and only if it is still the object that was
// targeted (CR 400.7). The land's entry stamp rides the tail as N.
var neoB07AnimateLand = CardTail("then/neo-roaring-earth-animate-land", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	land := o.Tail.Carry[0]
	c, ok := ctx.Game.LookupCardForEffect(land)
	if !ok || int(c.EnteredBattlefieldAt) != o.Tail.N || !onBattlefield(ctx.Game, land) {
		return nil
	}
	mods := []game.Mod{
		game.AddTypesMod("Creature"),
		game.AddSubtypesMod(ctx.TextSubtype("Spirit")),
		game.SetColorsMod(ctx.TextColors("G")...),
		game.AddKeywordsMod("haste"),
	}
	mods = append(mods, game.SetBasePTMods(0, 0)...)
	ctx.Game.RegisterScopedEffectForEffect(ctx.Source(),
		[]game.AffectedObject{game.PinObject(land, c.EnteredBattlefieldAt)}, mods,
		ctx.Game.PinnedTo(game.IndefiniteDuration(), land),
		"Roaring Earth — becomes a 0/0 green Spirit creature with haste that's still a land")
	return nil
})

// neoB07BounceTheTarget is "Return target creature to its owner's hand." as
// an activated ability's whole effect.
func neoB07BounceTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return BounceToHand{Target: id}.Apply(ctx)
}

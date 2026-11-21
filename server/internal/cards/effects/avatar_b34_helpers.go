package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b34_helpers.go — shared pieces of Avatar: The Last Airbender
// Eternal batches 34 to 37 (#5075). Append-only.

// counterTheTargetThenLootOne is "Counter target <spell>. Draw a card,
// then discard a card.": the announced spell is countered, then its
// controller loots. Refute and Deny Entry; each card's own clause says
// which spells are legal targets.
func CounterTheTargetThenLootOne(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
		return err
	}
	return LootOne(ctx.Game, item, 1)
}

// avB34ThenDraw is "draw N cards" after a keyed step settled (Water
// Whip's bounce): N rides the tail. An on-disk key: never renamed, never
// reused.
var AvB34ThenDraw = CardTail("avatar-b34/then-draw-n", func(ctx *Context, o game.TailOutcome) error {
	return DrawCards{N: o.Tail.N}.Apply(ctx)
})

// avB34ThenSpiritToken is "create a 1/1 colorless Spirit creature token
// with 'This token can't block or be blocked by non-Spirit creatures.'"
// after a keyed step settled (Lost in the Spirit World's bounce). An
// on-disk key: never renamed, never reused.
var AvB34ThenSpiritToken = CardTail("avatar-b34/then-spirit-token", func(ctx *Context, _ game.TailOutcome) error {
	return CreateToken{Template: KurukSpiritToken(), N: 1}.Apply(ctx)
})

// avB34CondAnOpponentLostLife is "if an opponent lost life this turn"
// (Lion Vulture), off the engine's per-turn tally of every opponent of the
// item's controller. It declares no footprint read: no condition kind
// names opponents' life loss.
type AvB34CondAnOpponentLostLife struct{}

func (AvB34CondAnOpponentLostLife) Holds(ctx *Context) bool {
	return OpponentLifeLostThisTurn(ctx.Game, ctx.Controller()) > 0
}

func (AvB34CondAnOpponentLostLife) Condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// avB34YouAttackedWithTwoLegendaryCreatures is "whenever you attack with
// two or more legendary creatures" (Master's Guidance): an attack the
// source's controller declared, with at least two legendary creatures
// they control attacking, read through the layered view. Paired with
// OncePerBatch, so a declaration is one trigger.
func AvB34YouAttackedWithTwoLegendaryCreatures(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !AttackDeclaredByYou(ev, source.Controller) {
		return false
	}
	n := 0
	for _, id := range B13AttackingCreaturesYouControl(g, source.Controller) {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsLegendary() {
			n++
		}
	}
	return n >= 2
}

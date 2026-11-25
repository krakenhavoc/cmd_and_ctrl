package cardsco

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Culmination of Studies — Sorcery {X}{U}{R} (Strixhaven: School of Mages,
// #5502):
//
//	"Exile the top X cards of your library. For each land card exiled this way,
//	 create a Treasure token. For each blue card exiled this way, draw a card.
//	 For each red card exiled this way, Culmination of Studies deals 1 damage to
//	 each opponent."
//
// A plain exile with a continuation (Riverwheel Sweep's ExileTopKeyedForEffect),
// told the cards that really reached exile, so a card a replacement sent
// elsewhere counts for nothing. The three counts are taken once from those
// cards, a blue-and-red card counting for both and a land (colorless) for
// neither colour, and are paid in printed order: Treasures, cards, damage. X of
// 0 exiles nothing and does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4e124af-d335-4c42-b60a-fe578f69f7ed",
		Name:         "Culmination of Studies",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.X() <= 0 {
				return nil
			}
			return ctx.Game.ExileTopKeyedForEffect(item.Controller, item.Controller, ctx.X(), ctx.ItemTail(culminationOfStudiesExiled))
		},
	})
}

// culminationOfStudiesExiled is the rest of the spell once the cards have
// landed in exile. An on-disk key: never renamed, never reused.
var culminationOfStudiesExiled = effects.CardTail("s82-old-b046/culmination-of-studies-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	lands, blue, red := 0, 0, 0
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if c.IsLand() {
			lands++
		}
		colors := c.EffectiveColors()
		if slices.Contains(colors, "U") {
			blue++
		}
		if slices.Contains(colors, "R") {
			red++
		}
	}
	if lands > 0 {
		if err := effects.TreasureTokens(ctx.Game, ctx.Controller(), lands); err != nil {
			return err
		}
	}
	if blue > 0 {
		if err := (effects.DrawCards{N: blue}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.DamageEachOpponent{N: red}.Apply(ctx)
})

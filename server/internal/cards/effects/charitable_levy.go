package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const charitableLevyCollection = "collection"

// Charitable Levy — Enchantment {1}{W} (Modern Horizons 3):
//
//	"Noncreature spells cost {1} more to cast.
//	 Whenever a player casts a noncreature spell, put a collection counter
//	 on this enchantment. Then if there are three or more collection
//	 counters on it, sacrifice it. If you do, draw a card, then you may
//	 search your library for a Plains card, put it onto the battlefield
//	 tapped, then shuffle."
//
// The tax is Thorn of Amethyst's, symmetrical. The trigger watches every
// player's casts of noncreature spells (it resolves above the spell, so a
// countered spell still counts). "If you do" is the sacrifice's keyed tail
// (SacrificeThisThen), so the draw and the search happen for the Levy's
// controller only if it really left; the search is Flagstones of Trokair's
// (a Plains card, tapped, then shuffle) and may find nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a628875-96c8-4a28-aaff-99fd99ae778e",
		Name:         "Charitable Levy",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsMore(1, "Noncreature spells cost {1} more to cast.", NoncreatureSpell()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventCast {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !c.IsCreature()
			}, "Charitable Levy — put a collection counter on this enchantment; with three or more, sacrifice it",
				charitableLevyCollect),
		},
	})
}

// charitableLevySacrificed is the "if you do" of the sacrifice: draw a
// card, then the Plains search. An on-disk key.
var charitableLevySacrificed = SacrificeThisThen("then/gap-b56-charitable-levy", func(ctx *Context) error {
	if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	return b31FetchPlainsTapped(ctx.Game, ctx.Item)
})

func charitableLevyCollect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: item.SourceCardID, Kind: charitableLevyCollection, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if c, ok := g.LookupCardForEffect(item.SourceCardID); !ok || c.Counters[charitableLevyCollection] < 3 {
		return nil
	}
	return charitableLevySacrificed(g, item)
}

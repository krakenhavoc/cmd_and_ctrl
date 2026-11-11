package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Communal Brewing — Enchantment {2}{G} (#5368):
//
//	"When this enchantment enters, any number of target opponents each draw
//	 a card. Put an ingredient counter on this enchantment, then put an
//	 ingredient counter on it for each card drawn this way.
//	 Whenever you cast a creature spell, that creature enters with X
//	 additional +1/+1 counters on it, where X is the number of ingredient
//	 counters on this enchantment."
//
// The enters trigger targets any number of opponents (none is allowed): each
// still-legal one draws, in turn order (CR 608.2b), then the enchantment gets
// one counter plus one per card actually drawn, counted off the batch like
// Struggle for Project Purity's (an opponent with an empty library draws
// nothing). The second ability is Aetherflux Conduit's shape (the cast
// trigger marks the spell so the counters ride its entry through every
// replacement), with X read off the enchantment as the trigger resolves, or
// its last-known counters if it has left (CR 608.2h). The trigger goes on the
// stack even at X = 0 and is dropped quietly.
//
// No simplification.
func init() {
	const draw = "Communal Brewing — any number of target opponents each draw a card; put an ingredient counter on this enchantment for it and for each card drawn"
	const cast = "Communal Brewing — that creature enters with X additional +1/+1 counters, where X is the number of ingredient counters on this enchantment"
	enters := WhenThisEnters(draw, Do(
		Instruction(cmdr2B23EachTargetPlayerDraws),
		Instruction(func(ctx *Context) error {
			n := 1 + opponentsDrawsThisResolution(ctx)
			return CounterOnThis{Kind: cmdr2B23IngredientCounter, N: n}.Apply(ctx)
		}),
	))
	enters.Targets = TargetPlayer("any number of target opponents", Opponent()).WithCount(0, 0)
	spell := On(game.EventCast, YouCast(Creature()), cast, func(g *game.Game, item *game.StackItem) error {
		n := 0
		if onBattlefield(g, item.SourceCardID) {
			if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
				n = c.Counters[cmdr2B23IngredientCounter]
			}
		} else {
			n = g.LastKnownCountersForEffect(item.SourceCardID)[cmdr2B23IngredientCounter]
		}
		if n <= 0 {
			return nil
		}
		g.MarkSpellEntersWithCountersForEffect(item.Params.Object, game.PromisedCounter{
			Kind:       game.CounterPlusOne,
			Count:      n,
			Source:     item.SourceCardID,
			SourceName: "Communal Brewing",
		})
		return nil
	})
	spell.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, cast)
		item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: g.ObjectEpochForEffect(ev.CardID)}
		return item
	}
	Register(Spec{
		OracleID:     "99deac41-584b-43f2-804e-c56ae0fef047",
		Name:         "Communal Brewing",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters, spell},
	})
}

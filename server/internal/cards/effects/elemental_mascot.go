package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elemental Mascot — Creature — Elemental Bird {1}{U}{R}, 1/4 (Secrets of
// Strixhaven, #4456):
//
//	"Flying, vigilance
//	 Opus — Whenever you cast an instant or sorcery spell, this creature
//	 gets +1/+0 until end of turn. If five or more mana was spent to cast
//	 that spell, exile the top card of your library. You may play that
//	 card until the end of your next turn."
//
// A cast trigger (a copy is not cast, CR 707.10). The mana spent is read off
// the payment record as the spell is cast and carried on the item (the
// ruling: the ability resolves before the spell, even if the spell is
// countered), Molten-Core Maestro's way, so the card declares
// ReadsManaSpent. The pump is unconditional; the impulse exile is the
// "five or more" half. The exiled card may be played (a land too, using a
// land play) until the end of its controller's next turn, a seat-turn
// window (ADR 0063), paying its costs as usual.
//
// No simplification.
func init() {
	const label = "Elemental Mascot — +1/+0 until end of turn; if five or more mana was spent, exile the top card of your library and you may play it until the end of your next turn"
	t := On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return instantOrSorceryCastByYou(ev, source, g)
	}, label, Do(
		Instruction(func(ctx *Context) error {
			return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Elemental Mascot — +1/+0 until end of turn"}.Apply(ctx)
		}),
		Instruction(func(ctx *Context) error {
			if ctx.Item.Params.Amount < 5 {
				return nil
			}
			window := ctx.Game.UntilEndOfYourNextTurnDuration(ctx.Controller())
			return ExileTopWithPermission{
				From:    ctx.Controller(),
				GrantTo: ctx.Controller(),
				N:       1,
				Window:  &window,
			}.Apply(ctx)
		}),
	))
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:        "882f30e4-4ce1-4581-8d2d-d763aa13befe",
		Name:            "Elemental Mascot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered:       []game.TriggeredAbility{t},
	})
}

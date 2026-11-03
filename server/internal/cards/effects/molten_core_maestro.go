package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molten-Core Maestro — Creature — Goblin Bard {1}{R}, 2/2 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Menace
//	 Opus — Whenever you cast an instant or sorcery spell, put a +1/+1
//	 counter on this creature. If five or more mana was spent to cast
//	 that spell, add an amount of {R} equal to this creature's power."
//
// A cast trigger (a copy is not cast, CR 707.10). The mana spent is read
// off the payment record as the spell is cast and carried on the item
// (the ruling: it resolves before the spell, even if the spell is
// countered), the same way Aetherflux Conduit reads it, so the card
// declares ReadsManaSpent. A waived or free cast spent none.
//
// The counter goes on first, and the mana is added as the ability
// resolves, so {R} equal to the power AFTER the counter: the power is read
// as the Maestro is then (counters and anthems included) or, if it has
// left, as it last existed on the battlefield (CR 608.2h). It is not a
// mana ability (the ruling), so it uses the stack and the mana empties
// with the step (CR 106.4).
//
// No simplification.
func init() {
	const label = "Molten-Core Maestro — put a +1/+1 counter on this creature; if five or more mana was spent, add {R} equal to its power"
	t := On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return instantOrSorceryCastByYou(ev, source, g)
	}, label, Do(
		CounterOnThis{Kind: game.CounterPlusOne, N: 1},
		Instruction(func(ctx *Context) error {
			if ctx.Item.Params.Amount < 5 {
				return nil
			}
			info, ok := ctx.SourcePermanent()
			if !ok || info.Power <= 0 {
				return nil
			}
			return AddMana{Amount: ManaN(info.Power, "R")}.Apply(ctx)
		}),
	))
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:        "485233bf-dc4b-4f31-8dfb-5054aee70c08",
		Name:            "Molten-Core Maestro",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered:       []game.TriggeredAbility{t},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bone Devourer — Creature — Dragon {3}{B}, 2/2 (Tarkir: Dragonstorm
// Commander):
//
//	"Flash
//	 Flying
//	 This creature enters with a number of +1/+1 counters on it equal to the
//	 number of creatures that died this turn.
//	 When this creature dies, you draw X cards and you lose X life, where X is
//	 the number of +1/+1 counters on it."
//
// The counters are Blacksnag Buzzard's CR 614.1c self-replacement counted off
// the table-wide tally of creatures that died this turn (b11CreaturesDiedThisTurn,
// anyone's). X is the number of +1/+1 counters the Devourer had as it last
// existed on the battlefield (CR 608.2h, the ruling), the leaving-the-battlefield
// counter snapshot, so -1/-1 counters that killed it don't hide the +1/+1
// ones. The draw settles before the life is lost.
//
// No simplification.
func init() {
	const label = "Bone Devourer — draw X cards and lose X life, X the number of +1/+1 counters on it"
	dies := On(game.EventLTB, ThisDied, label, Do(
		Instruction(func(ctx *Context) error {
			x := ctx.Item.Params.Amount
			if x <= 0 {
				return nil
			}
			return DrawCards{N: x}.Apply(ctx)
		}),
		Instruction(func(ctx *Context) error {
			x := ctx.Item.Params.Amount
			if x <= 0 {
				return nil
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -x)
		}),
	))
	// X is fixed as the ability triggers, from the counters the Devourer had as
	// it left the battlefield (CR 608.2h); a fill-in Build records it.
	dies.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne]
		return item
	}
	Register(Spec{
		OracleID:        "0ee41ef9-0b39-4d00-a68e-7914839a7175",
		Name:            "Bone Devourer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, _ *game.Card) int {
				return b11CreaturesDiedThisTurn(g)
			}, "Bone Devourer enters with a number of +1/+1 counters equal to the number of creatures that died this turn"),
		},
		Triggered: []game.TriggeredAbility{dies},
	})
}

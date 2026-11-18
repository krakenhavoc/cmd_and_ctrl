package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kitsune Ace — Creature — Fox Pilot {1}{W}, 2/2 (Kamigawa: Neon Dynasty,
// S82):
//
//	"Whenever a Vehicle you control attacks, choose one —
//	 • That Vehicle gains first strike until end of turn.
//	 • Untap this creature."
//
// Miriam, Herd Whisperer's per-attacker trigger (a Vehicle you control, read
// on the attacker's effective subtypes; the attacking object is carried on the
// item) with a two-way mode. "That Vehicle" is the attacking object: a Vehicle
// that has left the battlefield, or come back as a new object, gets nothing
// (CR 400.7). "This creature" is the Ace, untapped only while it is still the
// object that triggered.
//
// No simplification.
func init() {
	const label = "Kitsune Ace — a Vehicle attacked"
	t := On(game.EventAttack, vehicleYouControlAttacked, label, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Build = eventCardTriggerItem(label)
	t.Modes = ChooseOne(
		ModeDoing("That Vehicle gains first strike until end of turn.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				o := item.Params.Object
				c, ok := ctx.Game.LookupCardForEffect(o.ID)
				if !ok || c.ObjectEpoch != o.Epoch || !onBattlefield(ctx.Game, o.ID) {
					return nil
				}
				return GrantKeywordUntilEOT{Target: o.ID, Keywords: []string{"first strike"}, Label: "Kitsune Ace — first strike"}.Apply(ctx)
			}),
		ModeDoing("Untap this creature.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if sourceIsNewObject(ctx.Game, item) || !onBattlefield(ctx.Game, item.SourceCardID) {
					return nil
				}
				return UntapTarget{Target: item.SourceCardID}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "1dbadf75-28cd-4525-8f51-9c53049b9b54",
		Name:         "Kitsune Ace",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

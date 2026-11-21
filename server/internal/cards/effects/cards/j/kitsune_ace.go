package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	t := effects.On(game.EventAttack, effects.VehicleYouControlAttacked, label, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Build = effects.EventCardTriggerItem(label)
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("That Vehicle gains first strike until end of turn.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				o := item.Params.Object
				c, ok := ctx.Game.LookupCardForEffect(o.ID)
				if !ok || c.ObjectEpoch != o.Epoch || !effects.OnBattlefield(ctx.Game, o.ID) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: o.ID, Keywords: []string{"first strike"}, Label: "Kitsune Ace — first strike"}.Apply(ctx)
			}),
		effects.ModeDoing("Untap this creature.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				if effects.SourceIsNewObject(ctx.Game, item) || !effects.OnBattlefield(ctx.Game, item.SourceCardID) {
					return nil
				}
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "1dbadf75-28cd-4525-8f51-9c53049b9b54",
		Name:         "Kitsune Ace",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

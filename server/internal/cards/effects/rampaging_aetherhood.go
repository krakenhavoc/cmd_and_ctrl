package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rampaging Aetherhood — Creature — Snake Hydra {4}{G}, 4/4:
//
//	"Trample, ward {2}
//	 At the beginning of your upkeep, you get an amount of {E} (energy
//	 counters) equal to this creature's power. Then you may pay one or
//	 more {E}. If you do, put that many +1/+1 counters on this creature."
//
// ADR 0129 §3 (#1995, owner decision 3): the power is read as the
// trigger resolves (CR 608.2h), its last-known power if it has left.
// "One or more" is the pay_amount prompt with a floor of 1; the bot
// pays everything it has, since every counter is another +1/+1. The
// counters go on the Hydra only if it is still the permanent that
// triggered (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9b1f0780-072a-4fe6-91d8-e546a8aaea20",
		Name:            "Rampaging Aetherhood",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Rampaging Aetherhood — ward {2}"),
			AtYourUpkeep("Rampaging Aetherhood — get {E} equal to its power, then may pay one or more {E}",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					power := b13LastKnownPower(g, item.SourceCardID, game.Characteristic{})
					if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && sourceIsStillThisPermanent(g, item) {
						power = c.CurrentPower()
					}
					if err := (GetEnergy{N: power}).Apply(ctx); err != nil {
						return err
					}
					return PayEnergyAmount{
						Min:      1,
						Question: "Rampaging Aetherhood — pay one or more {E} for that many +1/+1 counters",
						Unit:     game.PayAmountCounters,
						Goal:     AsMuchAsYouCan,
						Then: func(ctx *Context, paid int) error {
							if paid <= 0 || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
								return nil
							}
							return plusOneCountersOnThis(paid)(ctx.Game, ctx.Item)
						},
					}.Apply(ctx)
				}),
		},
	})
}

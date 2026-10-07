package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assaultron Dominator — Artifact Creature — Robot {1}{R}, 2/2:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 Whenever an artifact creature you control attacks, you may pay {E}.
//	 If you do, put your choice of a +1/+1, first strike, or trample
//	 counter on that creature."
//
// ADR 0129 §3 (#1995): one trigger per attacking artifact creature you
// control, the Dominator included. The energy is paid as it resolves
// (CR 118.12), and the counter kind is chosen after paying, as the
// instruction says; it goes on the attacker if it is still on the
// battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c7818f4-a818-4603-9af1-05aee2d56ec8",
		Name:         "Assaultron Dominator",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Assaultron Dominator", 2),
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventAttack {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Controller == source.Controller && c.IsArtifact() && c.IsCreature()
			}, "Assaultron Dominator — you may pay {E} for a counter on that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					attacker := item.Trigger.Event.CardID
					return MayPayEnergy{
						N:        1,
						Question: "Assaultron Dominator — pay {E} to put a +1/+1, first strike or trample counter on that creature?",
						OnPay: func(ctx *Context) error {
							return putChosenCounterOn(ctx, ctx.Controller(), attacker,
								"Assaultron Dominator — choose a counter for that creature",
								[]string{game.CounterPlusOne, game.CounterFirstStrike, game.CounterTrample})
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

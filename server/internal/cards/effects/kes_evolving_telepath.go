package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kes, Evolving Telepath — {4}{U} Legendary Creature — Ocampa Advisor
// 2/6 (Star Trek Commander, #4974):
//
//	"At the beginning of your end step, put a time counter on Kes, then
//	 draw X cards, where X is the number of time counters on her. Then
//	 if she has nine or more time counters on her, sacrifice her."
//
// One trigger, three instructions in order. X is read after the counter
// lands, from Kes as she is now, or as she last existed on the
// battlefield if she left in response (CR 608.2h), in which case the
// counter itself does nothing (CR 608.2b) and the draw still happens.
// The sacrifice re-reads her after the draw and acts only on the same
// object still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "714dcaf0-442b-4849-917a-7e6af2ea0210",
		Name:         "Kes, Evolving Telepath",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Kes, Evolving Telepath — put a time counter on Kes, then draw a card for each time counter on her; with nine or more, sacrifice her",
				kesEndStep),
		},
	})
}

// kesEndStep is the trigger's body.
func kesEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterTime, N: 1}).Apply(ctx); err != nil {
		return err
	}
	self, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	if err := (DrawCards{N: self.Counters[game.CounterTime]}).Apply(ctx); err != nil {
		return err
	}
	if self, ok = ctx.SourcePermanent(); !ok || self.Left || self.Counters[game.CounterTime] < 9 {
		return nil
	}
	return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
}

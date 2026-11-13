package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beta Ray Bill — Legendary Creature — Alien Warrior Hero {2}{G/W}, 2/3
// (Secret Lair Countdown, #5463):
//
//	"Valiant — Whenever Beta Ray Bill becomes the target of a spell or
//	 ability you control for the first time each turn, put a +1/+1
//	 counter on him and up to one target God you control."
//
// The shared valiant trigger (valiant.go), with an "up to one" target
// clause: a God you control, any permanent with the subtype. With no God
// the trigger still goes on the stack with no target and Bill gets his
// counter. With a God chosen that is no longer legal at resolution, the
// ability's only target is illegal and it does nothing at all, Bill's
// counter included (CR 608.2b). Bill himself is not a target, so his
// counter is put whether or not a God was chosen; if he has left the
// battlefield in response, he gets none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebb58466-dca1-4a5a-83fe-f48d75063cbe",
		Name:         "Beta Ray Bill",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(Valiant("Beta Ray Bill — Valiant: put a +1/+1 counter on him and up to one target God you control", betaRayBillCounters),
				TargetPermanent("up to one target God you control", Subtype("God"), YouControl()).WithCount(0, 1)),
		},
	})
}

// betaRayBillCounters puts a +1/+1 counter on Bill, while he is still the
// permanent that triggered, and one on the God if it is still a legal
// target.
func betaRayBillCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	return nil
}

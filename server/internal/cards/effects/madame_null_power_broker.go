package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Madame Null, Power Broker — {2}{B} Legendary Creature — Demon Advisor 1/3
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Deathtouch
//	 Whenever another creature you control enters, you may pay life equal
//	 to its power. If you do, put that many +1/+1 counters on it."
//
// The alliance trigger (WheneverAnotherCreatureEntersUnderYourControl)
// with Arni, Renowned Champion's read of the entering creature's power as
// the trigger resolves (its last-known power if it has gone). The payment
// is Bringer of the Black Dawn's "you may pay N life" asked as the
// trigger resolves (MayChoice with LifeCost, CR 608.2d), re-checked when
// the answer arrives (CR 119.4): a player with less life than that can't
// pay, and nothing happens. "That many" is the life paid. The creature
// rides the prompt as an object with its epoch, so one that left and came
// back while the question was open is a new object and gets no counters
// (CR 400.7), though the life, once paid, stays paid.
//
// A creature with power 0 or less asks nothing: paying 0 life to put no
// counters on it changes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "73413bd5-8c40-4d45-8f05-c20c39ee4017",
		Name:            "Madame Null, Power Broker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Madame Null, Power Broker — you may pay life equal to its power for that many +1/+1 counters",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					obj := ctx.Trigger().Object
					if !ok || obj == nil || info.Power <= 0 {
						return nil
					}
					return MayChoice{
						Question: "Madame Null, Power Broker — pay " + strconv.Itoa(info.Power) + " life to put that many +1/+1 counters on it?",
						LifeCost: info.Power,
						Then:     madameNullPaid,
						Objects:  []game.ObjectRef{obj.Ref()},
					}.Apply(ctx)
				}),
		},
	})
}

// madameNullPaid is the "yes": pay the life, and only if it was paid put
// that many +1/+1 counters on the creature, while it is still the object
// that entered. Keyed (#4291).
var madameNullPaid = MayYes("confirm/madame-null-pay-life-for-counters", func(ctx *Context, r game.ConfirmAnswer) error {
	n := r.LifeCost
	p := ctx.Game.PlayerByIDForEffect(r.Chooser)
	if n <= 0 || p == nil || p.Eliminated || p.Life < n {
		return nil
	}
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), r.Chooser, n); err != nil {
		return err
	}
	if len(r.Objects) != 1 {
		return nil
	}
	ref := r.Objects[0]
	c, ok := ctx.Game.LookupCardForEffect(ref.ID)
	if !ok || !onBattlefield(ctx.Game, ref.ID) || c.ObjectEpoch != ref.Epoch {
		return nil
	}
	return AddCounter{Target: ref.ID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
})

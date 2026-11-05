package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ensign Harry Kim, Operations — Legendary Creature — Human Officer
// {1}{W}, 1/2 (Star Trek Commander, #4974):
//
//	"Whenever another creature you control enters, if it has greater power
//	 than Ensign Kim, put a +1/+1 counter on Ensign Kim.
//	 {2}, Remove two +1/+1 counters from among creatures you control: Draw
//	 a card."
//
// The trigger's "if" is an intervening one (CR 603.4): the entering
// creature's power is compared with Kim's, both through the layers and
// counters included, as the creature enters and again as the ability
// resolves, with last-known information for a creature that has left
// (CR 608.2h). A counter in response that makes Kim as big removes the
// trigger. The activated cost is Hopeful Initiate's: two +1/+1 counters
// taken from any creatures you control in any split, Kim's own included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9bc0f83-c522-44c9-baa6-fffdeb02a0f5",
		Name:         "Ensign Harry Kim, Operations",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, ensignKimBiggerEntered,
				"Ensign Harry Kim — put a +1/+1 counter on Ensign Kim",
				Do(If{Cond: CondBoard{Name: "the entering creature has greater power than Ensign Kim", Test: ensignKimStillBigger},
					Then: []Applier{CounterOnThis{Kind: game.CounterPlusOne, N: 1}}})),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}, Remove two +1/+1 counters from among creatures you control: Draw a card.",
			Cost: Plus(
				ManaCost("{2}"),
				RemoveCountersAmong(game.CounterPlusOne, 2, "creatures you control", Creature()),
			),
			Effect:  drawOneCardEffect,
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
		}},
	})
}

// ensignKimBiggerEntered is "another creature you control enters, if it
// has greater power than Ensign Kim", as it triggers.
func ensignKimBiggerEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.IsCreature() && c.PowerForComparison() > source.PowerForComparison()
}

// ensignKimStillBigger is the same "if" as the ability resolves.
func ensignKimStillBigger(ctx *Context) bool {
	entered, ok := ctx.TriggeringPermanent()
	if !ok {
		return false
	}
	kim, ok := ctx.SourcePermanent()
	return ok && entered.Power > kim.Power
}

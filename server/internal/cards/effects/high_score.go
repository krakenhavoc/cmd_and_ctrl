package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// High Score — {2}{G} Enchantment (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"If one or more +1/+1 counters would be put on a creature you control,
//	 that many plus one +1/+1 counters are put on it instead.
//	 At the beginning of your end step, draw a card if you control a
//	 creature with the greatest power among creatures on the battlefield."
//
// The first ability is Hardened Scales: a CR 614.1a replacement over a
// +1/+1 placement on a creature its controller controls
// (plusOneCounterPlacementOnYourCreature), so a creature entering with
// counters gets one more too, two High Scores each add one, and beside a
// doubler the affected creature's controller orders them (CR 616.1).
//
// The draw is not an intervening "if" (CR 603.4: the condition does not
// follow the trigger event), so the ability triggers at every one of its
// controller's end steps and asks only as it resolves whether that player
// controls a creature whose power, through the layers, is the greatest (or
// tied for the greatest) among all creatures on the battlefield. With no
// creature on the battlefield nobody does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d81fc181-ecb0-43a5-88e8-c61aca3428d8",
		Name:         "High Score",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			{
				Watches:   []game.EventKind{game.EventCounterPlaced},
				AppliesTo: plusOneCounterPlacementOnYourCreature,
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.CounterDelta++
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "High Score: +1 +1/+1 counter",
			},
		},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("High Score — draw a card if you control a creature with the greatest power", Do(If{
				Cond: CondBoard{Name: "a creature with the greatest power among creatures on the battlefield", Test: func(ctx *Context) bool {
					return highScoreControlsGreatestPower(ctx.Game, ctx.Controller())
				}},
				Then: []Applier{DrawCards{N: 1}},
			})),
		},
	})
}

// highScoreControlsGreatestPower reports whether `controller` controls a
// creature whose power is the greatest, or tied for the greatest, among
// every creature on the battlefield.
func highScoreControlsGreatestPower(g *game.Game, controller uuid.UUID) bool {
	if g.Battlefield == nil {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	best, found, mine := 0, false, false
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsCreature() {
			continue
		}
		p := c.PowerForComparison()
		switch {
		case !found || p > best:
			best, found, mine = p, true, c.Controller == controller
		case p == best && c.Controller == controller:
			mine = true
		}
	}
	return mine
}

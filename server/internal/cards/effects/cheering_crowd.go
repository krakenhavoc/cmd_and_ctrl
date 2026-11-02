package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cheering Crowd — Creature — Human Citizen {1}{R/G}, 2/2 (Marvel's
// Spider-Man):
//
//	"At the beginning of each player's first main phase, that player may
//	 put a +1/+1 counter on this creature. If they do, they add {C} for each
//	 counter on it."
//
// The trigger is the Crowd's controller's, but every choice and the mana
// belong to the player whose main phase it is (the step event's actor),
// asked as a keyed "may" (MayChoice with that player). The first main
// phase is the precombat main (Hulking Raptor's reading). "If they do"
// is the counter really arriving: a Crowd that has left by then gives
// nothing. The mana counts every counter on the Crowd, of any kind, after
// the new one is on it, and goes to that player (CR 106.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e11839d5-99fe-4cc8-802e-be669ad274ee",
		Name:         "Cheering Crowd",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginPrecombatMain, AnyPlayer,
				"Cheering Crowd — that player may put a +1/+1 counter on it; if they do, they add {C} for each counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Player:   item.Trigger.Event.Actor,
						Question: "Cheering Crowd — put a +1/+1 counter on it and add {C} for each counter on it?",
						YesLabel: "Put the counter",
						NoLabel:  "Don't",
						Then:     cheeringCrowdAnswer,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// cheeringCrowdAnswer is the "yes": the counter, then {C} for each counter on
// the Crowd to the player who chose. Keyed, an on-disk identity.
var cheeringCrowdAnswer = MayYes("confirm/spm-cheering-crowd", func(ctx *Context, r game.ConfirmAnswer) error {
	g := ctx.Game
	if !onBattlefield(g, r.Source) {
		return nil
	}
	if err := (AddCounter{Target: r.Source, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(r.Source)
	if !ok {
		return nil
	}
	n := 0
	for _, k := range c.Counters {
		if k > 0 {
			n += k
		}
	}
	if n == 0 {
		return nil
	}
	return g.AddManaForEffect(r.Chooser, r.Source, strings.Repeat("{C}", n))
})

package cardsch

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "e11839d5-99fe-4cc8-802e-be669ad274ee",
		Name:         "Cheering Crowd",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginPrecombatMain, effects.AnyPlayer,
				"Cheering Crowd — that player may put a +1/+1 counter on it; if they do, they add {C} for each counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Player:   item.Trigger.Event.Actor,
						Question: "Cheering Crowd — put a +1/+1 counter on it and add {C} for each counter on it?",
						YesLabel: "Put the counter",
						NoLabel:  "Don't",
						Then:     cheeringCrowdAnswer,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// cheeringCrowdAnswer is the "yes": the counter, then {C} for each counter on
// the Crowd to the player who chose. Keyed, an on-disk identity.
var cheeringCrowdAnswer = effects.MayYes("confirm/spm-cheering-crowd", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	g := ctx.Game
	if !effects.OnBattlefield(g, r.Source) {
		return nil
	}
	if err := (effects.AddCounter{Target: r.Source, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
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

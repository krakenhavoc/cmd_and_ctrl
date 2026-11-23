package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sharktocrab — Creature — Shark Octopus Crab {2}{G}{U}, 4/4 (Ravnica
// Remastered, #5948):
//
//	"{2}{G}{U}: Adapt 1.
//	 Whenever one or more +1/+1 counters are put on this creature, tap
//	 target creature an opponent controls. That creature doesn't untap
//	 during its controller's next untap step."
//
// Benthic Biomancer's shape: adapt checks for counters as it resolves
// (CR 701.46a), and the trigger is one per batch of counters put on this
// creature (CR 603.2c), by anyone. The tap and the untap restriction are one
// resolution on the one target, which is tapped even if it already was; the
// restriction names the target's controller's next untap step (CR 502.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb8e43f2-41d5-4ffb-a61e-940477749e00",
		Name:         "Sharktocrab",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}{U}: Adapt 1",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{G}{U}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.AdaptN(g, item, 1) },
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OncePerBatch(effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && ev.Target == source.InstanceID && ev.Label == game.CounterPlusOne && effects.B39CountersPlacedOnThisObject(ev, g) > 0
			}, "Sharktocrab — tap target creature an opponent controls; it doesn't untap during its controller's next untap step",
				tapTargetAndItDoesntUntap("Sharktocrab"))),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}

// tapTargetAndItDoesntUntap is "tap target creature. That creature doesn't
// untap during its controller's next untap step."
func tapTargetAndItDoesntUntap(name string) effects.Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.DoesntUntapNextUntapStep{Targets: []uuid.UUID{t.ID}, Label: name + " — doesn't untap during its controller's next untap step"}.Apply(ctx)
		}
		return nil
	}
}

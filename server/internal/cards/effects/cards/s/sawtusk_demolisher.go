package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sawtusk Demolisher — Creature — Beast {4}{G}{G}, 6/6 (Commander 2020,
// #5513):
//
//	"Mutate {3}{G} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature
//	 on top plus all abilities from under it.)
//	 Trample
//	 Whenever this creature mutates, destroy target noncreature permanent.
//	 Its controller creates a 3/3 green Beast creature token."
//
// Mutate is the ADR 0223 seam (mutate.go), and the trigger is Everquill
// Phoenix's, firing whether the Demolisher went over or under. Its effect is
// Beast Within's: the permanent's controller is read before it is destroyed
// and gets the Beast whether or not the destruction happened (an
// indestructible target still gives its controller the token). A target
// gone by resolution means the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "7f93a368-9568-456c-9bd6-20d5fe5ad2f1",
		Name:             "Sawtusk Demolisher",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{3}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Sawtusk Demolisher — destroy target noncreature permanent; its controller creates a 3/3 green Beast",
				effects.Do(effects.OnTheTargetCard(sawtuskDemolish))),
				effects.TargetPermanent("target noncreature permanent", effects.Noncreature())),
		},
	})
}

// sawtuskDemolish destroys the target and gives its controller the Beast.
func sawtuskDemolish(ctx *effects.Context, id uuid.UUID) error {
	controller, ok := effects.ControllerOfTarget(ctx, id)
	if !ok {
		return nil
	}
	if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return effects.CreateToken{Controller: controller, Template: effects.TokenCard("3/3 green Beast"), N: 1}.Apply(ctx)
}

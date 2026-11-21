package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Souvenir Snatcher — Creature — Bird {4}{U}, 4/4 (Commander 2020, #5513):
//
//	"Mutate {5}{U} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature
//	 on top plus all abilities from under it.)
//	 Flying
//	 Whenever this creature mutates, gain control of target noncreature
//	 artifact."
//
// Mutate is the ADR 0223 seam (mutate.go), and the trigger is Everquill
// Phoenix's, firing whether the Snatcher went over or under. The theft is
// Thieving Skydiver's: no duration is printed, so it lasts indefinitely (CR
// 611.2a), and it is the trigger's controller who gains control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "ee8fb1d0-db86-489f-8d23-4238b2e8eba4",
		Name:             "Souvenir Snatcher",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{5}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Souvenir Snatcher — gain control of target noncreature artifact",
				effects.Do(effects.OnTheTargetCard(souvenirSnatcherSteal))),
				effects.TargetPermanent("target noncreature artifact", effects.Artifact(), effects.Noncreature())),
		},
	})
}

// souvenirSnatcherSteal takes the artifact for good.
func souvenirSnatcherSteal(ctx *effects.Context, id uuid.UUID) error {
	return effects.GainControl{
		Target:     id,
		Controller: ctx.Controller(),
		Duration:   game.IndefiniteDuration(),
		Label:      "Souvenir Snatcher — gain control (no stated duration)",
	}.Apply(ctx)
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fortune, Loyal Steed — Legendary Creature — Beast Mount {2}{W}, 2/4:
//
//	"When Fortune enters, scry 2.
//	 Whenever Fortune attacks while saddled, at end of combat, exile it
//	 and up to one creature that saddled it this turn, then return
//	 those cards to the battlefield under their owner's control.
//	 Saddle 1"
//
// The attack trigger schedules a delayed trigger at the beginning of the
// end of combat step (CR 603.7, 511.2) whose source is Fortune's object
// (ADR 0186 §1), the shape Dorothea's sacrifice uses. When it resolves
// the creature is NOT targeted (the 2024-04-12 ruling): the controller
// picks up to one of SaddlersOfThis in a card-set prompt, so the pick is
// made as the ability resolves and works whether or not Fortune is still
// on the battlefield. Fortune is exiled only if it is still the same
// object (CR 400.7) and, with the chosen creature, comes back as one
// entry under its owner's control (the flicker return). A token that
// saddled it is exiled and does not return (CR 111.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc440b62-0bf0-4f23-9210-3550ff0a8d4d",
		Name:         "Fortune, Loyal Steed",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(1)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fortune, Loyal Steed — scry 2", Do(Scry{N: 2})),
			AttacksWhileSaddled("Fortune, Loyal Steed — at end of combat, exile it and a creature that saddled it, then return them",
				func(g *game.Game, item *game.StackItem) error {
					return ScheduleDelayedTrigger{
						At:    game.StepEndCombat,
						Label: "Fortune, Loyal Steed — exile it and up to one creature that saddled it, then return them",
						Body:  fortuneEndOfCombatBody,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// fortuneEndOfCombatBody is the delayed trigger. Assigned in init for the
// reason dorotheaSacrificeBody is: a var initialiser would be an
// initialisation cycle through the exit primitives.
var fortuneEndOfCombatBody game.BodyRef

func init() {
	fortuneEndOfCombatBody = game.SimpleDelayedBody("fortune-loyal-steed/end-of-combat", fortuneEndOfCombat)
}

func fortuneEndOfCombat(g *game.Game, item *game.StackItem) error {
	saddlers := SaddlersOfThis(NewContext(g, item))
	if len(saddlers) == 0 {
		return fortuneExileAndReturn(g, item, nil)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Fortune, Loyal Steed — exile up to one creature that saddled it along with Fortune",
		Cards:    saddlers,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			return fortuneExileAndReturn(g, item, picked)
		},
	})
	return nil
}

// fortuneExileAndReturn exiles Fortune (if still the same object) and the
// chosen creatures as one event, then returns what reached exile.
func fortuneExileAndReturn(g *game.Game, item *game.StackItem, chosen []uuid.UUID) error {
	var ids []uuid.UUID
	if !g.AbilitySourceGoneForEffect(item) {
		ids = append(ids, item.SourceCardID)
	}
	ids = append(ids, chosen...)
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
		return ReturnFromExileTogether{Targets: exiled}.Apply(NewContext(g, item))
	})
}

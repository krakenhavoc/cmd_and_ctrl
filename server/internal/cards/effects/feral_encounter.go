package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Feral Encounter — Sorcery {G}{G} (Wilds of Eldraine, #4471):
//
//	"Look at the top five cards of your library. You may exile a creature
//	 card from among them. Put the rest on the bottom of your library in a
//	 random order. You may cast the exiled card this turn. At the beginning
//	 of the next combat phase this turn, target creature you control deals
//	 damage equal to its power to up to one target creature you don't
//	 control."
//
// The look is private; the pick is a floor-of-none choice of one creature
// card (The Key to the Vault's pick), exiled with a this-turn permission to
// cast it (CR 601), and the rest go to the bottom in a random order. The
// last sentence is a delayed trigger, created whatever the pick was
// (CR 603.7a), on the first beginning-of-combat step this turn (the Wrath
// of Khan's condition) and gone at the end of the turn if no combat comes.
// Its two target clauses are chosen as it goes on the stack, not when the
// spell is cast (CR 603.3d, the first ruling): a creature you control
// (required: with none the trigger is removed) and up to one you don't
// control. The first deals damage equal to its power as it resolves, only
// to a victim still legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a8196ef-24ee-40b6-97d1-8a2257180f86",
		Name:         "Feral Encounter",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DelayedOnEvent{
				Label:     "Feral Encounter — target creature you control deals damage equal to its power to up to one target creature you don't control",
				On:        []game.EventKind{game.EventStepBegan},
				Condition: theWrathOfKhanNextCombatCondition,
				Body:      feralEncounterBiteBody,
			}).Apply(ctx); err != nil {
				return err
			}
			g, controller := ctx.Game, ctx.Controller()
			looked := g.LookAtTopOfLibraryForEffect(controller, 5)
			if len(looked) == 0 {
				return nil
			}
			var eligible []uuid.UUID
			for _, id := range looked {
				if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
					eligible = append(eligible, id)
				}
			}
			if len(eligible) == 0 {
				return feralEncounterFinish(g, controller, ctx.Source(), looked, nil)
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  controller,
				Source:   ctx.Source(),
				Question: "Feral Encounter — you may exile a creature card from among them (you may cast it this turn)",
				Cards:    eligible,
				Min:      0,
				Max:      1,
				Zone:     game.ZoneLibrary,
				ThenKey:  feralEncounterPicked,
				Carry:    looked,
			})
			return nil
		},
	})
}

// feralEncounterPicked is the answer to the pick: the cards looked at are
// its Carry. An on-disk key.
var feralEncounterPicked = game.RegisterChooseCardsThen("choose-cards/feral-encounter", func(g *game.Game, r game.CardsChosen) error {
	return feralEncounterFinish(g, r.Chooser, r.Source, r.Carry, r.Picked)
})

// feralEncounterFinish exiles the chosen card, then puts the rest on the
// bottom in a random order and lets the controller cast the exiled card
// this turn.
func feralEncounterFinish(g *game.Game, controller, source uuid.UUID, looked, picked []uuid.UUID) error {
	var chosen uuid.UUID
	if len(picked) > 0 {
		chosen = picked[0]
	}
	var rest []uuid.UUID
	for _, id := range cardsStillInALibrary(g, looked) {
		if id != chosen {
			rest = append(rest, id)
		}
	}
	tail := game.KeyedTail{Ref: feralEncounterExiled, Player: controller, Source: source, Carry: rest}
	if chosen == uuid.Nil {
		return tail.Run(g, nil)
	}
	return g.ExileCardKeyedForEffect(chosen, tail)
}

// feralEncounterExiled is the exile's tail: the rest (its Carry) to the
// bottom in a random order, then the cast permission over the exiled card
// for the rest of the turn. An on-disk key.
var feralEncounterExiled = game.RegisterTail("then/feral-encounter-bottom-then-permission", func(g *game.Game, o game.TailOutcome) error {
	if err := g.PutOnBottomInRandomOrderForEffect(o.Tail.Player, game.ZoneLibrary, o.Tail.Carry); err != nil {
		return err
	}
	if len(o.IDs) == 1 {
		g.GrantCastPermissionOverCardForEffect(o.IDs[0], game.CastPermission{
			Player: o.Tail.Player,
			Zone:   game.ZoneExile,
			Source: o.Tail.Source,
			Label:  "Feral Encounter — you may cast the exiled card this turn",
		})
	}
	return nil
})

// feralEncounterBiteBody is the delayed trigger's body, registered with its
// two target clauses (ADR 0187 amendment 2026-10-16).
var feralEncounterBiteBody = game.TargetedDelayedBody("feral-encounter/bite",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		biter, ok := ctx.ClauseTarget(0)
		if !ok || biter.Kind != game.TargetCard {
			return nil
		}
		victim, ok := ctx.ClauseTarget(1)
		if !ok || victim.Kind != game.TargetCard {
			return nil
		}
		c, found := g.LookupCardForEffect(biter.ID)
		if !found {
			return nil
		}
		return DealDamage{Source: biter.ID, Target: victim.ID, Amount: c.CurrentPower()}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec {
		return Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("up to one target creature you don't control", OpponentControls()).WithCount(0, 1),
		)
	}))

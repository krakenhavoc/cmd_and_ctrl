package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tamiyo, Field Researcher — Legendary Planeswalker — Tamiyo {1}{G}{W}{U},
// starting loyalty 4:
//
//	"+1: Choose up to two target creatures. Until your next turn,
//	 whenever either of those creatures deals combat damage, you draw a
//	 card.
//	 −2: Tap up to two target nonland permanents. They don't untap during
//	 their controller's next untap step.
//	 −7: Draw three cards. You get an emblem with "You may cast nonland
//	 cards from your hand without paying their mana costs.""
//
// The +1 is a delayed triggered ability with a stated duration (Garruk,
// Curse Breaker's −4 is the template), watching combat damage dealt by
// either chosen creature as the objects they were when it was created. It
// is an each-condition with OncePerBatch: each creature that deals combat
// damage draws once for that damage step however many players, creatures
// or planeswalkers it hurt (CR 603.2c), and both creatures drawing is two
// cards. The emblem offers Omniscience's free cast from hand to its owner
// for every nonland card; a land is not cast.
func init() {
	Register(Spec{
		OracleID:        "9cf99129-f51a-4418-a8cc-8ca2992b17fa",
		Name:            "Tamiyo, Field Researcher",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label:                   "Tamiyo, Field Researcher emblem",
			Text:                    "You may cast nonland cards from your hand without paying their mana costs.",
			GrantedAlternativeCosts: []game.GrantedAlternativeCost{CastFromHandWithoutPayingManaCost()},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Choose up to two target creatures. Until your next turn, whenever either of those creatures deals combat damage, you draw a card.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to two target creatures").WithCount(0, 2),
				Effect:  tamiyoFieldResearcherPlusOne,
			},
			{
				Label:   "−2: Tap up to two target nonland permanents. They don't untap during their controller's next untap step.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetPermanent("up to two target nonland permanents", Nonland()).WithCount(0, 2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b751TapFreezeTargets(NewContext(g, item), uuid.Nil, "Tamiyo, Field Researcher")
				},
			},
			{
				Label:   "−7: Draw three cards. You get an emblem with \"You may cast nonland cards from your hand without paying their mana costs.\"",
				Cost:    LoyaltyCost(-7),
				Purpose: game.Purpose{Draws: 3},
				Effect:  Do(DrawCards{N: 3}, CreateEmblem{}),
			},
		},
	})
}

// tamiyoEitherDealsCombatDamageCondition is "whenever either of those
// creatures deals combat damage": the damaging creature, when it is one of
// CondParams.Objects and still that object on the battlefield (CR 603.7c).
var tamiyoEitherDealsCombatDamageCondition = game.DelayedEachCondition(
	"damage/either-of-those-creatures-deals-combat-damage", tamiyoEitherDealsCombatDamage)

func tamiyoEitherDealsCombatDamage(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return nil
	}
	for _, o := range p.Objects {
		if o.ID != ev.Source {
			continue
		}
		if c, ok := battlefieldCardForEffect(g, o.ID); ok && c.ObjectEpoch == o.Epoch {
			return []uuid.UUID{o.ID}
		}
	}
	return nil
}

func tamiyoFieldResearcherPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var objs []game.ObjectRef
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if c, ok := battlefieldCardForEffect(g, t.ID); ok {
			objs = append(objs, game.ObjectRef{ID: t.ID, Epoch: c.ObjectEpoch})
		}
	}
	if len(objs) == 0 {
		return nil
	}
	controller := ctx.Controller()
	until := DurationUntilYourNextTurn(ctx, controller)
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: ctx.Source(),
		Label:        "Tamiyo, Field Researcher — you draw a card",
		On:           []game.EventKind{game.EventDealDamage},
		Condition:    tamiyoEitherDealsCombatDamageCondition,
		CondParams:   game.EffectParams{Objects: objs},
		Body:         drawACardForTheControllerBody,
		Repeats:      true,
		OncePerBatch: true,
		Duration:     &until,
	})
	return nil
}

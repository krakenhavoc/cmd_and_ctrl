package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Northampton Farm — Land (Teenage Mutant Ninja Turtles, #4457):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Exile target creature you own.
//	 {2}, {T}, Sacrifice this land: Return a creature card exiled with
//	 this land to the battlefield under your control. Return each other
//	 card exiled with this land to its owner's hand."
//
// The exile is linked to the Farm object (CR 607.2a: ExileTarget.Linked,
// Negative Zone Portal's link), and "exiled with this land" is the cards
// in exile that carry that link. The sacrifice is a cost, so the third
// ability reads the link through the object it was activated from
// (Context.SourceRef, the epoch it had then), as The Creation of Avacyn's
// later chapters read theirs.
//
// "A creature card" is one card: with one creature card exiled with it,
// that one returns; with several, you choose which as the ability
// resolves. The rest, creature cards included, go to their owners' hands
// once the return has finished (a returning card that asks a question
// as it enters is answered first). A card that has left exile is not
// returned. A token exiled this way ceased to exist (CR 111.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "e05f1a43-16ce-4e88-b0ad-2202efb25516",
		Name:          "Northampton Farm",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{colorlessTapRow()},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Exile target creature you own.",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Targets: TargetCreature("target creature you own", YouOwn()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					target := FirstLegalBattlefieldTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					return ExileTarget{Target: target, Linked: true}.Apply(ctx)
				},
			},
			{
				Label:   "{2}, {T}, Sacrifice this land: Return a creature card exiled with this land to the battlefield under your control. Return each other card exiled with this land to its owner's hand.",
				Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  northamptonFarmReturn,
			},
		},
	})
}

// northamptonFarmReturn lists the cards exiled with the Farm object and
// returns them: one creature card to the battlefield (chosen when there is
// more than one), the rest to hand.
func northamptonFarmReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var linked, creatures []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		linked = append(linked, c.InstanceID)
		if c.IsCreature() {
			creatures = append(creatures, c.InstanceID)
		}
	}
	switch len(creatures) {
	case 0:
		return northamptonFarmPutBack(g, item.Controller, item.SourceCardID, uuid.Nil, linked)
	case 1:
		return northamptonFarmPutBack(g, item.Controller, item.SourceCardID, creatures[0], linked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Northampton Farm — choose a creature card to return to the battlefield",
		Cards:    creatures,
		Min:      1,
		Max:      1,
		ThenKey:  northamptonFarmPicked,
		Carry:    linked,
		LiveZone: game.ZoneExile,
	})
	return nil
}

// northamptonFarmPicked is the creature card chosen. An on-disk key.
var northamptonFarmPicked = game.RegisterChooseCardsThen("choose-cards/northampton-farm-return", func(g *game.Game, r game.CardsChosen) error {
	creature := uuid.Nil
	if len(r.Picked) == 1 {
		creature = r.Picked[0]
	}
	return northamptonFarmPutBack(g, r.Chooser, r.Source, creature, r.Carry)
})

// northamptonFarmPutBack returns `creature` (if any) to the battlefield
// under `controller`'s control, then every other card of `linked` to its
// owner's hand.
func northamptonFarmPutBack(g *game.Game, controller, source, creature uuid.UUID, linked []uuid.UUID) error {
	var rest []uuid.UUID
	for _, id := range linked {
		if id != creature {
			rest = append(rest, id)
		}
	}
	then := game.KeyedTail{Ref: northamptonFarmRestToHand, Player: controller, Source: source, Carry: rest}
	if creature == uuid.Nil || !northamptonInExile(g, creature) {
		return northamptonFarmRestToHandNow(g, then)
	}
	return g.ReturnFromExileToBattlefieldKeyedForEffect(creature, controller, false, then)
}

// northamptonFarmRestToHand is "return each other card exiled with this
// land to its owner's hand", once the creature has entered. An on-disk key.
var northamptonFarmRestToHand = game.RegisterTail("then/northampton-farm-rest-to-hand", func(g *game.Game, o game.TailOutcome) error {
	return northamptonFarmRestToHandNow(g, o.Tail)
})

func northamptonFarmRestToHandNow(g *game.Game, t game.KeyedTail) error {
	for _, id := range t.Carry {
		if !northamptonInExile(g, id) {
			continue
		}
		if _, err := g.ReturnFromAnyZoneToHandForEffect(t.Player, t.Source, id); err != nil {
			return err
		}
	}
	return nil
}

// northamptonInExile reports whether the card is in exile now.
func northamptonInExile(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneExile
}

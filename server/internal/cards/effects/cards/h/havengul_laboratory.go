package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Havengul Laboratory // Havengul Mystery — Legendary Land // Legendary Land
// (Secret Lair Drop, #5462):
//
//	Front: "{T}: Add {C}.
//	        {4}, {T}: Investigate.
//	        At the beginning of your end step, if you sacrificed three or more
//	        Clues this turn, transform Havengul Laboratory."
//	Back:  "When this land transforms into Havengul Mystery, return target
//	        creature card from your graveyard to the battlefield.
//	        When the creature put onto the battlefield with Havengul Mystery
//	        leaves the battlefield, transform Havengul Mystery.
//	        {T}, Pay 1 life: Add {B}."
//
// The end-step trigger is an intervening "if" (CR 603.4): the tally of Clues
// the controller sacrificed this turn (SacrificedWithSubtypeThisTurn, read as
// each was sacrificed, tokens included) is asked as the step begins and again
// as the ability resolves. The back's first trigger returns the creature and
// records it as the land's linked permanent (CR 607.2e: an object, with its
// CR 400.7 epoch); the second triggers when that creature leaves the
// battlefield, transforms the land and forgets it. Every transform is in
// place, so the land is one object throughout and its link survives.
//
// No simplification.
func init() {
	const oracle = "e71ac446-02a4-4468-8d29-f28b21617665"
	effects.Register(effects.Spec{
		OracleID:      oracle,
		Name:          "Havengul Laboratory",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.ColorlessTapRow()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Investigate.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect:  effects.Do(effects.Investigate{N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you sacrificed three or more Clues this turn", func(g *game.Game, you uuid.UUID) bool {
				return g.SacrificedWithSubtypeThisTurn(you, "Clue") >= 3
			}, effects.AtYourEndStep("Havengul Laboratory — if you sacrificed three or more Clues this turn, transform it", effects.Do(effects.TransformThis{}))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Havengul Mystery",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Life: 1},
			Produced: "{B}",
			Label:    "{T}, Pay 1 life: Add {B}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventTransform, effects.ThisTransformedInto(1),
					"Havengul Mystery — return target creature card from your graveyard to the battlefield", havengulMysteryReturn),
				effects.TargetCardInGraveyard("target creature card in your graveyard", effects.And(effects.Creature(), effects.YouOwn()))),
			effects.On(game.EventLTB, havengulMysteryCreatureLeft,
				"Havengul Mystery — the creature put onto the battlefield with it left, transform it", havengulMysteryTransform),
		},
	})
}

// havengulMysteryReturn returns the target creature card and records it as
// the land's linked permanent, as the new object it became.
func havengulMysteryReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	self, hasSelf := ctx.SourceRef()
	if hasSelf {
		g.SetLinkedPermanentForEffect(self.ID, game.ObjectRef{})
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
			return err
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok && hasSelf && effects.OnBattlefield(g, t.ID) {
			g.SetLinkedPermanentForEffect(self.ID, game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
		}
		return nil
	}
	return nil
}

// havengulMysteryCreatureLeft is "the creature put onto the battlefield with
// Havengul Mystery leaves the battlefield": the permanent leaving is the one
// this land recorded.
func havengulMysteryCreatureLeft(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	l := source.LinkedPermanent
	return l.ID != uuid.Nil && ev.CardID == l.ID
}

// havengulMysteryTransform forgets the creature and turns the land over.
func havengulMysteryTransform(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if self, ok := ctx.SourceRef(); ok {
		g.SetLinkedPermanentForEffect(self.ID, game.ObjectRef{})
	}
	return effects.TransformThis{}.Apply(ctx)
}

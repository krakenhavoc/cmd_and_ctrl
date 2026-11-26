package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sutured Ghoul — Creature — Zombie {4}{B}{B}{B}, */* (Magic 2012, S82 OLD
// batch 452, #5573):
//
//	"Trample
//	 As this creature enters, exile any number of creature cards from your
//	 graveyard.
//	 Sutured Ghoul's power is equal to the total power of the exiled cards
//	 and its toughness is equal to their total toughness."
//
// Mimeoplasm, Revered One's as-enters exile (CR 614.12), with "any number"
// as every creature card in the graveyard. Each pick is linked to this
// object (Card.ExiledWith, CR 607.2a), so a Ghoul that leaves and returns
// is a new object with no claim on them (CR 400.7). Its size is a layer 7a
// characteristic-defining ability (CR 604.3, 613.4a, #6269) summing the
// power and toughness of the cards exiled with this object; naming none
// makes it 0/0, and it dies to state-based actions.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8b4715b0-63d9-4ff8-a667-21a762518687",
		Name:            "Sutured Ghoul",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		AsEnters:        suturedGhoulExileAsEnters,
		Static: []game.StaticAbility{{
			Label:     "Sutured Ghoul's power is equal to the total power of the exiled cards and its toughness is equal to their total toughness.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				power, toughness := 0, 0
				for _, ex := range g.ExiledWithForEffect(game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}) {
					power += ex.Power
					toughness += ex.Toughness
				}
				c.Power, c.Toughness = power, toughness
			},
		}},
	})
}

// suturedGhoulExileAsEnters queues "exile any number of creature cards
// from your graveyard" for the controller; with no creature card there is
// nothing to ask.
func suturedGhoulExileAsEnters(card *game.Card, ctx *effects.Context) error {
	p := ctx.Game.PlayerByIDForEffect(card.Controller)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  card.Controller,
		Source:   card.InstanceID,
		Question: "Sutured Ghoul — exile any number of creature cards from your graveyard",
		Cards:    candidates,
		Min:      0,
		Max:      len(candidates),
		ThenKey:  suturedGhoulExiled,
		LiveZone: game.ZoneGraveyard,
	})
	return nil
}

// suturedGhoulExiled is the keyed answer: each card named is exiled linked
// to the Ghoul. An on-disk key: never renamed, never reused.
var suturedGhoulExiled = game.RegisterChooseCardsThen("choose-cards/s82-old-b452-sutured-ghoul-exile", func(g *game.Game, r game.CardsChosen) error {
	ref, ok := g.PermanentRefForEffect(r.Source)
	if !ok {
		return nil
	}
	for _, id := range r.Picked {
		if err := g.ExileLinkedForEffect(ref, id); err != nil {
			return err
		}
	}
	return nil
})

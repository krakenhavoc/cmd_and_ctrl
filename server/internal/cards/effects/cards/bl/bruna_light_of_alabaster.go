package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bruna, Light of Alabaster — Legendary Creature — Angel {3}{W}{W}{U}, 5/5
// (Secret Lair Drop):
//
//	"Flying, vigilance
//	 Whenever Bruna attacks or blocks, you may attach to it any number of
//	 Auras on the battlefield and you may put onto the battlefield attached to
//	 it any number of Aura cards that could enchant it from your graveyard
//	 and/or hand."
//
// One trigger on both events (Morkrut Necropod's OnAny). As it resolves the
// controller is asked three times, each with "any number" and none being the
// "you may": which Auras on the battlefield, whoever controls them, to attach
// (Heavenly Blademaster's pick, over every Aura Bruna could legally carry, CR
// 701.3b), then which Aura cards in their hand and which in their graveyard
// that could enchant Bruna to put onto the battlefield attached to her (One
// Last Job's entry, as one simultaneous entry per zone, CR 303.4f). A card
// that has left its zone, or a Bruna that is a new object, is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8f213eb-58db-4fab-bd0b-2acfd0a86ac3",
		Name:            "Bruna, Light of Alabaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
				},
				"Bruna, Light of Alabaster — attach any number of Auras to it and put any number of Aura cards from your hand and graveyard onto the battlefield attached to it",
				brunaAttach),
		},
	})
}

func brunaAttach(g *game.Game, item *game.StackItem) error {
	host := item.SourceCardID
	if !effects.OnBattlefield(g, host) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	var onBattlefield []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == host || c.IsAttachedTo(host) || !c.HasSubtype("Aura") {
			continue
		}
		if g.AttachLegalForEffect(c.InstanceID, host) {
			onBattlefield = append(onBattlefield, c.InstanceID)
		}
	}
	if len(onBattlefield) > 0 {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:   item.Controller,
			Source:    host,
			Question:  "Bruna, Light of Alabaster — attach any number of Auras on the battlefield to Bruna",
			Cards:     onBattlefield,
			Min:       0,
			Max:       len(onBattlefield),
			ThenKey:   effects.ArdennAttachPick,
			LiveZone:  game.ZoneBattlefield,
			Carry:     []uuid.UUID{host},
			CarryWord: "permanent",
		})
	}
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil {
		return nil
	}
	for _, leg := range []struct {
		zone     game.ZoneKind
		cards    []game.Card
		question string
		key      string
	}{
		{game.ZoneHand, p.Hand.Cards, "Bruna, Light of Alabaster — put any number of Aura cards from your hand onto the battlefield attached to Bruna", "hand"},
		{game.ZoneGraveyard, p.Graveyard.Cards, "Bruna, Light of Alabaster — put any number of Aura cards from your graveyard onto the battlefield attached to Bruna", "graveyard"},
	} {
		var ids []uuid.UUID
		for _, c := range leg.cards {
			if c.HasSubtype("Aura") && !c.IsToken() && effects.CanBeAttachedTo(g, item.Controller, c, host) {
				ids = append(ids, c.InstanceID)
			}
		}
		if len(ids) == 0 {
			continue
		}
		then := brunaFromHand
		if leg.zone == game.ZoneGraveyard {
			then = brunaFromGraveyard
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  item.Controller,
			Source:   host,
			Question: leg.question,
			Cards:    ids,
			Min:      0,
			Max:      len(ids),
			ThenKey:  then,
			LiveZone: leg.zone,
			Carry:    []uuid.UUID{host},
		})
	}
	return nil
}

// brunaPut puts the picked Aura cards onto the battlefield attached to Bruna,
// the ones still in `from` and still able to enchant her (CR 608.2b).
func brunaPut(g *game.Game, r game.CardsChosen, from game.ZoneKind) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 || !effects.OnBattlefield(g, r.Carry[0]) {
		return nil
	}
	var entries []game.BatchEntry
	for _, id := range r.Picked {
		z := g.FindCardZoneForEffect(id)
		card, ok := g.LookupCardForEffect(id)
		if z == nil || z.Kind != from || !ok || !effects.CanBeAttachedTo(g, r.Chooser, card, r.Carry[0]) {
			continue
		}
		entries = append(entries, game.BatchEntry{CardID: id, From: from})
	}
	if len(entries) == 0 {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries,
		game.ZoneEntryOptions{Controller: r.Chooser, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]}},
		game.KeyedTail{})
}

// The two picks, keyed (on-disk identities: never renamed, never reused).
var brunaFromHand = game.RegisterChooseCardsThen("choose-cards/bruna-light-of-alabaster-hand", func(g *game.Game, r game.CardsChosen) error {
	return brunaPut(g, r, game.ZoneHand)
})

var brunaFromGraveyard = game.RegisterChooseCardsThen("choose-cards/bruna-light-of-alabaster-graveyard", func(g *game.Game, r game.CardsChosen) error {
	return brunaPut(g, r, game.ZoneGraveyard)
})

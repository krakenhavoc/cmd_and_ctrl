package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dawnbreak Reclaimer — Creature — Angel {4}{W}{W}, 5/5 (Secret Lair Drop,
// #5462):
//
//	"Flying
//	 At the beginning of your end step, choose a creature card in an
//	 opponent's graveyard, then that player chooses a creature card in your
//	 graveyard. You may return those cards to the battlefield under their
//	 owners' control."
//
// Three answers in one resolution, each by the player the card names. The
// controller picks a creature card in any opponent's graveyard (none, no
// ability); that card's owner, "that player", then picks a creature card in the
// controller's graveyard (none, and only the first is on offer); and the
// controller is asked once whether to return both ("you may return those
// cards"). The return is one simultaneous entry from the graveyards, each card
// under its OWNER's control (CR 110.2 would otherwise give the controller the
// opponent's creature, which the card does not say), and a card that left its
// graveyard in the meantime is simply not returned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6fa7229d-18b6-43d8-aa57-a56cdb815767",
		Name:            "Dawnbreak Reclaimer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Dawnbreak Reclaimer — choose a creature card in an opponent's graveyard, then that player chooses a creature card in your graveyard; you may return those cards to the battlefield under their owners' control",
				dawnbreakReclaimerAsk),
		},
	})
}

// dawnbreakReclaimerAsk finds the opponents with a creature card in their
// graveyard and, when more than one has, asks the controller which one's
// graveyard to look in: the card is chosen from ONE player's graveyard, so the
// choice is the player and then the card.
func dawnbreakReclaimerAsk(g *game.Game, item *game.StackItem) error {
	var owners []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == item.Controller || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() {
				owners = append(owners, p.ID)
				break
			}
		}
	}
	switch len(owners) {
	case 0:
		return nil
	case 1:
		dawnbreakReclaimerPickCard(g, item, owners[0])
		return nil
	}
	return effects.ChoosePlayer{
		Only:     owners,
		Question: "Dawnbreak Reclaimer — choose an opponent whose graveyard to choose a creature card from",
		ThenKey:  dawnbreakReclaimerPlayer,
	}.Apply(effects.NewContext(g, item))
}

// dawnbreakReclaimerPlayer is the answer to the opponent pick. An on-disk
// key: never renamed, never reused.
var dawnbreakReclaimerPlayer = effects.ChoosePlayerThen("option-pick/dawnbreak-reclaimer-opponent", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen == uuid.Nil {
		return nil
	}
	dawnbreakReclaimerPickCard(ctx.Game, ctx.Item, chosen)
	return nil
})

// dawnbreakReclaimerPickCard offers the controller the creature cards in
// `owner`'s graveyard.
func dawnbreakReclaimerPickCard(g *game.Game, item *game.StackItem, owner uuid.UUID) {
	p := g.PlayerByIDForEffect(owner)
	if p == nil || p.Graveyard == nil {
		return
	}
	var cards []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() {
			cards = append(cards, c.InstanceID)
		}
	}
	if len(cards) == 0 {
		return
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: owner,
		Source:     item.SourceCardID,
		Question:   "Dawnbreak Reclaimer — choose a creature card in " + dawnbreakReclaimerName(g, owner) + "'s graveyard",
		Cards:      cards,
		Min:        1,
		Max:        1,
		ThenKey:    dawnbreakReclaimerFirst,
		LiveZone:   game.ZoneGraveyard,
		Carry:      []uuid.UUID{item.Controller},
		Asked:      game.NewAskedItem(item),
	})
}

// dawnbreakReclaimerFirst hands the second choice to the first card's owner.
// An on-disk key: never renamed, never reused.
var dawnbreakReclaimerFirst = game.RegisterChooseCardsThen("choose-cards/dawnbreak-reclaimer-first", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	first, controller := r.Picked[0], r.Carry[0]
	card, ok := g.LookupCardForEffect(first)
	if !ok {
		return nil
	}
	var mine []uuid.UUID
	if p := g.PlayerByIDForEffect(controller); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() {
				mine = append(mine, c.InstanceID)
			}
		}
	}
	if len(mine) == 0 || dawnbreakReclaimerGone(g, card.Owner) {
		return dawnbreakReclaimerOffer(g, r.Asked, r.Source, controller, first)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    card.Owner,
		FromPlayer: controller,
		Source:     r.Source,
		Question:   "Dawnbreak Reclaimer — choose a creature card in " + dawnbreakReclaimerName(g, controller) + "'s graveyard",
		Cards:      mine,
		Min:        1,
		Max:        1,
		ThenKey:    dawnbreakReclaimerSecond,
		LiveZone:   game.ZoneGraveyard,
		Carry:      []uuid.UUID{controller, first},
		Asked:      game.NewAskedItem(r.Asked),
	})
	return nil
})

// dawnbreakReclaimerSecond asks the controller whether to return both. An
// on-disk key: never renamed, never reused.
var dawnbreakReclaimerSecond = game.RegisterChooseCardsThen("choose-cards/dawnbreak-reclaimer-second", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) < 2 {
		return nil
	}
	controller, first := r.Carry[0], r.Carry[1]
	if !r.Answered || len(r.Picked) == 0 {
		return dawnbreakReclaimerOffer(g, r.Asked, r.Source, controller, first)
	}
	return dawnbreakReclaimerOffer(g, r.Asked, r.Source, controller, first, r.Picked[0])
})

// dawnbreakReclaimerOffer is "you may return those cards".
func dawnbreakReclaimerOffer(g *game.Game, asked *game.StackItem, source, controller uuid.UUID, cards ...uuid.UUID) error {
	ctx, _ := effects.AskedContext(g, asked, controller, source, nil)
	return effects.MayChoice{
		Player:   controller,
		Question: "Dawnbreak Reclaimer — return those creature cards to the battlefield under their owners' control?",
		YesLabel: "Return them",
		NoLabel:  "Don't",
		Then:     dawnbreakReclaimerReturn,
		Carry:    cards,
	}.Apply(ctx)
}

// dawnbreakReclaimerReturn is the "yes": one entry, each card under its
// owner's control. An on-disk key: never renamed, never reused.
var dawnbreakReclaimerReturn = effects.MayYes("confirm/dawnbreak-reclaimer-return", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	// MayChoice puts the asking controller first on the carry.
	ids := r.Carry
	if len(ids) > 0 && ids[0] == r.Chooser {
		ids = ids[1:]
	}
	return effects.PutCardsFromZoneTogether(ctx, game.ZoneGraveyard, ids, uuid.Nil, false)
})

func dawnbreakReclaimerName(g *game.Game, id uuid.UUID) string {
	if p := g.PlayerByIDForEffect(id); p != nil {
		return p.Name
	}
	return "its controller"
}

// dawnbreakReclaimerGone reports that the card's owner can't choose: they have
// left the game.
func dawnbreakReclaimerGone(g *game.Game, owner uuid.UUID) bool {
	p := g.PlayerByIDForEffect(owner)
	return p == nil || p.Eliminated
}

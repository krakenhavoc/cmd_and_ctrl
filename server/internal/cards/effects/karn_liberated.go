package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karn Liberated — Legendary Planeswalker — Karn for {7}, starting
// loyalty 6 (EDHREC rank 7173):
//
//	"+4: Target player exiles a card from their hand.
//	 −3: Exile target permanent.
//	 −14: Restart the game, leaving in exile all non-Aura permanent
//	      cards exiled with Karn. Then put those cards onto the
//	      battlefield under your control."
//
// The −3 is colorless unconditional removal that answers anything,
// including an indestructible or regenerating permanent, because exile
// is not destruction. Routed through the engine's shared exit, so a
// COMMANDER exiled by it gets the CR 903.9 prompt and its owner may
// choose the command zone instead (#529 / #539).
//
// The +4: the TARGET player picks the card from their own hand and it
// is exiled (#3861). It is the pick-from-hand prompt Chrome Mox uses
// (game.ChooseCardsPrompt with Zone: ZoneHand), with the target as the
// chooser and a floor of one, because "exiles a card" is not a "may".
// A player with an empty hand is not asked: there is nothing to exile.
//
// Both exiles are LINKED to this Karn object (CR 607.2a,
// Game.ExileLinkedForEffect), which is what "cards exiled with Karn
// Liberated" reads. A Karn that left the battlefield and came back is a
// new object with no claim on what the old one exiled (CR 400.7), and
// so is another Karn Liberated.
//
// The −14 restarts the game (CR 727, #3856, ADR 0201). The non-Aura
// permanent cards exiled with this Karn are exempt (CR 727.5): they
// stay in exile while every other card goes back to its owner's deck,
// the new game begins with Karn's controller as the starting player
// (CR 727.1a), and once the mulligans are over those cards enter the
// battlefield under that player's control (CR 727.4). An Aura, an
// instant or a sorcery exiled with Karn is not exempt and goes back to
// its owner's deck like everything else. The engine does the restart;
// this card only names the exempt cards.
func init() {
	Register(Spec{
		OracleID:     "0ca233f4-1b7f-4807-ab6e-2b1f5439b3db",
		Name:         "Karn Liberated",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 6,
		Activated: []ActivatedAbility{
			{
				Label:   "+4: Target player exiles a card from their hand.",
				Cost:    LoyaltyCost(4),
				Targets: TargetPlayer("target player"),
				Effect:  karnPlusFour,
			},
			{
				Label:   "−3: Exile target permanent.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetPermanent("target permanent"),
				Effect:  karnMinusThree,
			},
			{
				Label:  "−14: Restart the game, leaving in exile all non-Aura permanent cards exiled with Karn. Then put those cards onto the battlefield under your control.",
				Cost:   LoyaltyCost(-14),
				Effect: karnMinusFourteen,
			},
		},
	})
}

// karnPlusFour asks the target player to exile a card from their hand.
// Its pick is keyed (karnExilePick).
func karnPlusFour(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		p := g.PlayerByIDForEffect(t.ID)
		if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			continue
		}
		cards := make([]uuid.UUID, 0, p.Hand.Size())
		for _, c := range p.Hand.Cards {
			cards = append(cards, c.InstanceID)
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    t.ID,
			FromPlayer: t.ID,
			Source:     item.SourceCardID,
			Question:   "Karn Liberated — exile a card from your hand",
			Cards:      cards,
			Min:        1,
			Max:        1,
			// Re-checked on submit: the pick must still be in that
			// player's hand when the answer arrives.
			ThenKey:  karnExilePick,
			LiveZone: game.ZoneHand,
			Asked:    game.NewAskedItem(item),
		})
	}
	return nil
}

// karnExilePick is the card the target player picked, exiled linked to the
// Karn object that asked, keyed (#4253): the asking item (and the object it
// names) rides the prompt as data.
var karnExilePick = askedPick("choose-cards/karn-liberated-exile-from-hand", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	link, _ := ctx.SourceRef()
	return ctx.Game.ExileLinkedForEffect(link, r.Picked[0])
})

// karnMinusThree exiles the target permanent, linked to this Karn.
func karnMinusThree(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ExileTarget{Target: t.ID, Linked: true}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// karnMinusFourteen restarts the game, exempting the non-Aura permanent
// cards in exile that this Karn object exiled.
func karnMinusFourteen(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var exempt []uuid.UUID
	if link, ok := ctx.SourceRef(); ok {
		for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(link)) {
			if c.IsPermanent() && !c.IsAura() {
				exempt = append(exempt, c.InstanceID)
			}
		}
	}
	g.RestartTheGameForEffect(item.Controller, item.SourceCardID, exempt)
	return nil
}

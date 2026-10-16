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
// The −3 is complete and is what the card does at a four-player
// table: colorless unconditional removal that answers anything,
// including an indestructible or regenerating permanent, because
// exile is not destruction. Routed through the engine's shared exit,
// so a COMMANDER exiled by it gets the CR 903.9 prompt and its owner
// may choose the command zone instead (#529 / #539).
//
// The +4 is complete (#3861): the TARGET player picks the card from
// their own hand and it is exiled. It is the pick-from-hand prompt
// Chrome Mox uses (game.ChooseCardsPrompt with Zone: ZoneHand), with
// the target as the chooser and a floor of one, because "exiles a
// card" is not a "may". A player with an empty hand is not asked:
// there is nothing to exile. Until #3861 it was a discard, which put
// the card in a graveyard it could be recurred from.
//
// THE −14 IS NOT REGISTERED, and this is the one omission in the
// catalog that needs no argument. "Restart the game" is not an
// effect; it is a second game, seeded from the exile pile of the
// first, with new hands, new mulligans and a new turn order. There
// is no approximation of it that is recognisably the same ability,
// and a −14 that did something smaller would be a different card
// wearing Karn's text. The loyalty still accrues past 14.
func init() {
	Register(Spec{
		OracleID:     "0ca233f4-1b7f-4807-ab6e-2b1f5439b3db",
		Name:         "Karn Liberated",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The -14 isn't offered — restarting the game isn't something this engine can do.",
		},
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
				Effect:  exileEveryLegalCardTarget,
			},
		},
	})
}

// karnPlusFour asks the target player to exile a card from their hand.
// Package-level, and its continuation too, so the queued prompt
// captures nothing.
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
			Zone: game.ZoneHand,
			// Chrome Mox's continuation, shared as Semblance Anvil
			// shares it: exile the one card picked.
			Then: chromeMoxExileThePick,
		})
	}
	return nil
}

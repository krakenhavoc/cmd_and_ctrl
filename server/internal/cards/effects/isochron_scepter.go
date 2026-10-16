package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Isochron Scepter — Artifact {2} (EDHREC rank 760):
//
//	"Imprint — When this artifact enters, you may exile an instant card
//	 with mana value 2 or less from your hand.
//	 {2}, {T}: You may copy the exiled card. If you do, you may cast the
//	 copy without paying its mana cost."
//
// The imprint is an ETB trigger (Chrome Mox's shape, #578): it uses the
// stack, and the pick is a choose-cards prompt over the hand with a
// floor of zero for the "may". The exile is LINKED to the Scepter
// object (ExileLinkedForEffect, CR 607.2a), so the activated ability
// reads "the exiled card" off Card.ExiledWith and a second Scepter's
// card, or a card this Scepter imprinted before it left and came back,
// is never the one copied.
//
// The activation copies that card in exile and offers a free cast of
// the copy (CopyAndCast, #3258): free locks X at 0 and still asks for
// additional costs (the 2020-08-07 rulings). Not casting it is the
// "may": the copy ceases to exist. A Scepter that left the battlefield
// with the ability on the stack still copies (the link is to the
// object, read from the item), and a card that has left exile can't
// be copied, so nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ddd649f4-dbaf-48ec-8ff7-95581257772d",
		Name:         "Isochron Scepter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Isochron Scepter — you may exile an instant card with mana value 2 or less from your hand",
				isochronImprint),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: You may copy the exiled card. If you do, you may cast the copy without paying its mana cost.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			// What it answers is whatever the imprinted instant does, so
			// the row declares the union of every answer an instant can
			// give (ADR 0142's ruling for a conditional row).
			Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerSurvive | game.AnswerPump |
				game.AnswerPrevent | game.AnswerRemove | game.AnswerSacOutlet | game.AnswerRestrict |
				game.AnswerCombatGrant | game.AnswerAnimate | game.AnswerMakesBlocker},
			Effect: isochronCopyAndCast,
		}},
	})
}

// isochronImprint asks for the card and exiles it linked to the Scepter.
func isochronImprint(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	ref, ok := g.SourceObjectForEffect(item)
	if !ok {
		return nil
	}
	p := g.PlayerByIDForEffect(controller)
	if p == nil || p.Hand == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.IsInstant() && c.ManaValue() <= 2 {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   source,
		Question: "Isochron Scepter — you may exile an instant card with mana value 2 or less from your hand",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return g.ExileLinkedForEffect(ref, picked[0])
		},
	})
	return nil
}

// isochronCopyAndCast copies the card this Scepter object imprinted and
// offers the free cast.
func isochronCopyAndCast(g *game.Game, item *game.StackItem) error {
	ref, ok := g.SourceObjectForEffect(item)
	if !ok {
		return nil
	}
	var cards []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		cards = append(cards, c.InstanceID)
	}
	if len(cards) == 0 {
		return nil
	}
	return CopyAndCast{Cards: cards, Free: true}.Apply(NewContext(g, item))
}

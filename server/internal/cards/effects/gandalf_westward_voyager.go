package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gandalf, Westward Voyager — Legendary Creature — Avatar Wizard
// {3}{G}{U}, 5/5 (Tales of Middle-earth Commander):
//
//	"Whenever you cast a spell with mana value 5 or greater, each
//	 opponent reveals the top card of their library. If any of those
//	 cards shares a card type with that spell, copy that spell, you may
//	 choose new targets for the copy, and each opponent draws a card.
//	 Otherwise, you draw a card."
//
// Every opponent reveals (an empty library reveals nothing), then the
// revealed cards are compared with the spell's card types (CR 205.2a) as
// the spell last existed (CR 608.2h), so one match among them is enough.
// The copy is Krark's: made from the spell's last known information, with
// new targets offered, and a permanent spell's copy becomes a token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4563ae4b-6226-4cc9-abde-582d0ab83433",
		Name:         "Gandalf, Westward Voyager",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(ManaValueGE(5), "Gandalf, Westward Voyager — each opponent reveals the top card of their library; if any shares a card type with the spell, copy it and each opponent draws a card, otherwise you draw a card",
				s82GapB117GandalfVoyager),
		},
	})
}

func s82GapB117GandalfVoyager(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spellID := ctx.Trigger().Event.CardID
	if spellID == uuid.Nil {
		return nil
	}
	spell, spellOK := g.LookupCardForEffect(spellID)
	shared := false
	for _, opp := range ctx.Opponents() {
		var revealed []uuid.UUID
		if err := (RevealTopOfLibrary{Player: opp, N: 1, Reason: "Gandalf, Westward Voyager", Revealed: &revealed}).Apply(ctx); err != nil {
			return err
		}
		for _, id := range revealed {
			if c, ok := g.LookupCardForEffect(id); ok && spellOK && vowB16SharesATypeWithAny(c, []*game.Card{&spell}) {
				shared = true
			}
		}
	}
	if !shared {
		return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
	}
	if err := (CopySpell{StackID: spellID, Controller: item.Controller, ChooseNewTargets: true, FromLastKnown: true}).Apply(ctx); err != nil {
		return err
	}
	for _, opp := range ctx.Opponents() {
		if err := (DrawCards{Player: opp, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

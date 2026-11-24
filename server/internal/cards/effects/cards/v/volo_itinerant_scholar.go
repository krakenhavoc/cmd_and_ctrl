package cardsv

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volo, Itinerant Scholar — Legendary Creature — Human Wizard {2}{U}, 2/3:
//
//	"When Volo enters, create Volo's Journal, a legendary colorless Book
//	 artifact token with hexproof and "Whenever you cast a creature spell,
//	 note one of its creature types that hasn't been noted for Volo's
//	 Journal."
//	 {2}, {T}: Draw a card for each creature type noted for target
//	 permanent you control named Volo's Journal.
//	 Choose a Background"
//
// The Journal is the printed token (VolosJournalToken); its trigger notes
// on the Journal itself (game.Card.Notes, #6097), so each Journal keeps its
// own list and a Journal that leaves and returns starts over (CR 400.7).
// The activated ability counts the target's noted types as it resolves.
// "Choose a Background" is a deck-building rule (deck/partner.go).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b61f5b1-7688-4336-8fe9-66da37a34df9",
		Name:         "Volo, Itinerant Scholar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Volo — create Volo's Journal", effects.Do(effects.CreateToken{Template: effects.VolosJournalToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Draw a card for each creature type noted for target permanent you control named Volo's Journal.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target permanent you control named Volo's Journal", effects.YouControl(), voloNamedJournal),
			Effect:  voloDrawPerNotedType,
		}},
	})
}

func voloNamedJournal(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return strings.EqualFold(strings.TrimSpace(c.Effective().Name), "Volo's Journal")
}

func voloDrawPerNotedType(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	journal, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := len(g.NotesForEffect(journal))
	if n == 0 {
		return nil
	}
	return g.DrawNForEffect(ctx.Controller(), n)
}

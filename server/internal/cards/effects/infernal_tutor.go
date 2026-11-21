package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Infernal Tutor — Sorcery {1}{B} (Ravnica Remastered):
//
//	"Reveal a card from your hand. Search your library for a card with the
//	 same name as that card, reveal it, put it into your hand, then shuffle.
//	 Hellbent — If you have no cards in hand, instead search your library for
//	 a card, put it into your hand, then shuffle."
//
// Hellbent is read as the spell resolves, and Infernal Tutor is no longer in
// the hand by then, so a spell cast from a one-card hand finds an empty one.
// Otherwise the caster picks the card to reveal (a forced pick of one), it
// is revealed from the hand, and the search is for its name (an empty or
// missing name finds nothing). The search reveals what it finds, as printed;
// the hellbent search names no quality and reveals nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68a203a3-2283-4e29-b177-b3be43f76899",
		Name:         "Infernal Tutor",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g := ctx.Game
			p := g.PlayerByIDForEffect(item.Controller)
			if p == nil {
				return nil
			}
			if p.Hand == nil || p.Hand.Size() == 0 {
				return SearchLibrary{
					Player:  item.Controller,
					Source:  item.SourceCardID,
					Dest:    game.ZoneHand,
					Limit:   1,
					Shuffle: true,
					Reason:  "Infernal Tutor — hellbent: search for a card",
				}.Apply(ctx)
			}
			var hand []uuid.UUID
			for _, c := range p.Hand.Cards {
				hand = append(hand, c.InstanceID)
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  item.Controller,
				Source:   item.SourceCardID,
				Question: "Infernal Tutor — reveal a card from your hand",
				Cards:    hand,
				Min:      1,
				Max:      1,
				LiveZone: game.ZoneHand,
				ThenKey:  infernalTutorReveal,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// infernalTutorReveal is the reveal and the same-name search, keyed. An
// on-disk key: never renamed, never reused.
var infernalTutorReveal = askedPick("choose-cards/infernal-tutor-reveal", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	card, ok := ctx.Game.LookupCardForEffect(r.Picked[0])
	if !ok {
		return nil
	}
	name := card.Name
	ctx.Game.RevealForEffect(game.RevealSpec{
		Player: ctx.Controller(), Source: ctx.Source(),
		Reason: "Infernal Tutor — the revealed card", Cards: r.Picked,
	})
	return SearchLibrary{
		Player:    ctx.Controller(),
		Source:    ctx.Source(),
		Predicate: func(c game.Card) bool { return name != "" && c.Name == name },
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Infernal Tutor — a card named " + name,
	}.Apply(ctx)
})

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vault 21: House Gambit — Enchantment — Saga {1}{R} (Fallout):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Discard a card, then draw a card.
//	 III — Reveal up to five nonland cards from your hand. For each of those
//	 cards that has the same mana value as another card revealed this way,
//	 create a Treasure token."
//
// I and II are The Ruinous Wrecking Crew's rummage: the discard is chosen
// first and the draw follows even from an empty hand. III asks for up to
// five nonland cards from the hand (a keyed choose_cards prompt, 0 to 5),
// reveals them, then groups them by mana value (X counts as zero in a
// hand, CR 202.3e): every card whose mana value another revealed card
// shares makes one Treasure, so three cards of one value make three.
//
// No simplification.
func init() {
	const name = "Vault 21: House Gambit"
	rummage := func(g *game.Game, item *game.StackItem) error {
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player: item.Controller, Source: item.SourceCardID, N: 1,
		}, theRuinousWreckingCrewDraw, nil)
	}
	Register(Spec{
		OracleID:     "8d83d816-518c-4557-a3c4-0255f772c57a",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "discard a card, then draw a card"), rummage),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "discard a card, then draw a card"), rummage),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "reveal up to five nonland cards from your hand; create a Treasure token for each that shares a mana value with another"),
				func(g *game.Game, item *game.StackItem) error {
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Hand == nil {
						return nil
					}
					var cards []uuid.UUID
					for _, c := range p.Hand.Cards {
						if !c.HasCardType("land") {
							cards = append(cards, c.InstanceID)
						}
					}
					if len(cards) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:    item.Controller,
						FromPlayer: item.Controller,
						Source:     item.SourceCardID,
						Question:   name + " — reveal up to five nonland cards from your hand",
						Cards:      cards,
						Min:        0,
						Max:        5,
						LiveZone:   game.ZoneHand,
						ThenKey:    vault21Reveal,
						Asked:      game.NewAskedItem(NewContext(g, item).Item),
					})
					return nil
				}),
		},
	})
}

// vault21Reveal reveals the chosen cards and makes a Treasure for each that
// has the same mana value as another of them. An on-disk key: never
// renamed, never reused.
var vault21Reveal = askedPick("choose-cards/vault-21-house-gambit-reveal", func(ctx *Context, c game.CardsChosen) error {
	if len(c.Picked) == 0 {
		return nil
	}
	g := ctx.Game
	g.RevealForEffect(game.RevealSpec{
		Player: c.Chooser,
		Source: ctx.Source(),
		Reason: "Vault 21: House Gambit — reveal cards from your hand",
		Cards:  c.Picked,
	})
	byValue := map[int]int{}
	for _, id := range c.Picked {
		if card, ok := g.LookupCardForEffect(id); ok {
			if mv, ok := g.ManaValueForEffect(card); ok {
				byValue[mv]++
			}
		}
	}
	n := 0
	for _, count := range byValue {
		if count >= 2 {
			n += count
		}
	}
	if n == 0 {
		return nil
	}
	return CreateToken{Template: TreasureToken(), N: n}.Apply(ctx)
})

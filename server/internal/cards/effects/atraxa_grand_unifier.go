package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Atraxa, Grand Unifier — Legendary Creature — Phyrexian Angel
// {3}{G}{W}{U}{B}, 7/7:
//
//	"Flying, vigilance, deathtouch, lifelink
//	 When Atraxa enters, reveal the top ten cards of your library. For each
//	 card type, you may put a card of that type from among the revealed
//	 cards into your hand. Put the rest on the bottom of your library in a
//	 random order. (Artifact, battle, creature, enchantment, instant, land,
//	 planeswalker, and sorcery are card types.)"
//
// The ten are revealed (they stay on the library, public), then the
// controller is asked once per card type in the printed order, for a card
// of that type still among the revealed cards (a prompt is skipped when no
// such card is left; the pick is optional). A chosen card goes to the hand
// at once and cannot be chosen again, so a card with two types fills one
// slot and leaves the other open (the printed ruling). Any set of picks the
// rule allows, one distinct card per type, is reachable. When the last type
// has been asked, everything still on the library goes to the bottom in a
// random order.
func init() {
	Register(Spec{
		OracleID:        "abbcb153-0763-44c6-964f-b4ff0eb64257",
		Name:            "Atraxa, Grand Unifier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "deathtouch", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Atraxa, Grand Unifier — reveal the top ten cards; for each card type, you may put a card of that type from among them into your hand; put the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 10,
						"Atraxa, Grand Unifier — reveal the top ten cards of your library")
					if len(revealed) == 0 {
						return nil
					}
					return atraxaAsk(ctx, revealed, 0)
				}),
		},
	})
}

// atraxaCardTypes are the eight card types, in printed order, with the
// predicate that finds a card of each.
var atraxaCardTypes = []struct {
	name string
	is   func() CardPredicate
}{
	{"artifact", Artifact}, {"battle", Battle}, {"creature", Creature}, {"enchantment", Enchantment},
	{"instant", Instant}, {"land", Land}, {"planeswalker", Planeswalker}, {"sorcery", Sorcery},
}

// atraxaAsk asks for the card of the first type at or after `from` that
// has a candidate among the revealed cards still on the library, and puts
// the rest on the bottom once no type is left.
func atraxaAsk(ctx *Context, revealed []uuid.UUID, from int) error {
	g, you := ctx.Game, ctx.Controller()
	left := cardsStillInALibrary(g, revealed)
	for i := from; i < len(atraxaCardTypes); i++ {
		is := atraxaCardTypes[i].is()
		var candidates []uuid.UUID
		for _, id := range left {
			if c, ok := g.LookupCardForEffect(id); ok && is(g, you, c) {
				candidates = append(candidates, id)
			}
		}
		if len(candidates) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    you,
			FromPlayer: you,
			Source:     ctx.Source(),
			Question:   "Atraxa, Grand Unifier — you may put a " + atraxaCardTypes[i].name + " card from among the revealed cards into your hand",
			Cards:      candidates,
			Min:        0,
			Max:        1,
			LiveZone:   game.ZoneLibrary,
			ThenKey:    atraxaPicked,
			Carry:      revealed,
			CarryNums:  []int{i},
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	return g.PutOnBottomInRandomOrderForEffect(you, game.ZoneLibrary, left)
}

// atraxaPicked is one type's answer: the card to the hand, then the next
// type. An on-disk key.
var atraxaPicked game.ChooseCardsThen

func init() {
	atraxaPicked = askedPick("choose-cards/atraxa-grand-unifier-card-type", func(ctx *Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
				continue
			}
			if err := ctx.Game.BounceToHandForEffect(id); err != nil {
				return err
			}
		}
		next := 0
		if len(r.Nums) > 0 {
			next = r.Nums[0] + 1
		}
		return atraxaAsk(ctx, r.Carry, next)
	})
}

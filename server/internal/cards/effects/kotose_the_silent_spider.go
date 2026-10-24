package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kotose, the Silent Spider — Legendary Creature — Human Ninja {3}{U}{B},
// 4/4:
//
//	"When Kotose enters, exile target card other than a basic land card
//	 from an opponent's graveyard. Search that player's graveyard, hand,
//	 and library for any number of cards with the same name as that card
//	 and exile them. Then that player shuffles. For as long as you control
//	 Kotose, you may play one of the exiled cards, and you may spend mana
//	 as though it were mana of any color to cast it."
//
// The target is exiled first, and its name, as it was in the graveyard,
// is the name searched for. The search is the extraction search (#3988,
// ADR 0015's 2026-10-16 amendment): Kotose's controller looks through
// that player's graveyard, hand and library in one prompt (seeing their
// whole hand as they do, CR 701.23), takes any number of the cards with
// that name, including none, and they are exiled; then that player's
// library is shuffled whatever was taken, because all three zones were
// searched. If the target has left the graveyard by the time the trigger
// resolves, it does nothing (CR 608.2b).
//
// The cards actually exiled, the target among them, are one permission
// (CR 611.2c locks the set) held by Kotose's controller: a play, so a land
// among them can be played with the turn's land drop, and a budget of ONE
// (CastPermission.CastsLeft) that a cast or a land play spends. Its window
// is "for as long as you control Kotose" (ForAsLongAs + WhileYouControl-
// Source); a Kotose that left before the trigger resolved gives a window
// that never started, and the cards stay exiled (CR 611.2b). Mana is spent
// as though it were mana of any color to cast it (AnyColor).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bfef4daa-b950-46ce-8c08-11f9983741cf",
		Name:         "Kotose, the Silent Spider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Kotose, the Silent Spider — exile a card from an opponent's graveyard and every card with its name", kotoseEnters),
				TargetCardInGraveyard("target card other than a basic land card from an opponent's graveyard",
					Not(YouOwn()), kotoseNotBasicLand())),
		},
	})
}

// kotoseNotBasicLand is "card other than a basic land card".
func kotoseNotBasicLand() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return !IsBasicLand(c) }
}

func kotoseEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	card, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	victim, name := card.Owner, card.Name
	return ExileTarget{
		Target: target,
		Then: func(ctx *Context, exiled bool) error {
			var first []uuid.UUID
			if exiled {
				first = []uuid.UUID{target}
			}
			return SearchLibrary{
				Player:               item.Controller,
				LibraryOwner:         victim,
				Predicate:            func(c game.Card) bool { return name != "" && c.Name == name },
				Dest:                 game.ZoneExile,
				Unbounded:            true,
				AlsoGraveyardAndHand: true,
				Shuffle:              true,
				Reason:               "Kotose, the Silent Spider — search their graveyard, hand and library for cards named " + name,
				// #4253 slice 4: the first card rides the tail.
				ThenTail: kotoseTail(ctx, first),
			}.Apply(ctx)
		},
	}.Apply(ctx)
}

// kotoseTail is the search's continuation: the card exiled first rides it
// as Carry, and the cards the search found are the outcome.
func kotoseTail(ctx *Context, first []uuid.UUID) game.KeyedTail {
	t := ctx.ItemTail(kotoseGrantTail)
	t.Carry = first
	return t
}

// kotoseGrantTail is the permission over every exiled card. An on-disk key.
var kotoseGrantTail = CardTail("kotose/grant-play-permission", func(ctx *Context, o game.TailOutcome) error {
	return kotoseGrant(ctx.Game, ctx.Item, append(append([]uuid.UUID(nil), o.Tail.Carry...), o.IDs...))
})

// kotoseGrant is "for as long as you control Kotose, you may play one of
// the exiled cards": one permission over the cards that are in exile now.
func kotoseGrant(g *game.Game, item *game.StackItem, ids []uuid.UUID) error {
	var cards []game.Card
	for _, id := range ids {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !g.Exile.Contains(id) {
			continue
		}
		cards = append(cards, c)
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     item.Controller,
		Zone:       game.ZoneExile,
		AnyColor:   true,
		CastsLeft:  1,
		Duration:   forAsLongAsYouControlThisWindow(g, item),
		Source:     item.SourceCardID,
		SourceName: "Kotose, the Silent Spider",
		Label:      "Play one of the exiled cards for as long as you control Kotose",
	}, cards)
	return nil
}

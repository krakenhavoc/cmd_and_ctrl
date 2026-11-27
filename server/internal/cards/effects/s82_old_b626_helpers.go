package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b626_helpers.go — Judgment cards, S82 OLD batch 626 (#5611).

// S82OldB626ReturnLinkedExiledCards is the leaves-the-battlefield half of the
// original Wormfang pairs: "When this creature leaves the battlefield, return
// the exiled card to the battlefield under its owner's control" (toHand
// false), or "return the exiled cards to their owner's hand" (toHand true).
//
// The two abilities are linked (CR 607.2a): "the exiled card" is whatever the
// enters trigger exiled with THIS object, read off the leaving object's link
// (Card.ExiledWith) as it last existed (CR 603.10). A creature that left
// before its enters trigger resolved has already had this trigger resolve
// with nothing to return, so what it exiles afterwards stays exiled, as the
// Oracle rulings for the cycle say. A card that has since left exile is not
// pulled back (CR 400.7).
func S82OldB626ReturnLinkedExiledCards(label string, toHand bool) game.TriggeredAbility {
	return WhenThisLeaves(label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ref, ok := ctx.SourceRef()
		if !ok {
			return nil
		}
		var ids []uuid.UUID
		for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
			ids = append(ids, c.InstanceID)
		}
		if len(ids) == 0 {
			return nil
		}
		if toHand {
			return g.BounceCardsToHandKeyedForEffect(ids, game.KeyedTail{})
		}
		return ReturnFromExileTogether{Targets: ids}.Apply(ctx)
	})
}

// S82OldB626ExileALandPair is Wormfang Turtle's and Wormfang Newt's whole
// text, two cards with one body: "When this creature enters, exile a land you
// control. When this creature leaves the battlefield, return the exiled card
// to the battlefield under its owner's control." Both abilities come back as
// registered triggers. The land is chosen as the enters trigger resolves
// (nothing happens with no land) and is exiled linked to this object. Each
// card registers its own pick key: an on-disk identity, never renamed or
// reused.
func S82OldB626ExileALandPair(cardName, pickKey string) []game.TriggeredAbility {
	pick := AskedPick(pickKey, func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 1 || !OnBattlefield(ctx.Game, r.Picked[0]) {
			return nil
		}
		return ExileTarget{Target: r.Picked[0], Linked: true}.Apply(ctx)
	})
	return []game.TriggeredAbility{
		WhenThisEnters(cardName+" — exile a land you control", func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			var lands []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.IsLand() {
					lands = append(lands, c.InstanceID)
				}
			}
			if len(lands) == 0 {
				return nil
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:    ctx.Controller(),
				FromPlayer: ctx.Controller(),
				Source:     item.SourceCardID,
				Question:   cardName + " — exile a land you control",
				Cards:      lands,
				Min:        1,
				Max:        1,
				LiveZone:   game.ZoneBattlefield,
				ThenKey:    pick,
				Asked:      game.NewAskedItem(item),
			})
			return nil
		}),
		S82OldB626ReturnLinkedExiledCards(cardName+" — return the exiled card to the battlefield under its owner's control", false),
	}
}

// S82OldB626YouWinTheGame is "you win the game" as a trigger body: the
// trigger's controller wins (Epic Struggle). Nothing happens for an item with
// no controller.
func S82OldB626YouWinTheGame(g *game.Game, item *game.StackItem) error {
	if item.Controller == uuid.Nil {
		return nil
	}
	return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
}

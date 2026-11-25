package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// targets_chosen_contest.go — #6228: Psychic Battle.
//
//	"Whenever a player chooses one or more targets, each player reveals
//	 the top card of their library. The player who reveals the card with
//	 the greatest mana value may change the target or targets. If two or
//	 more cards are tied for greatest, the target or targets remain
//	 unchanged. Changing targets this way doesn't trigger abilities of
//	 permanents named Psychic Battle."
//
//	Triggered: []game.TriggeredAbility{TopCardContestChangesTargets("Psychic Battle")}
//
// The engine half is game/targets_chosen.go (game.EventTargetsChosen, one
// per announcement: a cast, an activation, a triggered ability put on
// the stack, a copy, a change of targets) and game/retarget_change_all.go
// (CR 115.7a's all-or-none "change the target or targets").

// TopCardContestChangesTargets is the whole of Psychic Battle's trigger,
// for the card named `name`. It watches game.EventTargetsChosen and
// ignores a change of targets made by a card of the same name, which is
// the last sentence (the change is stamped with its source's name,
// Event.RetargetSourceName). On resolution every living player, in turn
// order, reveals the top card of their library (CR 701.20a; an empty
// library reveals nothing). The one player whose card has the greatest
// mana value (CR 202.3) may change the target or targets of the spell or
// ability that chose them (CR 115.7a, game.RetargetChangeAll); a tie for
// greatest, or nobody revealing, changes nothing. A spell or ability
// that has left the stack in the meantime is not changed.
func TopCardContestChangesTargets(name string) game.TriggeredAbility {
	label := name + " — each player reveals the top card of their library"
	return OnAny([]game.EventKind{game.EventTargetsChosen},
		func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.StackItemID != uuid.Nil && ev.RetargetSourceName != name
		},
		label,
		func(g *game.Game, item *game.StackItem) error {
			return topCardContestChangesTargets(g, item, name)
		})
}

func topCardContestChangesTargets(g *game.Game, item *game.StackItem, name string) error {
	ctx := NewContext(g, item)
	target := ctx.Trigger().Event.StackItemID
	best, winner, tied := -1, uuid.Nil, false
	for _, p := range ApnapLivingSeats(g) {
		for _, id := range g.RevealTopOfLibraryForEffect(p, item.SourceCardID, 1, name+" — reveals the top card of their library") {
			c, ok := g.LookupCardForEffect(id)
			if !ok {
				continue
			}
			switch mv := c.ManaValue(); {
			case mv > best:
				best, winner, tied = mv, p, false
			case mv == best:
				tied = true
			}
		}
	}
	if winner == uuid.Nil || tied {
		return nil
	}
	err := g.OfferRetargetForEffect(game.RetargetOffer{
		ItemID:   target,
		Chooser:  winner,
		Policy:   game.RetargetChangeAll,
		Optional: true,
		Source:   item.SourceCardID,
		Reason:   name + " — you may change the target or targets",
	})
	if errors.Is(err, game.ErrCardNotFound) || errors.Is(err, game.ErrInvalidParam) {
		// Gone from the stack, or nothing left to change.
		return nil
	}
	return err
}

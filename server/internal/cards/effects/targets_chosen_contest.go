package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// targets_chosen_contest.go — #6228, Psychic Battle:
//
//	"Whenever a player chooses one or more targets, each player reveals
//	 the top card of their library. The player who reveals the card with
//	 the greatest mana value may change the target or targets. If two or
//	 more cards are tied for greatest, the target or targets remain
//	 unchanged. Changing targets this way doesn't trigger abilities of
//	 permanents named Psychic Battle."
//
// The trigger is game.EventTargetsChosen: one per choosing, at an
// announcement (CR 601.2c, CR 602.2b, CR 603.3d) or a change of targets
// (CR 115.7). The change this effect makes carries the name of the
// object that made it, so a permanent named `name` does not trigger on
// its own change.

// WheneverAPlayerChoosesTargets is "Whenever a player chooses one or
// more targets", except for a change of targets made by an object named
// `exceptChangedBy` ("" excepts nothing).
func WheneverAPlayerChoosesTargets(exceptChangedBy, label string, effect Effect) game.TriggeredAbility {
	return On(game.EventTargetsChosen, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.StackItemID != uuid.Nil && (exceptChangedBy == "" || ev.SourceName != exceptChangedBy)
	}, label, effect)
}

// TopCardContestMayRetarget is the effect: each player still in the game
// reveals the top card of their library, in APNAP order (CR 101.4); the
// one player whose card has the greatest mana value may change the
// target or targets of the spell or ability that triggered it (CR 115.7,
// "change the target(s)": all of them to other legal targets, or none). A
// tie for greatest, or nobody revealing anything, changes nothing. A
// player with an empty library reveals nothing and so has no mana value
// to win with. The spell or ability gone from the stack, or left with
// no target, changes nothing either.
func TopCardContestMayRetarget(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	target := item.Trigger.Event.StackItemID
	if !g.RetargetableForEffect(target) {
		return nil
	}
	reason := item.Label
	best, winner, tied := -1, uuid.Nil, false
	for _, p := range ApnapPlayers(g) {
		revealed := g.RevealTopOfLibraryForEffect(p, item.SourceCardID, 1, reason)
		if len(revealed) == 0 {
			continue
		}
		mv := 0
		if c, ok := g.LookupCardForEffect(revealed[0]); ok {
			mv = c.ManaValue()
		}
		switch {
		case mv > best:
			best, winner, tied = mv, p, false
		case mv == best:
			tied = true
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
		Reason:   "You revealed the greatest mana value — you may change the target or targets",
	})
	if err == game.ErrCardNotFound || err == game.ErrInvalidParam {
		// Gone, or no target left to change: nothing happens.
		return nil
	}
	return err
}

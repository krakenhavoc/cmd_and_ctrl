package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gapB71LookAtHandThenMayCastFree is "look at that player's hand. If you do,
// you may cast a spell from among those cards without paying its mana cost"
// for the player a combat damage trigger names: Silent-Blade Oni's body with
// the card's own name on the offer. The look is the information the player
// is owed once they say yes; the offer lists the spells among those cards they
// could legally cast free, and nothing is asked when there is none
// (CR 608.2c, 118.9).
func GapB71LookAtHandThenMayCastFree(g *game.Game, item *game.StackItem, name string) error {
	if item == nil || item.Trigger == nil {
		return nil
	}
	hand := g.LookAtHandForEffect(item.Controller, item.Trigger.Event.Target)
	perm := g.FreeCastFromHandPermission(item.Controller, item.SourceCardID, name)
	candidates := FreeCastCandidates(g, &perm, hand)
	if len(candidates) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      item.Controller,
		Source:       item.SourceCardID,
		Reason:       name,
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it in their hand",
	})
	return nil
}

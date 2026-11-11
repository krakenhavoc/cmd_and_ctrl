package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silent-Blade Oni — Creature — Demon Ninja {3}{U}{U}{B}{B}, 6/5 (Outlaws
// of Thunder Junction Commander, #5457):
//
//	"Ninjutsu {4}{U}{B}
//	 Whenever this creature deals combat damage to a player, look at that
//	 player's hand. You may cast a spell from among those cards without
//	 paying its mana cost."
//
// The look is a look (CR 701.20e): you alone learn the damaged player's
// hand (game.LookAtHandForEffect), and the cards stay where they are. The
// cast is made while the trigger resolves (CR 608.2g), through the
// resolution_cast prompt over the cards of that hand a spell can be made
// of: one free, any-timing, pass-closed permission over those card
// objects (game.FreeCastFromHandPermission), spent by the first cast,
// which is "a spell" (one). The card is cast out of its owner's hand
// (game: a stored permission is the only thing that reaches a card in
// another player's hand, and only the cast that claims it), you control
// the spell, and it goes to its owner's graveyard as any card does. X is
// 0 and additional costs are still paid (CR 107.3b, CR 118.9). Declining,
// or a hand with no spell to cast, leaves every card in that hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f515250-0f1f-4fb7-aa75-b703b81ccd4a",
		Name:         "Silent-Blade Oni",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Silent-Blade Oni — look at that player's hand; you may cast a spell from among those cards without paying its mana cost",
				silentBladeOniTrigger),
		},
		Activated: []ActivatedAbility{Ninjutsu("{4}{U}{B}")},
	})
}

func silentBladeOniTrigger(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	hand := g.LookAtHandForEffect(item.Controller, item.Trigger.Event.Target)
	perm := g.FreeCastFromHandPermission(item.Controller, item.SourceCardID, "Silent-Blade Oni")
	candidates := freeCastCandidates(g, &perm, hand)
	if len(candidates) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      item.Controller,
		Source:       item.SourceCardID,
		Reason:       "Silent-Blade Oni",
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it in their hand",
	})
	return nil
}

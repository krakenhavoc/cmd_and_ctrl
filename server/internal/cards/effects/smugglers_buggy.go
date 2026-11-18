package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Smuggler's Buggy — Artifact — Vehicle {4}, 5/5 (Commander: Streets of New
// Capenna):
//
//	"Hideaway 4 (When this artifact enters, look at the top four cards of
//	 your library, exile one face down, then put the rest on the bottom in
//	 a random order.)
//	 Whenever this Vehicle deals combat damage to a player, you may cast
//	 the exiled card without paying its mana cost. If you do, return this
//	 Vehicle to its owner's hand.
//	 Crew 2"
//
// The hidden card is the one THIS Buggy hid, captured as the object as the
// trigger is built (CR 607.2a). "Cast" is a real cast offered during the
// resolution (a free {0} permission over the card, the Invasion of Alara
// shape), so a hidden land — which can't be cast — offers nothing and the
// Buggy stays. The Vehicle returns only if the offer was taken and the card
// was cast ("if you do"); declining leaves the card hidden for a later hit.
//
// No simplification.
const smugglersBuggyLabel = "Smuggler's Buggy — you may cast the exiled card without paying its mana cost; if you do, return this Vehicle to its owner's hand"

func init() {
	Register(Spec{
		OracleID:     "e19e1e83-75e5-4cfa-8d50-8ed0b027a202",
		Name:         "Smuggler's Buggy",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Smuggler's Buggy"),
		}},
		Triggered: []game.TriggeredAbility{
			Hideaway("Smuggler's Buggy", 4),
			{
				Watches:   []game.EventKind{game.EventDealDamage},
				AppliesTo: ThisDealtCombatDamageToAPlayer,
				Key:       smugglersBuggyLabel,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, smugglersBuggyLabel)
					item.Params.Object = game.ObjectRef(game.ObjectRefOf(*source))
					return item
				},
				Effect: smugglersBuggyHit,
			},
		},
	})
}

// smugglersBuggyHit offers the free cast over the card this Buggy hid.
func smugglersBuggyHit(g *game.Game, item *game.StackItem) error {
	ref := game.PermissionCardRef(item.Params.Object)
	hidden := g.HiddenCardsForEffect(ref)
	if len(hidden) == 0 {
		return nil
	}
	perm := game.CastPermission{
		Player:      item.Controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      item.SourceCardID,
		SourceName:  "Smuggler's Buggy",
		Label:       "Smuggler's Buggy — cast the exiled card without paying its mana cost",
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      item.Controller,
		Source:       item.SourceCardID,
		Reason:       "Smuggler's Buggy",
		Cards:        hidden[:1],
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it hidden",
		Then:         smugglersBuggyAfterCast,
		Carry:        hidden[:1],
	})
	return nil
}

// smugglersBuggyAfterCast is "if you do, return this Vehicle to its owner's
// hand": Picked is the card cast, none when declined. An on-disk key.
var smugglersBuggyAfterCast = game.RegisterChooseCardsThen("resolution-cast/smugglers-buggy-after-cast",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return nil
		}
		ctx, _ := askedContext(g, r.Asked, r.Chooser, r.Source, nil)
		return BounceToHand{Target: r.Source}.Apply(ctx)
	})

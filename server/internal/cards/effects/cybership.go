package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cybership — Artifact — Vehicle {6}, 8/8:
//
//	"Flying
//	 Whenever this Vehicle deals combat damage to a player, put the top
//	 two cards of that player's library onto the battlefield face down
//	 under your control. They're 2/2 Cyberman artifact creatures.
//	 Crew 4"
//
// Roadmap batch 56 (#463), skipped there on "#745 library to
// battlefield without a search (+#656)". #745 shipped the library put
// and #1270 the other half: CR 708.2's LISTED characteristics. The two
// cards arrive as CybermanBody() — nameless 2/2 Cyberman artifact
// creatures, whatever they are underneath (an instant is fine: a
// face-down entry lifts CR 110.4, as manifest's does).
//
// The put is ONE simultaneous entry through
// PutCardsFromLibraryOntoBattlefieldForEffect with the face-down state
// on the options, so no ETB trigger and no "as enters" hook runs
// (CR 708.2a) and the Cybership's controller is the only player who
// may look at them (CR 708.5) — the library's owner included, which is
// the rule and not a leak. The kind is `turned` (an effect, not a
// keyword): a stolen morph card can be turned up for its own morph
// cost (CR 702.37e); anything else stays a Cyberman for good
// (CR 708.7).
//
// Not revealed on the way: nothing in the text says to, and the cards
// go from one hidden place to a face-down one.
func init() {
	Register(Spec{
		OracleID:        "9c6ac895-0644-47ea-8a40-8f35cbbeba42",
		Name:            "Cybership",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
			},
			Key: "Cybership — put the top two cards of that player's library onto the battlefield face down as Cybermen",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return cybershipTakeTopTwo(g, item.Controller, item.Trigger.Event.Target)
			},
		}},
		Activated: []ActivatedAbility{{
			Label:  "Crew 4",
			Cost:   CrewCost(4),
			Effect: CrewEffect("Cybership"),
		}},
	})
}

// cybershipTakeTopTwo puts the top two cards of victim's library onto
// the battlefield face down under controller's control as Cybermen. A
// short library gives what it has; an empty one gives nothing.
func cybershipTakeTopTwo(g *game.Game, controller, victim uuid.UUID) error {
	p := g.PlayerByIDForEffect(victim)
	if p == nil || p.Library == nil {
		return nil
	}
	cards := p.Library.Cards
	var ids []uuid.UUID
	// The top is the LAST element.
	for i := len(cards) - 1; i >= 0 && len(ids) < 2; i-- {
		if cards[i].IsToken() {
			// CR 111.8: a token in a library is on its way out of the
			// game (CR 704.5d) and cannot come back. It is not one of
			// "the top two cards", and the put would refuse the batch.
			continue
		}
		ids = append(ids, cards[i].InstanceID)
	}
	if len(ids) == 0 {
		return nil
	}
	_, err := g.PutCardsFromLibraryOntoBattlefieldForEffect(ids, game.LibraryEntryOptions{
		Controller:     controller,
		FaceDown:       game.FaceDownTurned,
		FaceDownListed: CybermanBody(),
	})
	return err
}

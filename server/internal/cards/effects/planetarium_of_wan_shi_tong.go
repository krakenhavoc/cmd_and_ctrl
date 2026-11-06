package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Planetarium of Wan Shi Tong — Legendary Artifact {6} (Avatar: The Last
// Airbender, #4459):
//
//	"{1}, {T}: Scry 2.
//	 Whenever you scry or surveil, look at the top card of your library.
//	 You may cast that card without paying its mana cost. Do this only
//	 once each turn. (Look at the card after you scry or surveil.)"
//
// The trigger watches EventScry and EventSurveil, both emitted once the
// scry or surveil is complete, so the card looked at is the one left on
// top afterwards (the reminder text). A scry or surveil of an empty
// library looked at nothing and emits neither.
//
// The cast is made during the resolution (CR 608.2g, the ruling: it can't
// wait for later), through the resolution_cast prompt over that one card
// where it sits in the library (Sunbird's Invocation's shape): a free
// permission naming that object, so the spell reaches the stack from the
// library. "Without paying its mana cost" is the "{0}" price, which locks
// X at 0 (CR 107.3b); additional costs are still paid (the rulings). A
// land is played, never cast (CR 305.9), so a land on top is offered
// nothing. Declined, the card stays where it is.
//
// "Do this only once each turn" limits the CAST (G'raha Tia's and Krile
// Baldesion's reading): the use is recorded only when the card is cast,
// read as the ability would trigger (so it doesn't trigger again that
// turn, the ruling) and again as it resolves (so the second of two
// triggers stacked before either resolved offers nothing). A trigger whose
// cast is declined leaves the turn's use for a later scry.
//
// No simplification.
const (
	planetariumName = "Planetarium of Wan Shi Tong"
	// planetariumDoneLabel is the tally the once-each-turn limit reads.
	planetariumDoneLabel = "Planetarium of Wan Shi Tong — cast the top card (done)"
)

func init() {
	Register(Spec{
		OracleID:     "f89d093d-c815-4df7-a348-2005aa227586",
		Name:         planetariumName,
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Scry 2.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 2},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect:  Do(Scry{N: 2}),
		}},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventScry, game.EventSurveil}, AllOf(ByYou, planetariumNotYetDone),
				"Planetarium of Wan Shi Tong — look at the top card of your library; you may cast it without paying its mana cost",
				planetariumLookAndCast),
		},
	})
}

// planetariumNotYetDone keeps the trigger off the stack once the turn's
// one cast has been made.
func planetariumNotYetDone(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && g.ResolvedThisTurn(source.InstanceID, planetariumDoneLabel) == 0
}

// planetariumLookAndCast looks at the top card and offers its free cast.
func planetariumLookAndCast(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	looked := g.LookAtTopOfLibraryForEffect(controller, 1)
	if len(looked) == 0 || g.ResolvedThisTurn(source, planetariumDoneLabel) > 0 {
		return nil
	}
	top := looked[0]
	perm := game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneLibrary,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInHand,
		Source:      source,
		SourceName:  planetariumName,
		Label:       planetariumName + " — cast it without paying its mana cost",
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       planetariumName,
		Cards:        []uuid.UUID{top},
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it on top",
		Then:         planetariumAnswered,
		Carry:        []uuid.UUID{top},
	})
	return nil
}

// planetariumAnswered records the turn's use once the card was cast.
// An on-disk key: never renamed, never reused.
var planetariumAnswered = game.RegisterChooseCardsThen("choose-cards/planetarium-of-wan-shi-tong-cast",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) > 0 {
			g.NoteDoneThisTurn(r.Source, planetariumDoneLabel)
		}
		return nil
	})

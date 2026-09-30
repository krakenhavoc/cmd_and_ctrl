package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enters_only_if.go — "If this <permanent> would enter, <discard /
// sacrifice> <something> instead. If you do, put it onto the
// battlefield. If you don't, put it into its owner's graveyard."
// (ADR 0098.)
//
// One CR 614.1a replacement per card, with the choice inside it
// (game.EntryCardChoice). Its Replace is the "you don't" branch: it
// rewrites the entry's destination to the owner's graveyard, and the
// engine's entry finisher moves the card there
// (moveRedirectedEntryLocked). So the permanent never enters: it
// triggers nothing that watches an entry and is never tappable in
// response (Mox Diamond's 2008-05-01 ruling), and a later "exile it
// instead" applies to the modified event (CR 616.2).
//
// CR 113.6h makes the ability work however the permanent would enter —
// cast, played, put from a hand or a library, reanimated, returned from
// exile — which the engine gives for free, because every entry opens
// the same window.

// entersOnlyIf builds the shared replacement around a choice.
func entersOnlyIf(name, question string, choice *game.EntryCardChoice) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:         []game.EventKind{game.EventZoneMove},
		SelfReplacement: true,
		Label:           name,
		PromptQuestion:  question,
		EntryCardChoice: choice,
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Controller: EnteringPermanentChooser,
		Replace:    putIntoOwnersGraveyardInstead,
	}
}

// putIntoOwnersGraveyardInstead is "If you don't, put it into its
// owner's graveyard": the entry becomes a move to the owner's
// graveyard (CR 400.3), and anything that applies to THAT event —
// Rest in Peace, CR 903.9 — applies next (CR 616.2).
func putIntoOwnersGraveyardInstead(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
	ev.NewZone = game.ZoneGraveyard
	ev.NewZoneOwner = uuid.Nil
	if src != nil {
		ev.NewZoneOwner = src.Owner
	}
	return nil
}

// EntersOnlyIfYouDiscardFromHand is Mox Diamond's "If this artifact
// would enter, you may discard <clause> instead. If you do, put this
// artifact onto the battlefield. If you don't, put it into its owner's
// graveyard."
//
// `clause` is the card's words ("a land card"), `matches` the same
// clause over a card in hand (printed characteristics). The discard is
// an EFFECT's (DiscardCauseEffect, ADR 0098 owner decision 4), so
// Library of Leng may put the land on top of its owner's library and
// every "whenever you discard" payoff sees it. "If you do" means the
// card really left the hand (owner decision 5).
func EntersOnlyIfYouDiscardFromHand(name, clause string, matches func(game.Card) bool) game.ReplacementEffect {
	return entersOnlyIf(name,
		fmt.Sprintf("%s — discard %s so it enters? If you don't, it goes to the graveyard.", name, clause),
		&game.EntryCardChoice{
			Action:  game.EntryCardDiscard,
			Matches: matches,
			Min:     0,
			Max:     1,
		})
}

// EntersOnlyIfYouSacrifice is the Ice Age / Visions land shape: "If
// this land would enter, sacrifice <n> <clause> instead. If you do, put
// this land onto the battlefield. If you don't, put it into its owner's
// graveyard." (Heart of Yavimaya, Lotus Vale, ADR 0098 Decision 11.)
//
// Not a "may": the choice is which, never whether. With fewer than n
// permanents the clause admits, nothing is sacrificed and the land goes
// to its owner's graveyard (the Lotus Vale ruling). `matches` reads a
// permanent, so an Urborg-made Swamp is a Swamp.
func EntersOnlyIfYouSacrifice(name string, n int, clause string, matches func(game.Card) bool) game.ReplacementEffect {
	return entersOnlyIf(name,
		fmt.Sprintf("%s — sacrifice %s so it enters.", name, clause),
		&game.EntryCardChoice{
			Action:  game.EntryCardSacrifice,
			Matches: matches,
			Min:     n,
			Max:     n,
		})
}

// isLandCard is "a land card" over a card in hand.
func isLandCard(c game.Card) bool { return c.IsLand() }

// untappedLandWithSubtype is "an untapped <Mountain>" over a permanent.
func untappedLandWithSubtype(subtype string) func(game.Card) bool {
	has := IsLandWithSubtype(subtype)
	return func(c game.Card) bool { return !c.Tapped && has(c) }
}

// untappedLand is "an untapped land" over a permanent.
func untappedLand(c game.Card) bool { return !c.Tapped && c.IsLand() }

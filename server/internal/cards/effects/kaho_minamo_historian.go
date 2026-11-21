package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaho, Minamo Historian — Legendary Creature — Human Wizard {2}{U}{U}, 2/2
// (Commander Legends: Battle for Baldur's Gate):
//
//	"When Kaho enters, search your library for up to three instant cards,
//	 exile them, then shuffle.
//	 {X}, {T}: You may cast a spell with mana value X from among cards exiled
//	 with Kaho without paying its mana cost."
//
// The search is Hoarding Dragon's linked exile (game.SearchLibrarySpec.LinkTo,
// CR 607.2a): the cards carry Kaho's object, so a Kaho that left and came
// back finds none of them (CR 400.7). The activated ability is Krang &
// Shredder's free cast over the linked cards, filtered to those whose mana
// value is the X that was paid, granted for one cast during resolution
// (CR 608.2g), at instant timing since they are instants. A card cast this
// way has X = 0 for any X in its cost (CR 107.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "395a7bd6-7c2d-448e-842c-ca53256d7008",
		Name:         "Kaho, Minamo Historian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kaho, Minamo Historian — search your library for up to three instant cards, exile them, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					ref, ok := g.SourceObjectForEffect(item)
					if !ok {
						return nil
					}
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsInstant() },
						Dest:      game.ZoneExile,
						Limit:     3,
						Shuffle:   true,
						LinkTo:    game.PermissionCardRef(ref),
						Source:    item.SourceCardID,
						Reason:    "Kaho, Minamo Historian — search your library for up to three instant cards",
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{X}, {T}: You may cast a spell with mana value X from among cards exiled with Kaho without paying its mana cost.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{X}"), TapCost()),
			Effect:  kahoFreeCast,
		}},
	})
}

const kahoName = "Kaho, Minamo Historian"

// kahoFreeCast offers one free cast over the cards exiled with Kaho whose mana
// value equals X.
func kahoFreeCast(g *game.Game, item *game.StackItem) error {
	ref, ok := g.SourceObjectForEffect(item)
	if !ok {
		return nil
	}
	x := NewContext(g, item).X()
	var cards []game.Card
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.ManaValue() != x {
			continue
		}
		cards = append(cards, *c)
		ids = append(ids, c.InstanceID)
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      item.Controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		CastsLeft:   1,
		LapseOnPass: game.LapseStaysInExile,
		Source:      item.SourceCardID,
		SourceName:  kahoName,
		Label:       kahoName + " — cast a spell with mana value " + strconv.Itoa(x) + " exiled with Kaho without paying its mana cost",
	}, cards)
	offerResolutionCastOver(g, item.Controller, item.SourceCardID, kahoName,
		game.MayCastKeywordFreeCast, declineLeaveInExile, ids...)
	return nil
}

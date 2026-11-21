package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// edict_test.go — "each player sacrifices a creature of their choice"
// (CR 701.21a): Grave Pact, Dictate of Erebos, Fleshbag Marauder,
// Butcher of Malakir.
//
// The assertions that matter are about WHO chooses and WHO is asked,
// because those are the two things a loop-driven implementation gets
// wrong while still reducing everyone's board:
//
//   - each affected player gets their own prompt, so nobody picks for
//     anyone else;
//   - the controller is asked for "each player" and skipped for "each
//     other player";
//   - a player with no creature is skipped rather than left holding a
//     prompt nobody can answer;
//   - and the effect isn't targeted, so hexproof is irrelevant.

const (
	GravePactOracle        = "6f4ac4a4-53ec-4bc9-8f5c-d4b801d867b2"
	DictateOfErebosOracle  = "7c777a41-e40a-4b40-96bf-8ddd5c12924c"
	FleshbagMarauderOracle = "4b1bf05e-753e-4350-a913-894cf3cecc0c"
	ButcherOfMalakirOracle = "a85197ab-dc94-4b72-9716-8dbdbbe90ff8"
)

// isSacrificePrompt reports whether c asks its chooser to sacrifice:
// the single-seat sacrifice_choice kind, or an edict's choose_cards leg
// (#3664, game.ChooseSacrifice).
func IsSacrificePrompt(c *game.PendingChoice) bool {
	if c == nil {
		return false
	}
	return c.Kind == game.PendingChoiceSacrifice ||
		(c.Kind == game.PendingChoiceChooseCards && c.ChooseDestination == game.ChooseSacrifice)
}

// sacrificeOptions is the candidate list of a sacrifice prompt of
// either kind.
func SacrificeOptions(c *game.PendingChoice) []uuid.UUID {
	if c == nil {
		return nil
	}
	if c.Kind == game.PendingChoiceChooseCards {
		return c.ChooseCards
	}
	return c.SacrificeOptions
}

// returnPickFor is the open choose_cards prompt owed by chooser that is
// NOT an edict's sacrifice pick (both are choose_cards since #3664), or
// nil.
func ReturnPickFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser && !IsSacrificePrompt(c) {
			return c
		}
	}
	return nil
}

// edictOffers reports whether id is among opts.
func EdictOffers(opts []uuid.UUID, id uuid.UUID) bool {
	for _, o := range opts {
		if o == id {
			return true
		}
	}
	return false
}

// sacrificeChoiceFor returns the queued sacrifice prompt addressed to a
// player, or nil.
func SacrificeChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if IsSacrificePrompt(c) && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// answerSacrifice answers a player's sacrifice prompt with their
// permanents.
//
// An edict asks one player at a time in APNAP order and collects a
// player with exactly as many candidates as they owe without asking
// (ADR 0175, #3664). So when the player holds no prompt at all, this
// does nothing: the pick was forced (or that player's turn to choose
// has not come), and the test's later assertions on the board are what
// check it.
func AnswerSacrifice(t *testing.T, g *game.Game, chooser uuid.UUID, cardIDs ...uuid.UUID) {
	t.Helper()
	c := SacrificeChoiceFor(g, chooser)
	if c == nil {
		return
	}
	if c.Kind == game.PendingChoiceChooseCards {
		if err := g.ResolveChooseCards(c.ID, chooser, cardIDs); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		return
	}
	if len(cardIDs) != 1 {
		t.Fatalf("a sacrifice_choice takes one permanent, got %d", len(cardIDs))
	}
	if err := g.ResolveSacrificeChoice(c.ID, chooser, cardIDs[0]); err != nil {
		t.Fatalf("ResolveSacrificeChoice: %v", err)
	}
}

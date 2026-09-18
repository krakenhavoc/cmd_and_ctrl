package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Academy Manufactor — Artifact Creature — Assembly-Worker {3}, 1/3
// (EDHREC rank 262):
//
//	"If you would create a Clue, Food, or Treasure token, instead
//	 create one of each."
//
// The artifact deck's engine, and the card that decides the SHAPE of
// the CR 701.7b creation event: it does not change how MANY tokens
// are created, it changes WHICH. That is why the event carries
// groups — a template and a count each — rather than a bare number
// (#762, ADR 0061). A count alone could express Parallel Lives and
// nothing else.
//
// The replacement rewrites only the groups it names, so "create a
// Treasure and two Soldiers" comes out as a Clue, a Food, a Treasure
// and two Soldiers; and it keeps each group's count, so "create two
// Clues" is two of each (ruling, 2021-11-19).
//
// It stacks with a doubler through the CR 616 ordering prompt, and
// the order is observable: with a Doubling Season out, one Clue
// doubled first is two Clues and then two of each kind, six tokens;
// one Clue turned into one of each first is three tokens and then
// doubled, also six — but a Manufactor that fires on a Clue-only
// instruction and a Season that fires afterwards produce different
// KINDS on the way through, which is why the affected player is asked
// rather than the engine deciding.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f36d1d8b-8303-44a9-ab56-531931641ea2",
		Name:         "Academy Manufactor",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCreateTokens &&
					ev.TokenController == src.Controller &&
					ev.TokenTemplatesMatch(isClueFoodOrTreasure)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.ReplaceTokenKindsWhere(isClueFoodOrTreasure, ClueToken(), FoodToken(), TreasureToken())
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Academy Manufactor: a Clue, a Food and a Treasure instead",
		}},
	})
}

// isClueFoodOrTreasure is the Manufactor's clause, read off a token
// template's printed type line — the template is not on the
// battlefield yet, so there is nothing else to read, and no layer
// effect has had a chance to change it.
func isClueFoodOrTreasure(t game.Card) bool {
	return t.HasSubtype("Clue") || t.HasSubtype("Food") || t.HasSubtype("Treasure")
}

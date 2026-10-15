package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coin of Mastery — Artifact {4}:
//
//	"Each creature you control enters with an additional +1/+1 counter
//	 on it for each mana from an artifact source spent to cast it.
//	 {T}: Create a Treasure token."
//
// The headline is a CR 614.1c replacement on the Coin that applies to
// another creature's entry (ADR 0109 §11, #1552). It reads the entering
// spell's payment inside the entry window (game.EntrySpentForEffect),
// and each mana knows whether its source was an artifact from the
// moment it was made (#1212), so a Treasure the caster sacrificed to
// pay still counts. A creature that was not cast entered with nothing
// spent and gets no counter.
//
// The Treasure ability is an ordinary CR 602 activation, and its
// Treasure is the artifact source the first half wants.
//
// An artifact source's mana can be paid on paper — an artifact the
// engine does not run is exactly what Cast anyway is for — so the Coin
// declares ReadsYourCasts with ReadsSourceKinds: a waived creature cast
// under it asks what was paid outside the pool and, per mana, whether
// an artifact made it, and the count is exact (#3855, ADR 0068
// amendment 2026-10-15).
func init() {
	Register(Spec{
		OracleID:       "d78518ee-df79-48d1-b9d5-4f968b441899",
		Name:           "Coin of Mastery",
		ReadsManaSpent: game.ReadsYourCasts | game.ReadsSourceKinds,
		Completeness:   CompletenessFull,
		Replacements: []game.ReplacementEffect{
			CreaturesYouControlEnterWithCountersPerManaFrom(
				"Coin of Mastery: a +1/+1 counter for each mana from an artifact source", game.ManaSourceArtifact, false),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}

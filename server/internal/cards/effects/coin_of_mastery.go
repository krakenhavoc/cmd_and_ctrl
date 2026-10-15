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
// One declared simplification, the paid-cost record's: when a payment
// is waived only the mana that left the pool counts (ADR 0068 §3 and
// its 2026-10-15 amendment). Unlike Treasure mana, an artifact
// source's mana can be paid on paper — an artifact the engine does
// not run is exactly what Cast anyway is for — so the count is a
// lower bound.
func init() {
	Register(Spec{
		OracleID:     "d78518ee-df79-48d1-b9d5-4f968b441899",
		Name:         "Coin of Mastery",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"When “Charge mana costs” is off, or you choose Cast anyway, only mana spent from your mana pool counts, so a creature you cast can get fewer extra +1/+1 counters than the artifact mana you paid.",
		},
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

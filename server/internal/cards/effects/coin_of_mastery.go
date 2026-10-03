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
// One declared simplification, the paid-cost record's: with strict
// mana off the engine never saw what paid, so no counters (ADR 0068
// §3).
func init() {
	Register(Spec{
		OracleID:     "d78518ee-df79-48d1-b9d5-4f968b441899",
		Name:         "Coin of Mastery",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"With strict mana off, the game doesn't track which mana you spent, so your creatures get no extra +1/+1 counters.",
		},
		Replacements: []game.ReplacementEffect{
			CreaturesYouControlEnterWithCountersPerManaFrom(
				"Coin of Mastery: a +1/+1 counter for each mana from an artifact source", game.ManaSourceArtifact, false),
		},
		Activated: []ActivatedAbility{{
			Label: "{T}: Create a Treasure token.",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stormcatch Mentor — Creature — Otter Wizard {U}{R}, 1/1:
//
//	"Haste
//	 Prowess (Whenever you cast a noncreature spell, this creature
//	 gets +1/+1 until end of turn.)
//	 Instant and sorcery spells you cast cost {1} less to cast."
//
// Haste and prowess both ride PrintedKeywords — prowess is a canonical
// keyword the engine turns into its own trigger (game/prowess.go), so
// no trigger is written for it here, exactly as Monastery Mentor's own
// prowess line isn't. The discount is InstantOrSorcerySpell(), Goblin
// Electromancer's predicate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b2252471-01ca-4d58-aa99-4ab0aa5eae12",
		Name:            "Stormcatch Mentor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste", game.KeywordProwess},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				YourSpell(), InstantOrSorcerySpell()),
		},
	})
}

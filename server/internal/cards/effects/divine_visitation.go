package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Divine Visitation — Enchantment {3}{W}{W} (Marvel Super Heroes Commander):
//
//	"If one or more creature tokens would be created under your control,
//	 that many 4/4 white Angel creature tokens with flying and vigilance
//	 are created instead."
//
// Academy Manufactor's shape (hobA9InsteadCreateTokens): a CR 701.7b
// replacement scoped to creations under the controller, rewriting only
// the groups whose template is a creature, so a Treasure made beside a
// Soldier stays a Treasure and the Soldier becomes an Angel. Each group
// keeps its count ("that many").
//
// No simplification.
func init() {
	isCreatureToken := func(t game.Card) bool { return t.IsCreature() }
	Register(Spec{
		OracleID:     "5917216f-5b42-41fa-8976-401bdbeb782e",
		Name:         "Divine Visitation",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			hobA9InsteadCreateTokens("Divine Visitation: 4/4 white Angels with flying and vigilance instead",
				isCreatureToken,
				func() []game.Card { return []game.Card{TokenCard("4/4 white Angel with flying and vigilance")} }),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Triniform — Artifact Creature — Phyrexian Golem {9}, 9/9:
//
//	"When this creature dies, create three 3/3 colorless Phyrexian Golem
//	 artifact creature tokens.
//	 Encore {12}"
//
// Worldspine Wurm's dies trigger and the shared encore keyword (each
// token copy that is sacrificed at the end step has the dies trigger
// too, so it leaves three Golems behind).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f8526a36-eb8f-457d-b532-e12b47622cfa",
		Name:         "Phyrexian Triniform",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Phyrexian Triniform — create three 3/3 Phyrexian Golem artifact creature tokens",
				Do(CreateToken{Template: TokenCard("3/3 colorless Phyrexian Golem artifact"), N: 3})),
		},
		Activated: []ActivatedAbility{Encore("{12}")},
	})
}

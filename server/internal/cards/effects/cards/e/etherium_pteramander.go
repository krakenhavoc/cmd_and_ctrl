package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Etherium Pteramander — Artifact Creature — Salamander Drake {B}, 1/1
// (Modern Horizons 3):
//
//	"Flying
//	 This creature can block only creatures with flying.
//	 {6}{B}: Adapt 4. This ability costs {1} less to activate for each other
//	 artifact you control."
//
// The discount lowers only the generic part, so {B} is always paid.
func init() {
	const label = "{6}{B}: Adapt 4. This ability costs {1} less to activate for each other artifact you control."
	effects.Register(effects.Spec{
		OracleID:        "3e254cae-6e14-4a40-bc02-7ca23750bb26",
		Name:            "Etherium Pteramander",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      effects.NemBlocksOnlyFlyers(effects.OnSelf()),
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{6}{B}"),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.GapB49OtherArtifactsYouControl,
					"This ability costs {1} less to activate for each other artifact you control."),
			},
			Effect: func(g *game.Game, item *game.StackItem) error { return effects.AdaptN(g, item, 4) },
		}},
	})
}

package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icatian Skirmishers — Creature — Human Soldier {3}{W}, 1/1 (Fallen Empires,
// S82 OLD b854, #5637):
//
//	"First strike; banding
//	 Whenever this creature attacks, all creatures banded with it gain first
//	 strike until end of turn."
//
// Banding is the engine's (CR 702.22). The creatures banded with it are the
// other members of its attack band still in combat as the trigger resolves
// (Game.BandMatesForEffect), and the set is fixed then (CR 611.2c). It is none
// when the Skirmishers attacks alone.
//
// No simplification.
func init() {
	const label = "Icatian Skirmishers — all creatures banded with it gain first strike until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "8ddbe82b-7fc9-4b2c-85de-dc747adbb3ca",
		Name:            "Icatian Skirmishers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", game.KeywordBanding},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, id := range g.BandMatesForEffect(item.SourceCardID) {
					if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: label}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}

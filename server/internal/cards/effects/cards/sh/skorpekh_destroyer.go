package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skorpekh Destroyer — Artifact Creature — Necron {2}{B}{B}, 4/2 (Warhammer
// 40,000 Commander, #5948):
//
//	"Deathtouch
//	 Hyperphase Threshers — Whenever an artifact you control enters, this
//	 creature gains first strike until end of turn."
//
// Quicksmith Genius's artifact-enters condition, the Destroyer itself included
// (it is an artifact). The grant is until end of turn on the Destroyer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f96fcb37-5f17-4231-b1a2-64f5fe45730b",
		Name:            "Skorpekh Destroyer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.ArtifactEnteredUnderYourControl, "Skorpekh Destroyer — gains first strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"first strike"},
						Label: "Skorpekh Destroyer — first strike"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

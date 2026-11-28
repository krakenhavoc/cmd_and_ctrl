package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mishra's Groundbreaker — Artifact {4} (Alliances, #5634):
//
//	"{T}, Sacrifice this artifact: Target land becomes a 3/3 artifact creature
//	 that's still a land. (This effect lasts indefinitely.)"
//
// The land becomes an artifact creature with base power and toughness 3/3
// (layers 4 and 7b) in addition to what it was, pinned to that object with no
// end (CR 611.2a, 400.7): it stays a land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84fecd54-c08a-4f5a-b2fa-b254069ccadb",
		Name:         "Mishra's Groundbreaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Target land becomes a 3/3 artifact creature that's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BecomeArtifactCreature{
					Target:       id,
					SetPower:     3,
					SetToughness: 3,
					Label:        "Mishra's Groundbreaker — a 3/3 artifact creature that's still a land",
				}.Apply(ctx)
			},
		}},
	})
}

package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sky Scourer — {1}{B} Creature — Eldrazi Drone 1/2 (Oath of the Gatewatch,
// S82 OLD b296):
//
//	"Devoid (This card has no color.)
//	 Flying
//	 Whenever you cast a colorless spell, this creature gets +1/+0 until end of
//	 turn."
//
// Skittering Cicada's "whenever you cast a colorless spell" with a flat
// pump that goes on the Scourer as the trigger resolves, if it is still the
// same object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d816df03-c79b-4c63-913f-41b38afc173c",
		Name:            "Sky Scourer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Colorless(), "Sky Scourer — this creature gets +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) || g.AbilitySourceIsNewObjectForEffect(item) {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Sky Scourer — +1/+0"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

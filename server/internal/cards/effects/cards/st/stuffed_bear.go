package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stuffed Bear — Artifact {2} (Innistrad: Midnight Hunt):
//
//	"{2}: This artifact becomes a 4/4 green Bear artifact creature until
//	 end of turn."
//
// Levitating Statue's animation with a fixed 4/4 and a colour: one
// continuous effect pinned to this artifact (CR 611.2a, 611.2c) adds the
// Creature type and Bear subtype (layer 4), sets the colour green (layer 5)
// and sets the base power and toughness (layer 7b). It stays an artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1402c116-194e-49ca-a8c1-09ca659c220c",
		Name:         "Stuffed Bear",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This artifact becomes a 4/4 green Bear artifact creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.Cmdr2B13Animate(effects.NewContext(g, item), item.SourceCardID,
					"Stuffed Bear — a 4/4 green Bear artifact creature until end of turn", 4, 4, "Bear", []string{"G"})
			},
		}},
	})
}

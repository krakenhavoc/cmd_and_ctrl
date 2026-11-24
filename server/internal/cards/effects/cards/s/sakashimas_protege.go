package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sakashima's Protege — Creature — Shapeshifter {4}{U}{U}, 3/1 (Commander
// Legends):
//
//	"Flash
//	 Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast it
//	 without paying its mana cost. Put the exiled cards on the bottom in a
//	 random order.)
//	 You may have this creature enter as a copy of any permanent that entered
//	 this turn."
//
// Flash and the engine's cascade trigger, and Clone's entry replacement
// (CR 707.2 + CR 614.1c) over permanents of any type that entered this turn,
// anyone's, and not the Protege itself. Not targeting. Declined, or with
// nothing to copy, it is the 3/1 it prints. A copy of a noncreature permanent
// is that permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cf127842-5fb1-44df-8d91-8cc0f48c559a",
		Name:            "Sakashima's Protege",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{effects.Cascade()},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Sakashima's Protege", func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
				return effects.CopyCandidates(g, self, func(c game.Card) bool {
					return g.EnteredThisTurn(c.InstanceID)
				})
			}, nil),
		},
	})
}

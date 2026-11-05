package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dominion Saboteur — Creature — Shapeshifter {4}{U}, 3/3 (Star Trek,
// #4973):
//
//	"You may have this creature enter as a copy of any artifact or
//	 creature on the battlefield, except it isn't legendary. If you do, it
//	 enters with additional counters on it equal to the same number and
//	 kinds of counters the copied permanent has on it."
//
// Phyrexian Metamorph's candidates (any artifact or creature, anyone's)
// with Spark Double's "isn't legendary" (CR 707.9b). The counters are not
// copiable values: they ride the entry event (AddCounterAtETB, CR 614.1c),
// read off the copied permanent as the copy is chosen, so they are there
// before any enters trigger looks. A loyalty count copied onto a
// planeswalker goes through its starting loyalty instead, as Spark Double's
// extra loyalty does, so the walker still gets its printed loyalty too.
// Declining the copy is a plain 3/3 with no counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af03149b-9307-443c-aefa-69bda52607ba",
		Name:         "Dominion Saboteur",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Dominion Saboteur",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool {
						return c.IsCreature() || c.IsArtifact()
					})
				},
				dominionSaboteurExcept,
			),
		},
	})
}

// dominionSaboteurExcept removes Legendary and copies the copied
// permanent's counters onto the entry.
func dominionSaboteurExcept(ev *game.ReplacementEvent, v *game.PrintedValues, g *game.Game, _ *game.Card) {
	v.RemoveSupertype("Legendary")
	c, ok := g.LookupCardForEffect(ev.CopySourceForEffect())
	if !ok {
		return
	}
	for kind, n := range c.Counters {
		if n <= 0 {
			continue
		}
		if kind == game.CounterLoyalty && v.HasCardType("Planeswalker") {
			v.StartingLoyalty += n
			continue
		}
		ev.AddCounterAtETB(kind, n)
	}
}

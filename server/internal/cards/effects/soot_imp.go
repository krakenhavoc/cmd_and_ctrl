package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soot Imp — Creature — Imp {1}{B}{B}, 1/2:
//
//	"Flying
//	 Whenever a player casts a nonblack spell, that player loses 1 life."
//
// Flying rides PrintedKeywords. The trigger is castersLoseLife
// (caster_loses_life.go) over every player, Soot Imp's controller
// included. "Nonblack" is the spell's colour on the stack, so a
// colourless spell and a gold spell with no black in it both count, and
// a black-and-red one does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c867e31a-90ea-4d20-80db-3977ec68b16b",
		Name:            "Soot Imp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			castersLoseLife("Soot Imp — that player loses 1 life", 1,
				func(spell game.Card) bool { return !spell.HasColor("B") }),
		},
	})
}

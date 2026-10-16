package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Incendiary Oracle — Creature — Human Shaman {1}{R}, 2/2:
//
//	"{1}{R}: This creature gets +1/+0 until end of turn.
//	 If a creature dealt damage by this creature this turn would die,
//	 exile it instead."
//
// Firebreathing (Furnace Gremlin's row) and Kumano, Master Yamabushi's
// replacement (#1946, ExileIfDealtDamageByThisWouldDie).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c3ed6b9-1a2b-42ad-baa9-ee2a0ad505a7",
		Name:         "Incendiary Oracle",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Incendiary Oracle — +1/+0"),
		}},
		Replacements: []game.ReplacementEffect{
			ExileIfDealtDamageByThisWouldDie("Incendiary Oracle: exile it instead of dying"),
		},
	})
}

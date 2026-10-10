package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clergy of the Holy Nimbus — Creature — Human Cleric {W}, 1/1:
//
//	"If this creature would be destroyed, regenerate it.
//	 {1}: This creature can't be regenerated this turn. Only your
//	 opponents may activate this ability."
//
// The regeneration is ADR 0108 PR 1's static (a replacement of every
// destruction, nothing spent), and the ability's body is
// CantBeRegeneratedThisTurn on itself, which that static asks. The
// activator restriction is ADR 0106 §1's 2026-10-07 amendment (#1947):
// OpponentsOnly, so the Clergy's own controller cannot pay {1} to take
// away its protection.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "66566999-f70a-4f14-9bf0-23325295a977",
		Name:         "Clergy of the Holy Nimbus",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{RegenerateIfThisWouldBeDestroyed()},
		Activated:    []ActivatedAbility{noRegenerationRow("{1}")},
	})
}

// noRegenerationRow is the Holy Nimbus pair's "{N}: This creature can't
// be regenerated this turn. Only your opponents may activate this
// ability."
func noRegenerationRow(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label:         cost + ": This creature can't be regenerated this turn. Only your opponents may activate this ability.",
		Cost:          ManaCost(cost),
		OpponentsOnly: true,
		// ADR 0142 sweep rulings: turning off a protection is restrict.
		Purpose: game.Purpose{Answers: game.AnswerRestrict},
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			return CantBeRegeneratedThisTurn{Target: ctx.Source()}.Apply(ctx)
		},
	}
}

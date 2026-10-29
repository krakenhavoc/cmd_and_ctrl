package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dáin of the Ancient Halls — Legendary Creature — Dwarf Noble
// {3}{R}{W}, 4/5:
//
//	"Vigilance, haste
//	 Whenever Dáin attacks, he deals damage equal to the number of
//	 Dwarves you control to each opponent."
//
// The count is read as the trigger resolves (CR 608.2h), and Dáin counts
// himself while he is still a Dwarf on the battlefield; a Dáin that has
// left deals the damage as he last existed, so the count is of the other
// Dwarves then. "Dwarf" is a word of his text, read through it.
//
// No simplification. The Hobbit (#4454).
func init() {
	Register(Spec{
		OracleID:        "84702796-389a-4d7b-804b-de6ebd80a883",
		Name:            "Dáin of the Ancient Halls",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Dáin of the Ancient Halls — damage equal to the number of Dwarves you control to each opponent",
				Do(Instruction(func(ctx *Context) error {
					n := b26PermanentsOfSubtypeControlled(ctx.Game, ctx.Controller(), ctx.TextSubtype("Dwarf"))
					if n <= 0 {
						return nil
					}
					return DamageEachOpponent{N: n}.Apply(ctx)
				}))),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Combustion Technique — Instant — Lesson {1}{R}:
//
//	"Combustion Technique deals damage equal to 2 plus the number of
//	 Lesson cards in your graveyard to target creature. If that creature
//	 would die this turn, exile it instead."
//
// The Lesson cards are counted as the spell resolves (CR 608.2h), while
// Combustion Technique itself is still on the stack. "That creature" is
// the spell's replacement, not the damage's (ADR 0108 §1): the target is
// marked whether or not the damage is dealt.
//
// No simplifications.
func init() {
	lessons := g2CardsInYourGraveyard(g2IsLessonCard)
	Register(Spec{
		OracleID:     "522a04ff-cbfe-47b0-bd29-ef6fc27a6905",
		Name:         "Combustion Technique",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: damageFirstTargetExileIfItDies(func(item *game.StackItem, ctx *Context) int {
			return 2 + lessons(item, ctx)
		}),
	})
}

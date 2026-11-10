package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elenda, Saint of Dusk — Legendary Creature — Vampire Knight {2}{W}{B},
// 4/4 (Foundations, #4465):
//
//	"Lifelink, hexproof from instants
//	 As long as your life total is greater than your starting life
//	 total, Elenda gets +1/+1 and has menace. Elenda gets an additional
//	 +5/+5 as long as your life total is at least 10 greater than your
//	 starting life total."
//
// Hexproof from instants is the shared hexproof-from quality (CR 702.11e),
// card type "instants". The three life-gated effects are Righteous
// Valkyrie's shape (life_gated_statics.go): each declares
// DependsOnLifeTotal so the layers answer the life change at once, and the
// starting total is the table's, not the constant (ADR 0075). The +5/+5 is
// its own layer 7c effect, "an additional" bonus on top of the first.
//
// No simplification.
func init() {
	above := selfWhileYourLifeAboveStarting(1)
	tenAbove := selfWhileYourLifeAboveStarting(10)
	Register(Spec{
		OracleID:        "bfd614d5-4c7f-416b-9cb2-bb167c0893a0",
		Name:            "Elenda, Saint of Dusk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink", "hexproof from instants"},
		Static: []game.StaticAbility{
			LifeGatedPump(above, 1, 1),
			LifeGatedKeyword(above, "menace"),
			LifeGatedPump(tenAbove, 5, 5),
		},
	})
}

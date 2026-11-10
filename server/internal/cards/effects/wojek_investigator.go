package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wojek Investigator — Creature — Angel Detective {2}{W}, 2/4 (Murders at
// Karlov Manor, #4469):
//
//	"Flying, vigilance
//	 At the beginning of your upkeep, investigate once for each opponent
//	 who has more cards in hand than you."
//
// The opponents are counted as the ability resolves (CR 608.2c); each
// gets one Clue's worth, made together as one creation of N Clues.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ab41243d-b178-4ebe-a7ac-bea37427be99",
		Name:            "Wojek Investigator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Wojek Investigator — investigate once for each opponent with more cards in hand",
				Do(Instruction(func(ctx *Context) error {
					me := ctx.PlayerByID(ctx.Controller())
					if me == nil {
						return nil
					}
					n := 0
					for _, id := range ctx.Opponents() {
						if p := ctx.PlayerByID(id); p != nil && p.Hand.Size() > me.Hand.Size() {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return CreateToken{Template: ClueToken(), N: n}.Apply(ctx)
				}))),
		},
	})
}

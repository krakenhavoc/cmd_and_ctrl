package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wojek Investigator — Creature — Angel Detective {2}{W}, 2/4 (Foundations
// Commander):
//
//	"Flying, vigilance
//	 At the beginning of your upkeep, investigate once for each opponent
//	 who has more cards in hand than you."
//
// Counted as the trigger resolves: each opponent still in the game whose
// hand holds more cards than the controller's. One Clue (CR 111.10f) per
// such opponent, none for none.
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
				Do(CreateTokensForEach{
					Template: ClueToken(),
					Name:     "opponents with more cards in hand than you",
					Count: func(ctx *Context) int {
						me := ctx.Game.PlayerByIDForEffect(ctx.Controller())
						if me == nil {
							return 0
						}
						n := 0
						for _, id := range ctx.Opponents() {
							if p := ctx.Game.PlayerByIDForEffect(id); p != nil && !p.Eliminated && p.Hand.Size() > me.Hand.Size() {
								n++
							}
						}
						return n
					},
				})),
		},
	})
}

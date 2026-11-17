package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Blast-Runner — Creature — Goblin {R}, 1/2 (The Brothers' War, S82):
//
//	"This creature gets +2/+0 and has menace as long as you sacrificed a
//	 permanent this turn."
//
// Two statics (a layer 7c pump and a layer 6 keyword) that read the turn
// tally of permanents its controller sacrificed (Count Nefaria's read), so
// a token, a Treasure or a creature all count and an opponent's sacrifice
// does not. Both are on only while the tally is above zero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b69da01-3ba5-4a61-9e26-e879b6d8b63e",
		Name:         "Goblin Blast-Runner",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Gets +2/+0 as long as you sacrificed a permanent this turn.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: whileYouHave(goblinBlastRunnerSacrificed, selfCreature),
				Apply:     pumpBy(2, 0),
			},
			func() game.StaticAbility {
				ab := KeywordGrant(whileYouHave(goblinBlastRunnerSacrificed, selfCreature), "menace")
				ab.Label = "Has menace as long as you sacrificed a permanent this turn."
				return ab
			}(),
		},
	})
}

// goblinBlastRunnerSacrificed is "you sacrificed a permanent this turn".
func goblinBlastRunnerSacrificed(g *game.Game, you uuid.UUID) bool {
	return g.TurnTallyFor(you).PermanentsSacrificed > 0
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mausoleum Guard — Creature — Human Scout {3}{W}, 2/2 (Innistrad
// Remastered):
//
//	"When this creature dies, create two 1/1 white Spirit creature tokens
//	 with flying."
func init() {
	Register(Spec{
		OracleID:     "64f3af46-f8ab-4ba4-9dff-7412cefb4eea",
		Name:         "Mausoleum Guard",
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Mausoleum Guard — create two 1/1 white Spirit creature tokens with flying",
				Do(CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 2})),
		},
	})
}

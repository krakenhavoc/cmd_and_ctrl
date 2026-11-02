package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Professional Wrestler — Creature — Human Warrior Performer {3}{G}, 4/4
// (Marvel's Spider-Man):
//
//	"When this creature enters, create a Treasure token.
//	 This creature can't be blocked by more than one creature."
//
// The block limit is MaxBlockers on itself (Outland Colossus); the
// Treasure is the shared TreasureToken template.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3dd92eea-02ca-4e55-8f69-8c8059dd7ee6",
		Name:         "Professional Wrestler",
		Completeness: CompletenessFull,
		BlockRules:   []game.BlockRule{MaxBlockers(OnSelf(), 1)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Professional Wrestler — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanguine Evangelist — Creature — Vampire Cleric {2}{W}, 2/1 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Battle cry (Whenever this creature attacks, each other attacking
//	 creature gets +1/+0 until end of turn.)
//	 When this creature enters or dies, create a 1/1 black Bat creature
//	 token with flying."
//
// Battle cry is CR 702.91a's trigger row (BattleCry, battle_cry.go). The
// token trigger is one ability with two events (WhenThisEntersOrDies,
// Mogg War Marshal's shape); dying looks back in time (CR 603.10a), so it
// triggers from the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a70d9645-7c25-4fcb-9d16-fb73ce8b55e2",
		Name:         "Sanguine Evangelist",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			BattleCry("Sanguine Evangelist"),
			WhenThisEntersOrDies("Sanguine Evangelist — create a 1/1 black Bat with flying",
				Do(CreateToken{Template: TokenCard("1/1 black Bat with flying"), N: 1})),
		},
	})
}

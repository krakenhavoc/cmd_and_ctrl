package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Order of Midnight // Alter Fate — an adventure card:
//
//	Order of Midnight {1}{B}, Creature — Human Knight 2/2:
//	"Flying
//	 This creature can't block."
//
//	Alter Fate {1}{B}, Sorcery — Adventure:
//	"Return target creature card from your graveyard to your hand."
//
// The adventure layout is Brazen Borrower's. Can't block is Carrion
// Feeder's RestrictSelf; Alter Fate is Raise Dead's return.
//
// No simplification.
func init() {
	const oracleID = "2db17cfc-0947-407f-bd53-d64a48eec314"

	Register(Spec{
		OracleID:        oracleID,
		Name:            "Order of Midnight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})

	Register(Spec{
		OracleID:     oracleID + "#1",
		Name:         "Alter Fate",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}

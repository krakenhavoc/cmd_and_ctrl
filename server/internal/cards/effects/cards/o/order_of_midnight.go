package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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

	effects.Register(effects.Spec{
		OracleID:        oracleID,
		Name:            "Order of Midnight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})

	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Alter Fate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve:    effects.ReturnTargetFromGraveyardToHand,
	})
}

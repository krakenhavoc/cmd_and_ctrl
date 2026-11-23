package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Young Red Dragon // Bathe in Gold — an adventure card (Commander Legends: Battle for
// Baldur's Gate):
//
//	Young Red Dragon {3}{R}, Creature — Dragon 3/2:
//	"Flying
//	 This creature can't block."
//
//	Bathe in Gold {1}{R}, Instant — Adventure:
//	"Create a Treasure token. (Then exile this card. You may cast the creature later from
//	 exile.)"
//
// The adventure layout is Brazen Borrower's (CR 715.3d). Flying is printed and "can't
// block" is Blood Hypnotist's restriction on itself. The Adventure half makes one Treasure.
//
// No simplification.
func init() {
	const oracleID = "2f865ae9-9328-4f6f-924f-d5e0a65aaa96"

	effects.Register(effects.Spec{
		OracleID:        oracleID,
		Name:            "Young Red Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})

	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Bathe in Gold",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}

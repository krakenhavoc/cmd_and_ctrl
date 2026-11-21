package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lovestruck Beast // Heart's Desire — an adventure card (Commander Legends:
// Battle for Baldur's Gate):
//
//	Lovestruck Beast {2}{G}, Creature — Beast Noble 5/5:
//	"This creature can't attack unless you control a 1/1 creature."
//
//	Heart's Desire {G}, Sorcery — Adventure:
//	"Create a 1/1 white Human creature token. (Then exile this card. You may
//	 cast the creature later from exile.)"
//
// The adventure layout is Brazen Borrower's: the creature half under the bare
// oracle ID, the Adventure half under "<oracle_id>#1", and the engine exiles the
// resolved Adventure with permission to cast the creature (CR 715.3d). The attack
// restriction is a fact about the attacker's controller, read live at declaration
// (CR 508.1c) through the new AttackTargetRestriction.ControllerMustControl: a
// permanent that is exactly 1/1 after layers and counters, any creature of any
// type (Lovestruck Beast itself is a 5/5, so it never satisfies its own clause
// unless something shrinks it to 1/1).
//
// No simplification.
func init() {
	const oracleID = "7e84aff5-2cb6-4214-befd-3d2de31229e5"

	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Lovestruck Beast",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.CantAttackUnlessYouControl(game.PermanentQuery{Types: []string{"creature"}}.WithExactPT(1, 1)),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Heart's Desire",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 1}.Apply(ctx)
		},
	})
}

package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pegasus Guardian // Rescue the Foal — an adventure card (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	Pegasus Guardian {5}{W}, Creature — Pegasus 3/3:
//	"Flying
//	 At the beginning of your end step, if a permanent you controlled left the
//	 battlefield this turn, create a 1/1 white Pegasus creature token with
//	 flying."
//
//	Rescue the Foal {1}{W}, Instant — Adventure:
//	"Exile target creature you control, then return that card to the
//	 battlefield under its owner's control."
//
// The adventure layout is Brazen Borrower's (the creature under the bare
// oracle ID, the Adventure half under "<oracle_id>#1"). The end-step trigger
// is Solemn Recruit's revolt trigger; the Adventure is Displace's blink for
// one creature, returned under its owner's control as a new object.
//
// No simplification.
func init() {
	const oracleID = "e71863a7-0de1-4ab5-95e8-c39e6810d899"

	effects.Register(effects.Spec{
		OracleID:        oracleID,
		Name:            "Pegasus Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfRevolt("Pegasus Guardian — create a 1/1 white Pegasus creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Pegasus with flying"), N: 1})),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Rescue the Foal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.FlickerFirstLegalTargetUnderOwner(ctx.Game, item)
		},
	})
}

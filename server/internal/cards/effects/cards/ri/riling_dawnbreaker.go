package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riling Dawnbreaker // Signaling Roar — Tarkir: Dragonstorm (#4463), an
// omen card (CR 720, #5175):
//
//	Riling Dawnbreaker — Creature — Dragon {4}{W}, 3/4
//	  "Flying, vigilance
//	   At the beginning of combat on your turn, another target creature
//	   you control gets +1/+0 until end of turn."
//	Signaling Roar — Sorcery — Omen {1}{W}
//	  "Create a 2/2 white Soldier creature token. (Then shuffle this card
//	   into its owner's library.)"
//
// The Dragon under the bare oracle ID, the Omen under "<oracle>#1"; the
// shuffle is the engine's (CR 720.3d, game/omen.go). The combat trigger
// is War Machine's shape: "another" is object identity, and with no
// other creature to target the trigger is removed (CR 603.3d).
//
// No simplification.

const RilingDawnbreakerOracleID = "1fc31651-70fe-4a58-9b7e-7359510bba51"

func init() {
	const pump = "Riling Dawnbreaker — another target creature you control gets +1/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        RilingDawnbreakerOracleID,
		Name:            "Riling Dawnbreaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat(pump, effects.BoostTargetUntilEOT(pump, 1, 0)),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     RilingDawnbreakerOracleID + "#1",
		Name:         "Signaling Roar",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("2/2 white Soldier"), N: 1}.Apply(ctx)
		},
	})
}

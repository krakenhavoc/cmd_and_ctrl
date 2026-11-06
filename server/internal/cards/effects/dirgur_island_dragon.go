package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dirgur Island Dragon // Skimming Strike — Tarkir: Dragonstorm (#4463),
// an omen card (CR 720, #5175):
//
//	Dirgur Island Dragon — Creature — Dragon {5}{U}, 4/4
//	  "Flying
//	   Ward {2}"
//	Skimming Strike — Instant — Omen {1}{U}
//	  "Tap up to one target creature. Draw a card. (Then shuffle this
//	   card into its owner's library.)"
//
// The Dragon under the bare oracle ID, the Omen under "<oracle>#1"; the
// shuffle is the engine's (CR 720.3d, game/omen.go).
//
// No simplification.

const dirgurIslandDragonOracleID = "23cda76b-6e47-4922-a38d-93f7aae1eb00"

func init() {
	Register(Spec{
		OracleID:        dirgurIslandDragonOracleID,
		Name:            "Dirgur Island Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Dirgur Island Dragon — ward {2}"),
		},
	})
	Register(Spec{
		OracleID:     dirgurIslandDragonOracleID + "#1",
		Name:         "Skimming Strike",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("up to one target creature").WithCount(0, 1),
		OnResolve:    effectThenDrawACard(tapEachLegalTarget),
	})
}

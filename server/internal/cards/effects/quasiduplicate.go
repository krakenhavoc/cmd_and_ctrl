package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quasiduplicate — Sorcery {1}{U}{U} (Guilds of Ravnica, #5525):
//
//	"Create a token that's a copy of target creature you control.
//	 Jump-start"
//
// Cackling Counterpart's body, with jump-start (cmdr2S39JumpStart, CR
// 702.133a) in place of flashback.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "76b709ae-5765-4b7a-99c9-89eade7c8cb8",
		Name:             "Quasiduplicate",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{cmdr2S39JumpStart("{1}{U}{U}")},
		Targets:          TargetCreature("target creature you control", YouControl()),
		OnResolve:        TokenCopyOfSingleTarget,
	})
}

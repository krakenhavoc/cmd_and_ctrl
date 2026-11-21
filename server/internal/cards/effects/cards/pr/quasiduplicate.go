package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:         "76b709ae-5765-4b7a-99c9-89eade7c8cb8",
		Name:             "Quasiduplicate",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Cmdr2S39JumpStart("{1}{U}{U}")},
		Targets:          effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve:        effects.TokenCopyOfSingleTarget,
	})
}

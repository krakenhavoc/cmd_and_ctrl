package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fury Storm — Instant {2}{R}{R} (Commander 2018, #5526):
//
//	"When you cast this spell, copy it for each time you've cast your
//	 commander from the command zone this game. You may choose new targets
//	 for the copies.
//	 Copy target instant or sorcery spell. You may choose new targets for
//	 the copy."
//
// Echo Storm's cast trigger (echoStormCopies, with the new-targets offer)
// over Reverberate's body: each Fury Storm, original or copy, copies its own
// target instant or sorcery spell under your control (CR 707.10), and you
// may choose new targets for that copy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "371fa9e3-5432-4f2f-89d4-55061b0b4e57",
		Name:         "Fury Storm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetInstantOrSorcerySpell("target instant or sorcery spell"),
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell("Fury Storm — copy it for each time you've cast your commander from the command zone", effects.EchoStormCopies),
		},
		OnResolve: effects.CmdrB21CopyTheTargetSpell,
	})
}

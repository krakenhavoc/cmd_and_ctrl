package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Velvetwing Butterflies // Gaze in Wonder — an Adventure card (CR 715,
// The Hobbit, #4454), oracle 70a702fe…:
//
//	Velvetwing Butterflies — Creature — Insect {2}{W}, 2/2
//	  "Flying"
//	Gaze in Wonder — Instant — Adventure {1}{W}
//	  "Tap one or two target creatures. (Then exile this card. You may
//	   cast the creature later from exile.)"
//
// No simplification.

// velvetwingButterfliesOracleID is shared by both faces of the card.
const VelvetwingButterfliesOracleID = "70a702fe-1dd6-4d7b-a85a-45744ecad741"

func init() {
	effects.Register(effects.Spec{
		OracleID:        VelvetwingButterfliesOracleID,
		Name:            "Velvetwing Butterflies",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	effects.Register(effects.Spec{
		OracleID:     VelvetwingButterfliesOracleID + "#1",
		Name:         "Gaze in Wonder",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("one or two target creatures").WithCount(1, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TapEachLegalTarget(ctx.Game, item)
		},
	})
}

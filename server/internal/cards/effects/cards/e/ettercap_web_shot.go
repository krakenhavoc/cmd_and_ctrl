package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ettercapOracleID = "e0b5b3af-af89-4b91-8b55-947ec095d6f2"

// Ettercap // Web Shot — Adventure.
//
//	Ettercap {4}{G} Creature — Spider Beast 2/5: "Reach"
//	Web Shot {2}{G} Instant — Adventure: "Destroy target creature with flying."
func init() {
	effects.Register(effects.Spec{
		OracleID:        ettercapOracleID,
		Name:            "Ettercap",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
	effects.Register(effects.Spec{
		OracleID:     ettercapOracleID + "#1",
		Name:         "Web Shot",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}

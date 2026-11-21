package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sever the Bloodline — Sorcery {3}{B} (Innistrad):
//
//	"Exile target creature and all other creatures with the same name as
//	 that creature.
//	 Flashback {5}{B}{B} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Maelstrom Pulse's one-target-plus-name sweep, limited to creatures and
// exiling. The name is read as the spell resolves; a target that is gone or
// illegal by then fizzles the whole spell (CR 608.2b). Flashback is the
// shared alternative cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "b9019332-2f02-48d9-be0a-a2b41b79e136",
		Name:             "Sever the Bloodline",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		Targets:          effects.TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{5}{B}{B}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.RprB23ExileChosenCreatureAndOthersWithItsName(ctx)
		},
	})
}

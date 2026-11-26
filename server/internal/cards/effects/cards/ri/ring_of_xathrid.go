package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ring of Xathrid — Artifact — Equipment {2} (Magic 2013, S82 OLD, #5568):
//
//	"{2}: Regenerate equipped creature.
//	 At the beginning of your upkeep, put a +1/+1 counter on equipped creature
//	 if it's black.
//	 Equip {1}"
//
// Ring of Thune's upkeep counter, with a regeneration shield put on the creature
// equipped as the ability resolves (CR 701.19a). Nothing happens if the Ring is
// unattached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ceddad0d-94ad-459a-b109-f979ffc26283",
		Name:         "Ring of Xathrid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB084UpkeepCounterIfEquippedIs("Ring of Xathrid — put a +1/+1 counter on equipped creature if it's black", "B"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: Regenerate equipped creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{2}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					host, ok := effects.Cmdr2B21EquippedCreature(g, item.SourceCardID)
					if !ok {
						return nil
					}
					return effects.Regenerate{Target: host.InstanceID}.Apply(effects.NewContext(g, item))
				},
			},
			effects.EquipAbility("{1}"),
		},
	})
}

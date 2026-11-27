package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ring of Evos Isle — Artifact — Equipment {2} (Magic 2013, S82 OLD, #5568):
//
//	"{2}: Equipped creature gains hexproof until end of turn.
//	 At the beginning of your upkeep, put a +1/+1 counter on equipped creature
//	 if it's blue.
//	 Equip {1}"
//
// Ring of Thune's upkeep counter, read on the creature as the trigger resolves
// (CR 603.4 does not apply: the "if" is in the effect). The hexproof is granted
// as the ability resolves to the creature then equipped (CR 611.2c) and is gone
// at end of turn. Nothing happens if the Ring is unattached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4548f379-2647-4446-9e56-e7ef0662793a",
		Name:         "Ring of Evos Isle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB084UpkeepCounterIfEquippedIs("Ring of Evos Isle — put a +1/+1 counter on equipped creature if it's blue", "U"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: Equipped creature gains hexproof until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    effects.ManaCost("{2}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					host, ok := effects.Cmdr2B21EquippedCreature(g, item.SourceCardID)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: host.InstanceID, Keywords: []string{"hexproof"},
						Label: "Ring of Evos Isle — hexproof until end of turn"}.Apply(effects.NewContext(g, item))
				},
			},
			effects.EquipAbility("{1}"),
		},
	})
}

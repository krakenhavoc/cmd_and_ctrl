package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disruptive Stormbrood // Petty Revenge — Tarkir: Dragonstorm (#4463),
// an omen card (CR 720, #5175):
//
//	Disruptive Stormbrood — Creature — Dragon {4}{G}, 3/3
//	  "Flying
//	   When this creature enters, destroy up to one target artifact or
//	   enchantment."
//	Petty Revenge — Sorcery — Omen {1}{B}
//	  "Destroy target creature with power 3 or less. (Then shuffle this
//	   card into its owner's library.)"
//
// The Dragon under the bare oracle ID, the Omen under "<oracle>#1"; the
// shuffle is the engine's (CR 720.3d, game/omen.go). The power is read
// as Petty Revenge is cast and again as it resolves (CR 601.2c,
// CR 608.2b), so a creature pumped above 3 in response is not destroyed.
//
// No simplification.

const DisruptiveStormbroodOracleID = "ec74ae5d-1284-443c-9842-18954f8cf5a8"

func init() {
	effects.Register(effects.Spec{
		OracleID:        DisruptiveStormbroodOracleID,
		Name:            "Disruptive Stormbrood",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Disruptive Stormbrood — destroy up to one target artifact or enchantment", func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyEachLegalTarget(item, effects.NewContext(g, item))
			}),
				effects.TargetPermanent("up to one target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())).WithCount(0, 1)),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     DisruptiveStormbroodOracleID + "#1",
		Name:         "Petty Revenge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with power 3 or less", effects.PowerLE(3)),
		OnResolve:    effects.DestroyEachLegalTarget,
	})
}

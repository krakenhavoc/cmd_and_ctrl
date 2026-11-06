package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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

const disruptiveStormbroodOracleID = "ec74ae5d-1284-443c-9842-18954f8cf5a8"

func init() {
	Register(Spec{
		OracleID:        disruptiveStormbroodOracleID,
		Name:            "Disruptive Stormbrood",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Disruptive Stormbrood — destroy up to one target artifact or enchantment", func(g *game.Game, item *game.StackItem) error {
				return destroyEachLegalTarget(item, NewContext(g, item))
			}),
				TargetPermanent("up to one target artifact or enchantment", Or(Artifact(), Enchantment())).WithCount(0, 1)),
		},
	})
	Register(Spec{
		OracleID:     disruptiveStormbroodOracleID + "#1",
		Name:         "Petty Revenge",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with power 3 or less", PowerLE(3)),
		OnResolve:    destroyEachLegalTarget,
	})
}

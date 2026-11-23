package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twinmaw Stormbrood // Charring Bite — Tarkir: Dragonstorm (#4463), an
// omen card (CR 720, #5175):
//
//	Twinmaw Stormbrood — Creature — Dragon {5}{W}, 5/4
//	  "Flying
//	   When this creature enters, you gain 5 life."
//	Charring Bite — Sorcery — Omen {1}{R}
//	  "Charring Bite deals 5 damage to target creature without flying.
//	   (Then shuffle this card into its owner's library.)"
//
// The Dragon under the bare oracle ID, the Omen under "<oracle>#1"; the
// shuffle is the engine's (CR 720.3d, game/omen.go). The halves are
// different colours: the card is the white Dragon in every zone but the
// stack (CR 720.4), and the Omen on the stack is red (CR 720.3b).
//
// No simplification.

const TwinmawStormbroodOracleID = "ff812a78-3fe7-47bc-9984-1c2380cff5b1"

func init() {
	effects.Register(effects.Spec{
		OracleID:        TwinmawStormbroodOracleID,
		Name:            "Twinmaw Stormbrood",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Twinmaw Stormbrood — you gain 5 life", effects.Do(effects.GainLife{Amount: 5})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     TwinmawStormbroodOracleID + "#1",
		Name:         "Charring Bite",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		Targets:      effects.TargetCreature("target creature without flying", effects.WithoutKeyword("flying")),
		OnResolve:    effects.DamageToFirstTargetStep(5),
	})
}

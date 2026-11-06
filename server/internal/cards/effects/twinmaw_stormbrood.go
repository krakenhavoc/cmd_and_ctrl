package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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

const twinmawStormbroodOracleID = "ff812a78-3fe7-47bc-9984-1c2380cff5b1"

func init() {
	Register(Spec{
		OracleID:        twinmawStormbroodOracleID,
		Name:            "Twinmaw Stormbrood",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Twinmaw Stormbrood — you gain 5 life", Do(GainLife{Amount: 5})),
		},
	})
	Register(Spec{
		OracleID:     twinmawStormbroodOracleID + "#1",
		Name:         "Charring Bite",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		Targets:      TargetCreature("target creature without flying", WithoutKeyword("flying")),
		OnResolve:    damageToFirstTarget(5),
	})
}

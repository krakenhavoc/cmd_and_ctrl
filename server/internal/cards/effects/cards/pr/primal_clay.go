package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The three forms of a Primal Clay, as the anchor words its "as this creature
// enters" prompt offers (ADR 0071): the printed text.
const (
	primalClayThree = "3/3 artifact creature"
	primalClayFlyer = "2/2 artifact creature with flying"
	primalClayWall  = "1/6 Wall artifact creature with defender"
)

// Primal Clay — Artifact Creature — Shapeshifter {4}, */* (Magic 2013, S82 OLD, #5568):
//
//	"As this creature enters, it becomes your choice of a 3/3 artifact
//	 creature, a 2/2 artifact creature with flying, or a 1/6 Wall artifact
//	 creature with defender in addition to its other types."
//
// Corrupted Shapeshifter's as-enters choice: the form's lines are gated on the
// word, a layer 7b base power and toughness and a layer 6 keyword on the
// creature itself, with the Wall subtype added in layer 4 for the third form
// ("in addition to its other types": it stays a Shapeshifter and an artifact
// creature). There is never a moment on the battlefield with the printed */*
// (CR 614.12); counters and anthems still apply to the base.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dccf2ca8-8c87-41c9-8373-351859396d05",
		Name:         "Primal Clay",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseOptionAsEnters("Primal Clay", primalClayThree, primalClayFlyer, primalClayWall),
		Static: []game.StaticAbility{
			effects.StaticWhenChosen(primalClayThree, effects.GapB53SelfBasePT(3, 3, "This creature is a 3/3 artifact creature.")),
			effects.StaticWhenChosen(primalClayFlyer, effects.GapB53SelfBasePT(2, 2, "This creature is a 2/2 artifact creature with flying.")),
			effects.StaticWhenChosen(primalClayFlyer, effects.GapB53SelfKeyword("flying")),
			effects.StaticWhenChosen(primalClayWall, effects.GapB53SelfBasePT(1, 6, "This creature is a 1/6 Wall artifact creature with defender.")),
			effects.StaticWhenChosen(primalClayWall, effects.S82OldB429SelfAlsoSubtype("This creature is a Wall in addition to its other types.", "Wall")),
			effects.StaticWhenChosen(primalClayWall, effects.GapB53SelfKeyword("defender")),
		},
	})
}

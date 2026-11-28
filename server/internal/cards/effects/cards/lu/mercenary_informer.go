package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mercenary Informer — Creature — Human Rebel Mercenary {2}{W}, 2/1
// (Prophecy):
//
//	"This creature can't be the target of black spells or abilities from
//	 black sources.
//	 {2}{W}: Put target nontoken Mercenary on the bottom of its owner's
//	 library."
//
// Raiding Party's targeting bar (CantBeTargetedByColor: black spells and
// abilities of black sources, read as they are, CR 702.16b's shape without
// the other protection halves) and a tuck onto the bottom of the library
// (Bant Charm's, through PutIntoLibrary). The target is any nontoken
// Mercenary, any controller's, this one included; a token never is one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "245e26ec-39eb-4af8-99af-8cc22edbc261",
		Name:         "Mercenary Informer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.CantBeTargetedByColor("B", "This creature can't be the target of black spells or abilities from black sources."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Put target nontoken Mercenary on the bottom of its owner's library.",
			Cost:    effects.ManaCost("{2}{W}"),
			Targets: effects.TargetPermanent("target nontoken Mercenary", effects.HasSubtype("Mercenary"), effects.Not(effects.IsTokenPredicate())),
			Effect:  effects.OldB668TuckTheTargetOnTheBottom,
		}},
	})
}

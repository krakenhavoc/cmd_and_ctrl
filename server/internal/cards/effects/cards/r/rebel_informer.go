package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rebel Informer — Creature — Human Mercenary Rebel {2}{B}, 1/2 (Prophecy):
//
//	"This creature can't be the target of white spells or abilities from
//	 white sources.
//	 {3}: Put target nontoken Rebel on the bottom of its owner's library."
//
// Mercenary Informer's shape with the colours and the creature type
// swapped: Raiding Party's targeting bar against white, and the tuck of any
// nontoken Rebel, any controller's, this one included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85ef10fd-4a48-4a37-baad-f663fb5439e0",
		Name:         "Rebel Informer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.CantBeTargetedByColor("W", "This creature can't be the target of white spells or abilities from white sources."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Put target nontoken Rebel on the bottom of its owner's library.",
			Cost:    effects.ManaCost("{3}"),
			Targets: effects.TargetPermanent("target nontoken Rebel", effects.HasSubtype("Rebel"), effects.Not(effects.IsTokenPredicate())),
			Effect:  effects.OldB668TuckTheTargetOnTheBottom,
		}},
	})
}

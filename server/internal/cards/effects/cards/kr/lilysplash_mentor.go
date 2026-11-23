package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lilysplash Mentor — Creature — Frog Druid {2}{G}{U}, 4/4 (Bloomburrow,
// #4467):
//
//	"Reach
//	 {1}{G}{U}: Exile another target creature you control, then return it
//	 to the battlefield under its owner's control with a +1/+1 counter on
//	 it. Activate only as a sorcery."
//
// Spaceshift's flicker (Flicker.PlusOneCounters) as a sorcery-speed
// activated ability, with Baboon Spirit's "another target creature you
// control". The counter is part of the entry, so it is there before an
// enters trigger looks (CR 614.1c); the creature is a new object (CR
// 400.7), and a token exiled this way ceases to exist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "20194297-f2a4-4473-95bc-7c6835e5b6b3",
		Name:            "Lilysplash Mentor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{G}{U}: Exile another target creature you control, then return it to the battlefield under its owner's control with a +1/+1 counter on it. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{1}{G}{U}"),
			Targets:      effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SocB22FlickerFirstTargetWithACounter(item, effects.NewContext(g, item))
			},
		}},
	})
}

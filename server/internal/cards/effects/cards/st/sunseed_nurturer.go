package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunseed Nurturer — Creature — Human Druid Wizard {2}{W}, 1/1 (Shards of
// Alara, S82 OLD batch 517, #5588):
//
//	"At the beginning of your end step, if you control a creature with power
//	 5 or greater, you may gain 2 life.
//	 {T}: Add {C}."
//
// An intervening "if" (CR 603.4) read as the end step begins and again as the
// trigger resolves, with the "may" asked after the recheck; the power is read
// through the layers. The mana ability is the plain colourless tap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "3518e273-bd92-46bc-b361-da52c670dfe1",
		Name:          "Sunseed Nurturer",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.TapForColorless()},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.S82OldB517CondYouControlPower5{},
				effects.Optional(effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, effects.S82OldB517YouControlPower5),
					"Sunseed Nurturer — you may gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
					"Sunseed Nurturer — gain 2 life?")),
		},
	})
}

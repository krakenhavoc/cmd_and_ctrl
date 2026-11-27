package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rimewind Cryomancer — Creature — Human Wizard {3}{U}, 2/3 (Coldsnap, #5597):
//
//	"{1}, {T}: Counter target activated ability. Activate only if you control
//	 four or more snow permanents. (Mana abilities can't be targeted.)"
//
// Azorius Guildmage's counter of an activated ability on the stack (a spell, a
// triggered ability and a mana ability, which never uses the stack, are not
// targets) with Heidar's snow-count activation condition.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23c881f8-37c7-4899-890e-9f6e26f9f560",
		Name:         "Rimewind Cryomancer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}, {T}: Counter target activated ability. Activate only if you control four or more snow permanents.",
			Purpose:   game.Purpose{Answers: game.AnswerRemove},
			Cost:      effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Condition: effects.S82OldB567SnowPermanentsAtLeast(4),
			Targets:   effects.TargetSpellOrAbility("target activated ability", effects.AnActivatedAbilityItem()),
			Effect:    effects.CounterTheChosenSpell,
		}},
	})
}

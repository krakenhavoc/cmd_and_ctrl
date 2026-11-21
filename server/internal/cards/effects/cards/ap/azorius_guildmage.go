package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Azorius Guildmage — Creature — Vedalken Wizard {W/U}{W/U}, 2/2 (Ravnica
// Remastered, #5948):
//
//	"{2}{W}: Tap target creature.
//	 {2}{U}: Counter target activated ability. (Mana abilities can't be
//	 targeted.)"
//
// The tap is the ordinary targeted tap. The counter targets an activated
// ability on the stack (Reroute's AnActivatedAbilityItem): a spell, a triggered
// ability and a mana ability, which never uses the stack, are not legal targets.
// Countering an ability removes it from the stack without resolving it
// (CR 701.5a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16f61f69-cdde-4b6c-b855-c74db9038ceb",
		Name:         "Azorius Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{W}: Tap target creature.",
				Cost:    effects.ManaCost("{2}{W}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TapChosenPermanent,
			},
			{
				Label:   "{2}{U}: Counter target activated ability.",
				Cost:    effects.ManaCost("{2}{U}"),
				Targets: effects.TargetSpellOrAbility("target activated ability", effects.AnActivatedAbilityItem()),
				Effect:  effects.CounterTheChosenSpell,
			},
		},
	})
}

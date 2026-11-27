package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blind Seer (Invasion, S82 OLD batch 647, #5616):
//
//	Legendary Creature — Human Wizard {2}{U}{U}, 3/3:
//	"{1}{U}: Target spell or permanent becomes the color of your choice until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f905b72d-734a-44e8-b118-c8de889994a9",
		Name:         "Blind Seer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Target spell or permanent becomes the color of your choice until end of turn.",
			Cost:    effects.ManaCost("{1}{U}"),
			Targets: effects.TargetSpellOrPermanent("target spell or permanent", nil, nil),
			Effect:  effects.S82OldB647TargetSpellOrPermanentBecomesChosenColor,
		}},
	})
}

package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nurturing Licid — Creature — Licid {1}{G}, 1/1 (Tempest, S82 OLD b753,
// #5628):
//
//	"{G}, {T}: This creature loses this ability and becomes an Aura
//	 enchantment with enchant creature. Attach it to target creature. You
//	 may pay {G} to end this effect.
//	 {G}: Regenerate enchanted creature."
//
// The licid shape (LicidAbility, LicidForm, ReturnedAura). The second row
// is Blessing of Leeches' regeneration of the host; as a creature it is
// attached to nothing and does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "537a7af1-2b7e-4828-b645-7e34e86058d5",
		Name:         "Nurturing Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static:       effects.LicidForm(0),
		Activated: []effects.ActivatedAbility{
			effects.LicidAbility("{G}", "{G}"),
			{
				Label:   "{G}: Regenerate enchanted creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{G}"),
				Effect:  effects.RegenerateAttachedHost,
			},
		},
	})
}

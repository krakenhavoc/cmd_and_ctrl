package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harsh Judgment — Enchantment {2}{W}{W}:
//
//	"As this enchantment enters, choose a color.
//	 If an instant or sorcery spell of the chosen color would deal damage
//	 to you, it deals that damage to its controller instead."
//
// ADR 0108 §9 decision 4 (#1905): a static redirection (CR 614.9) of an
// instant or sorcery spell's damage to you, judged as the spell is when it
// would deal the damage (CR 609.7c), to its controller. Before a colour is
// chosen, no spell is "of the chosen color".
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1fa272bf-8759-4750-b40b-8e2f6972d570",
		Name:         "Harsh Judgment",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForProtection, "Harsh Judgment"),
		Replacements: []game.ReplacementEffect{
			effects.StaticRedirection("Harsh Judgment — that spell deals the damage to its controller instead",
				effects.RedirectWhere{Applies: effects.DamageToYouFromAnInstantOrSorceryOfTheChosenColor, To: effects.ToTheSourcesController}),
		},
	})
}

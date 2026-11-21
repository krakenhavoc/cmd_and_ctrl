package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandwurm Convergence — Enchantment {6}{G}{G}:
//
//	"Creatures with flying can't attack you or planeswalkers you control.
//	 At the beginning of your end step, create a 5/5 green Wurm creature token."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "429eff2b-635b-43f6-b20f-82e25992c0b9",
		Name:         "Sandwurm Convergence",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.Cmdr2B03CreaturesMatchingCantAttackYouOrPlaneswalkers(
			"Creatures with flying can't attack you or planeswalkers you control.",
			game.PermanentQuery{Types: []string{"creature"}, Keyword: "flying"})},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Sandwurm Convergence — create a 5/5 green Wurm creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 green Wurm"), N: 1})),
		},
	})
}

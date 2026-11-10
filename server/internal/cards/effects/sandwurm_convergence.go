package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandwurm Convergence — Enchantment {6}{G}{G}:
//
//	"Creatures with flying can't attack you or planeswalkers you control.
//	 At the beginning of your end step, create a 5/5 green Wurm creature token."
func init() {
	Register(Spec{
		OracleID:     "429eff2b-635b-43f6-b20f-82e25992c0b9",
		Name:         "Sandwurm Convergence",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{cmdr2B03CreaturesMatchingCantAttackYouOrPlaneswalkers(
			"Creatures with flying can't attack you or planeswalkers you control.",
			game.PermanentQuery{Types: []string{"creature"}, Keyword: "flying"})},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Sandwurm Convergence — create a 5/5 green Wurm creature token",
				Do(CreateToken{Template: TokenCard("5/5 green Wurm"), N: 1})),
		},
	})
}

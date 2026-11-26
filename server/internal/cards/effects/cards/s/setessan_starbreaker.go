package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Setessan Starbreaker — Creature — Human Warrior {3}{G}, 2/1 (Born of the
// Gods, S82 OLD b379, #5559):
//
//	"When this creature enters, you may destroy target Aura."
//
// War Priest of Thune's optional destroy, over an Aura. The target is chosen
// as the trigger goes on the stack and the "may" is asked as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc84a4e7-5f5f-4105-b0f8-c06ebcde6218",
		Name:         "Setessan Starbreaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Setessan Starbreaker — destroy target Aura", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target Aura", effects.HasSubtype("Aura"))),
				"Setessan Starbreaker — destroy target Aura?"),
		},
	})
}

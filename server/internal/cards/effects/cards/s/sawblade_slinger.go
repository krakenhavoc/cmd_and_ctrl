package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sawblade Slinger — Creature — Human Archer {3}{G}, 4/3 (Innistrad:
// Crimson Vow, #5495):
//
//	"When this creature enters, choose up to one —
//	 • Destroy target artifact an opponent controls.
//	 • This creature fights target Zombie an opponent controls."
//
// A modal enters trigger whose mode is chosen as it goes on the stack
// (CR 603.3c); "up to one" lets the controller choose neither (Ravager
// Wurm's template). The Zombie is read through the layers. The fight is
// the shared b10Fight: nothing is dealt if either creature has left
// (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df7397b6-9a37-4012-8677-ae8e3b7c4bc3",
		Name:         "Sawblade Slinger",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{sawbladeSlingerTrigger()},
	})
}

func sawbladeSlingerTrigger() game.TriggeredAbility {
	t := effects.WhenThisEnters("Sawblade Slinger — choose up to one", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseN("Choose up to one", 0, 1,
		effects.ModeDoing("Destroy target artifact an opponent controls.",
			effects.Harmful(effects.TargetPermanent("target artifact an opponent controls", effects.Artifact(), effects.OpponentControls())),
			effects.DestroyTheModesTarget),
		effects.ModeDoing("This creature fights target Zombie an opponent controls.",
			effects.Harmful(effects.TargetCreature("target Zombie an opponent controls", effects.OpponentControls(), effects.Subtype("Zombie"))),
			effects.ThisFightsTheModesTarget),
	)
	return t
}

package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pelia, Immortal Innovator — Legendary Creature — Lanthanite Artificer
// {3}{U}, 3/3 (Star Trek Commander, #4974):
//
//	"When Pelia enters, return target creature an opponent controls to
//	 its owner's hand."
//
// Stingscourger's enters trigger: the target is chosen as the trigger
// goes on the stack and re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c0be0d0-766b-4c50-b1ca-cce527196666",
		Name:         "Pelia, Immortal Innovator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Pelia, Immortal Innovator — return target creature an opponent controls to its owner's hand",
				effects.BounceTheTargetToHand), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}

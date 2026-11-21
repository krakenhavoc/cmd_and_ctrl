package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arbor Colossus — 6/6 Giant for {2}{G}{G}{G}:
//
//	"Reach
//	 {3}{G}{G}{G}: Monstrosity 3. (If this creature isn't monstrous,
//	 put three +1/+1 counters on it and it becomes monstrous.)
//	 When this creature becomes monstrous, destroy target creature
//	 with flying an opponent controls."
//
// A mandatory targeted trigger: with no flier across the table it is
// removed as it would go on the stack (CR 603.3d) and the Giant is
// simply monstrous.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bec585af-0c59-46a7-838f-d747c80a9e97",
		Name:            "Arbor Colossus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{3}{G}{G}{G}"), 3)},
		Triggered:       []game.TriggeredAbility{arborColossusTrigger()},
	})
}

func arborColossusTrigger() game.TriggeredAbility {
	t := effects.WhenBecomesMonstrous("Arbor Colossus — destroy target creature with flying", effects.DestroyChosenPermanent)
	t.Targets = effects.TargetCreature("target creature with flying an opponent controls", effects.HasKeyword("flying"), effects.OpponentControls())
	return t
}

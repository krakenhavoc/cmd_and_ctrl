package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foundation Breaker — Creature — Elemental {3}{G}, 2/2 (EDHREC rank
// 2927):
//
//	"When this creature enters, you may destroy target artifact or
//	 enchantment.
//	 Evoke {1}{G} (You may cast this spell for its evoke cost. If you
//	 do, it's sacrificed when it enters.)"
//
// Naturalize on a body, or Naturalize for two with evoke. The ETB is
// an optional targeted trigger (the "you may" is the trigger prompt,
// the artifact-or-enchantment clause the pick_target prompt), and
// the destroy is the single-target verb, so indestructible is
// honoured. Evoke is the S22 alternative cost: evoked, the Breaker
// enters, the destroy trigger and the evoke sacrifice trigger both
// go on the stack, and the destroy happens whichever order their
// controller stacks them — because the creature genuinely entered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b7a3613-f1bd-4262-9b11-b631167c2c2d",
		Name:         "Foundation Breaker",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Evoke("{1}{G}"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      b06SelfETB,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Foundation Breaker — destroy target artifact or enchantment?"},
			Targets:        TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Key:            "Foundation Breaker — destroy target artifact or enchantment",
			Effect:         destroyChosenPermanent,
		}},
	})
}

package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tyrant of Valakut — Creature — Dragon {5}{R}{R}, 5/4 (Oath of the
// Gatewatch, #5545):
//
//	"Surge {3}{R}{R} (You may cast this spell for its surge cost if you or a
//	 teammate has cast another spell this turn.)
//	 Flying
//	 When this creature enters, if its surge cost was paid, it deals 3 damage
//	 to any target."
//
// An intervening "if" over the permanent's cast provenance (CR 603.4,
// CR 400.7d). The Tyrant is the damage source, as it last existed if it has
// left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "a4ddd9e2-ea59-4bba-bb2f-a655cabbb65c",
		Name:             "Tyrant of Valakut",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Surge("{3}{R}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.CastWithSurge),
					"Tyrant of Valakut — 3 damage to any target",
					effects.B33DamageChosenTargetFromSource(3)),
				effects.TargetAny()),
		},
	})
}

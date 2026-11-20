package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azra Smokeshaper — Creature — Azra Ninja {3}{B}, 3/3 (Modern Horizons, #5519):
//
//	"Ninjutsu {1}{B} ({1}{B}, Return an unblocked attacker you control to
//	 hand: Put this card onto the battlefield from your hand tapped and
//	 attacking.)
//	 When this creature enters, target creature you control gains
//	 indestructible until end of turn."
//
// The shared ninjutsu and an enters trigger on firstTargetGainsUntilEndOfTurn.
// It may target the Smokeshaper itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f6e0abb-18d0-4a15-9b21-0a72542998ac",
		Name:         "Azra Smokeshaper",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{1}{B}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Azra Smokeshaper — target creature you control gains indestructible until end of turn",
					firstTargetGainsUntilEndOfTurn(0, "Azra Smokeshaper — indestructible until end of turn", "indestructible")),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

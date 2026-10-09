package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yoshimaru, Scrappy Stray — Legendary Creature — Dog {1}{G}, 1/1:
//
//	"When Yoshimaru enters, another target creature you control fights
//	 up to one target creature an opponent controls.
//	 {6}: Put a +1/+1 counter on target nonlegendary creature."
//
// Two target clauses on the trigger (CR 601.2c): your creature, which
// must be another creature than Yoshimaru, and an optional one of an
// opponent's. With no opponent's creature chosen, nothing fights. Each
// slot is re-checked at resolution (CR 608.2b); the fight is the shared
// domriFight (slot 0 fights slot 1 via b10Fight), so damage is read after both powers are fixed and a fighter
// that left means no damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c87da757-3daa-4874-a1e4-be0aa2adcb28",
		Name:         "Yoshimaru, Scrappy Stray",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Yoshimaru, Scrappy Stray — another target creature you control fights up to one target creature an opponent controls",
				domriFight),
			Clauses(
				Another(TargetCreature("another target creature you control", YouControl())),
				TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1),
			),
		)},
		Activated: []ActivatedAbility{{
			Label:   "{6}: Put a +1/+1 counter on target nonlegendary creature.",
			Cost:    ManaCost("{6}"),
			Targets: TargetCreature("target nonlegendary creature", Not(Legendary())),
			Effect:  rfCreatureFCounterOnTargetCreature,
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hapatra, the Desert Frost — Legendary Creature — Human Wizard
// {3}{U}, 4/3:
//
//	"When Hapatra enters, for each opponent, tap up to one target
//	 creature that player controls. Put a stun counter on each of those
//	 creatures. (If a permanent with a stun counter would become
//	 untapped, remove one from it instead.)
//	 {2}{U}: Untap target creature."
//
// The ETB is the clause-per-opponent shape of Molten Primordial; each
// still-legal pick is tapped and gets a stun counter (a creature that
// was already tapped still gets the counter).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Hapatra, the Desert Frost — tap up to one creature each opponent controls and stun it", rfCreatureCHapatraFrostEffect)
	enters.TargetsFrom = moltenPrimordialClauses
	Register(Spec{
		OracleID:     "ac0f161a-e5ec-4a95-9f37-305bdf1ac900",
		Name:         "Hapatra, the Desert Frost",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Untap target creature.",
			Cost:    ManaCost("{2}{U}"),
			Targets: TargetCreature("target creature"),
			Effect:  untapTheTarget,
		}},
	})
}

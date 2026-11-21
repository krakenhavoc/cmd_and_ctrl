package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	enters := effects.WhenThisEnters("Hapatra, the Desert Frost — tap up to one creature each opponent controls and stun it", effects.RfCreatureCHapatraFrostEffect)
	enters.TargetsFrom = effects.MoltenPrimordialClauses
	effects.Register(effects.Spec{
		OracleID:     "ac0f161a-e5ec-4a95-9f37-305bdf1ac900",
		Name:         "Hapatra, the Desert Frost",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Untap target creature.",
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.UntapTheTarget,
		}},
	})
}

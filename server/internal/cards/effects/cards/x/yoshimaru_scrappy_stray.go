package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c87da757-3daa-4874-a1e4-be0aa2adcb28",
		Name:         "Yoshimaru, Scrappy Stray",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WhenThisEnters("Yoshimaru, Scrappy Stray — another target creature you control fights up to one target creature an opponent controls",
				effects.DomriFight),
			effects.Clauses(
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
				effects.TargetCreature("up to one target creature an opponent controls", effects.OpponentControls()).WithCount(0, 1),
			),
		)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}: Put a +1/+1 counter on target nonlegendary creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{6}"),
			Targets: effects.TargetCreature("target nonlegendary creature", effects.Not(effects.Legendary())),
			Effect:  effects.RfCreatureFCounterOnTargetCreature,
		}},
	})
}

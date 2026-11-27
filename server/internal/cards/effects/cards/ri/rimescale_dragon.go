package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rimescale Dragon — Snow Creature — Dragon {5}{R}{R}, 5/5 (Coldsnap, #5597):
//
//	"Flying
//	 {2}{S}: Tap target creature and put an ice counter on it.
//	 Creatures with ice counters on them don't untap during their
//	 controllers' untap steps."
//
// {S} is snow mana (Icebind Pillar). The static is Back to Basics' untap-step
// restriction over every creature, whoever controls it, that has an ice counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3f8182e5-4df5-4beb-a3b7-387ed1e86c99",
		Name:            "Rimescale Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{S}: Tap target creature and put an ice counter on it.",
			Cost:    effects.ManaCost("{2}{S}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB565TapAndIceTheTarget,
		}},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(
			func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.IsCreature() && c.Counters[effects.S82OldB565IceCounter] > 0
			})},
	})
}

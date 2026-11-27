package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kithkin Daggerdare — Creature — Kithkin Soldier {1}{G}, 1/1 (Lorwyn, S82
// OLD b545, #5593):
//
//	"{G}, {T}: Target attacking creature gets +2/+2 until end of turn."
//
// Bloodrush's target clause on an ordinary tap ability: a creature that
// stopped attacking, or left, by resolution is an illegal target and the
// ability does nothing (CR 608.2b). It needs the Daggerdare to be untapped
// and no longer summoning sick, like any {T} cost (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4fd5a3f-3aec-4336-9dcd-6971c117515a",
		Name:         "Kithkin Daggerdare",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: Target attacking creature gets +2/+2 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostTheTargetUntilEOT(2, 2, "Kithkin Daggerdare — +2/+2")(item, effects.NewContext(g, item))
			},
		}},
	})
}

package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Esper Battlemage — Artifact Creature — Human Wizard {2}{U}, 2/2 (Shards of
// Alara, S82 OLD batch 511, #5588):
//
//	"{W}, {T}: Prevent the next 2 damage that would be dealt to you this turn.
//	 {B}, {T}: Target creature gets -1/-1 until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2551f383-7308-4670-9fda-406b20e372a5",
		Name:         "Esper Battlemage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}, {T}: Prevent the next 2 damage that would be dealt to you this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.PreventNextDamage{Target: item.Controller, Amount: 2, Label: "Esper Battlemage: prevent the next 2 damage"}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{B}, {T}: Target creature gets -1/-1 until end of turn.",
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
				Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Esper Battlemage — -1/-1 until end of turn", -1, -1),
			},
		},
	})
}

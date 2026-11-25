package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// House Guildmage — Creature — Human Wizard {U}{B}, 2/2 (Guilds of Ravnica):
//
//	"{1}{U}, {T}: Target creature doesn't untap during its controller's next
//	 untap step.
//	 {2}{B}, {T}: Surveil 2."
//
// Two {T} abilities. The first is Dreamshackle Geist's second mode: it does
// not tap its target, only names its controller's next untap step
// (DoesntUntapNextUntapStep). The second is a plain surveil 2 for its
// controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "257cc8e9-55b4-4510-9d1f-5df8c1ae2061",
		Name:         "House Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{U}, {T}: Target creature doesn't untap during its controller's next untap step.",
				Purpose: game.Purpose{Answers: game.AnswerRestrict},
				Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.DoesntUntapNextUntapStep{Targets: effects.LegalTargetCardIDs(ctx), Label: "House Guildmage — doesn't untap during its controller's next untap step"}.Apply(ctx)
				},
			},
			{
				Label:   "{2}{B}, {T}: Surveil 2.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()),
				Effect:  effects.Do(effects.Surveil{N: 2}),
			},
		},
	})
}

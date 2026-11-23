package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nacatl Hunt-Pride — Creature — Cat Warrior {5}{W}, 5/4 (Conflux,
// #5587):
//
//	"Vigilance
//	 {R}, {T}: Target creature can't block this turn.
//	 {G}, {T}: Target creature blocks this turn if able."
//
// Jamuraan Lion's can't-block row and Peema Aether-Seer's block
// requirement (CR 509.1c), each behind a {T}, so only one a turn and not
// while the Hunt-Pride is summoning sick (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "094bc780-3c37-4863-b6ed-e2d16ed94544",
		Name:            "Nacatl Hunt-Pride",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}, {T}: Target creature can't block this turn.",
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
				Effect:  effects.TargetCreatureCantBlockThisTurn("Nacatl Hunt-Pride — target creature can't block this turn"),
			},
			{
				Label:   "{G}, {T}: Target creature blocks this turn if able.",
				Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BlockRequirementUntilEOT{
						Target: effects.FirstLegalBattlefieldTarget(ctx),
						Kind:   game.BlockRequirementBlocks,
						Label:  "Nacatl Hunt-Pride — blocks this turn if able",
					}.Apply(ctx)
				},
			},
		},
	})
}

package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thornscape Master — Creature — Human Wizard {2}{G}{G}, 2/2 (Invasion, S82 #5616):
//
//	"{R}{R}, {T}: This creature deals 2 damage to target creature.
//	 {W}{W}, {T}: Target creature gains protection from the color of your choice
//	 until end of turn."
//
// Prodigal Sorcerer's ping for 2 and Mother of Runes's protection, the colour
// named as the ability resolves (CR 602.2b): the target is chosen first, so an
// opponent responding has to commit before knowing the colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea802eda-fc8d-45c7-b5a3-238a2cdc4dcb",
		Name:         "Thornscape Master",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}{R}, {T}: This creature deals 2 damage to target creature.",
				Cost:    effects.Plus(effects.ManaCost("{R}{R}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
			},
			{
				Label:   "{W}{W}, {T}: Target creature gains protection from the color of your choice until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
				Cost:    effects.Plus(effects.ManaCost("{W}{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							effects.ChooseProtectionColorFor(ctx, item.Controller, t.ID, "Thornscape Master", "Thornscape Master — choose a color")
							return nil
						}
					}
					return nil
				},
			},
		},
	})
}

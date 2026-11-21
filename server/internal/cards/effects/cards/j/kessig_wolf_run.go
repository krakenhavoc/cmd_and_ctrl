package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kessig Wolf Run — Land:
//
//	"{T}: Add {C}.
//	 {X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until
//	 end of turn."
//
// X is announced with the activation (CR 107.3k) and read back with
// ctx.X(), the way Deepwood Elder's is. Both halves are until-end-of-turn
// effects on the one target, applied at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6911265-54ef-4c16-bcf2-1ffb24b7d426",
		Name:         "Kessig Wolf Run",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until end of turn",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.PowerXUntilEOT(0, "trample")), game.AnswerPump|game.AnswerCombatGrant),
			Cost:    effects.Plus(effects.ManaCost("{X}{R}{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, target := range ctx.LegalTargets() {
					if target.Kind != game.TargetCard {
						continue
					}
					if err := (effects.BoostUntilEOT{
						Target: target.ID,
						Power:  ctx.X(),
						Label:  "Kessig Wolf Run — +X/+0",
					}).Apply(ctx); err != nil {
						return err
					}
					return (effects.GrantKeywordUntilEOT{
						Target:   target.ID,
						Keywords: []string{"trample"},
						Label:    "Kessig Wolf Run — trample",
					}).Apply(ctx)
				}
				return nil
			},
		}},
	})
}

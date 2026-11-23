package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bane's Invoker — Creature — Human Cleric {1}{W}, 2/2 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Wind Walk — {8}: Up to two target creatures each get +2/+2 and gain flying
//	 until end of turn."
//
// Pseudodragon Familiar's keyword grant with a +2/+2 boost, on each of up to
// two targets still legal as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "flying"))
	purpose.Answers = game.AnswerPump
	effects.Register(effects.Spec{
		OracleID:     "daa47749-7a20-4e64-b476-a44eaeb4d844",
		Name:         "Bane's Invoker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Wind Walk — {8}: Up to two target creatures each get +2/+2 and gain flying until end of turn.",
			Purpose: purpose,
			Cost:    effects.ManaCost("{8}"),
			Targets: effects.TargetCreature("up to two target creatures").WithCount(0, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Bane's Invoker — +2/+2"}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Bane's Invoker — flying"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

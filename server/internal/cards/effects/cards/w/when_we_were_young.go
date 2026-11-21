package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// When We Were Young — Instant {3}{W} (Kamigawa: Neon Dynasty, #5493):
//
//	"Up to two target creatures each get +2/+2 until end of turn. If you
//	 control an artifact and an enchantment, those creatures also gain
//	 lifelink until end of turn."
//
// The condition is read as the spell resolves, after the pump, and the
// lifelink goes to the creatures that were still legal targets (CR 608.2b).
// The Huntsman's Redemption's pump.
//
// No simplification.
func init() {
	const name = "When We Were Young"
	effects.Register(effects.Spec{
		OracleID:     "1f27aa69-d17e-4f1f-9aba-cee205036f93",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			lifelink := effects.NeoB08YouControlAnArtifactAndAnEnchantment(ctx.Game, ctx.Controller())
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: name + " — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				if !lifelink {
					continue
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"lifelink"}, Label: name + " — lifelink"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

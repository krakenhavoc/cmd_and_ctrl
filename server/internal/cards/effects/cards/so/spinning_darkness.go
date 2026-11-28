package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spinning Darkness — {4}{B}{B} Instant (Weatherlight, S82 OLD batch 771, #5629):
//
//	"You may exile the top three black cards of your graveyard rather than pay this
//	 spell's mana cost.
//	 Spinning Darkness deals 3 damage to target nonblack creature. You gain 3
//	 life."
//
// The alternative cost is a free one (CR 118.9) paid with the three topmost black
// cards of the caster's graveyard (S82OldB771ExileTopThreeBlack; the cards are
// the pile's, not the caster's to pick, CR 404.2), and unavailable with fewer than
// three. If the target is illegal as the spell resolves, none of it happens (CR
// 608.2b), the life included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "69eeee95-cc69-41d1-833b-8895ef72df82",
		Name:             "Spinning Darkness",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.S82OldB771ExileTopThreeBlack()},
		Targets:          effects.TargetCreature("target nonblack creature", effects.Not(effects.OfColor("B"))),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return spinningDarknessResolve(ctx, t.ID)
			}
			return nil
		},
	})
}

// spinningDarknessResolve deals the 3 damage to the creature, then gains the
// caster 3 life.
func spinningDarknessResolve(ctx *effects.Context, target uuid.UUID) error {
	if err := (effects.DealDamage{Source: ctx.Source(), Target: target, Amount: 3}).Apply(ctx); err != nil {
		return err
	}
	return effects.GainLife{Amount: 3}.Apply(ctx)
}

package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surge of Strength — Instant {R}{G} (Alliances, S82 OLD batch 817):
//
//	"As an additional cost to cast this spell, discard a red or green card.
//	 Target creature gains trample and gets +X/+0 until end of turn, where
//	 X is that creature's mana value."
//
// X is read once, as the spell resolves (CR 608.2h, CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:       "bd3e55cf-ffca-402c-91ea-2420035a9dd1",
		Name:           "Surge of Strength",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.DiscardColorCost("a red or green card", "R", "G"),
		Purpose:        game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(effects.PowerXUntilEOT(0, "trample"))},
		Targets:        effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Surge of Strength — trample until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return effects.BoostUntilEOT{Target: id, Power: c.ManaValue(), Label: "Surge of Strength — +X/+0 until end of turn"}.Apply(ctx)
		},
	})
}

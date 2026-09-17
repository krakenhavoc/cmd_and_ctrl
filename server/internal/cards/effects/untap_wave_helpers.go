package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func b751NonbasicLand() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLand() && !IsBasicLand(c) }
}

func b751Tapped() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Tapped }
}

func b751CreatureIDs(g *game.Game, player uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && (player == uuid.Nil || c.Controller == player) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

func b751TapFreezeTargets(ctx *Context, player uuid.UUID, label string) error {
	ids := make([]uuid.UUID, 0, len(ctx.LegalTargets()))
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	return (TapAndFreeze{Targets: ids, Player: player, Label: label}).Apply(ctx)
}

func b751FreezeTarget(ctx *Context, player uuid.UUID) error {
	if len(ctx.LegalTargets()) == 0 {
		return nil
	}
	return (DoesntUntapNextUntapStep{Targets: []uuid.UUID{ctx.LegalTargets()[0].ID}, Player: player}).Apply(ctx)
}

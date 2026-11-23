package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mists of Lórien — Sorcery {2}{U} (Tales of Middle-earth Commander):
//
//	"Replicate {U} (When you cast this spell, copy it for each time you
//	 paid its replicate cost. You may choose new targets for the
//	 copies.)
//	 Return target nonland permanent and each other nonland permanent
//	 with the same mana value as that permanent to their owners' hands."
//
// Replicate is Gigadrowse's pair (replicateMana, Replicate(), CR
// 702.56a), capped at twenty payments, far beyond any mana a table
// holds. The target's mana value is read as the spell resolves, before
// anything moves (CR 608.2h), and every nonland permanent on the
// battlefield with that value, the target included, returns together
// (a token or a face-down permanent has mana value 0). A target gone
// as the spell resolves makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "65259ba4-956d-46a3-abd9-791d02015b41",
		Name:          "Mists of Lórien",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OptionalCosts: []game.AdditionalCost{effects.ReplicateMana("{U}", 20)},
		Triggered:     []game.TriggeredAbility{effects.Replicate()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(t.ID)
			if !found {
				return nil
			}
			mv, known := ctx.Game.ManaValueForEffect(c)
			if !known {
				return nil
			}
			var ids []uuid.UUID
			for _, p := range effects.MatchingBattlefield(ctx, effects.And(effects.Nonland(), func(g *game.Game, _ uuid.UUID, o game.Card) bool {
				v, ok := g.ManaValueForEffect(o)
				return ok && v == mv
			})) {
				ids = append(ids, p.InstanceID)
			}
			return effects.ReturnAllToHand{Cards: ids}.Apply(ctx)
		},
	})
}

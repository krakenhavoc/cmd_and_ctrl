package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra's Flame Wave — Sorcery {3}{R}{R} (Core Set 2020, #5518):
//
//	"Chandra's Flame Wave deals 2 damage to target player and each creature
//	 that player controls. Search your library and/or graveyard for a card
//	 named Chandra, Flame's Fury, reveal it, and put it into your hand. If you
//	 search your library this way, shuffle."
//
// Chandra, Flame's Fury's −8 shape with 2 damage: the player and the
// creatures that player controls as it resolves take it as one instruction,
// then the search runs over both zones (the zones are asked first, so a
// graveyard-only search does not shuffle). With no legal target the spell does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f84fc91-8647-4cc8-9b57-d23a83f85ab1",
		Name:         "Chandra's Flame Wave",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok {
				return nil
			}
			ids := []uuid.UUID{t.ID}
			for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.ControlledBy(t.ID))) {
				ids = append(ids, c.InstanceID)
			}
			if err := effects.DealDamageToEachThen(ctx, ids, 2, effects.RecipientThen{}); err != nil {
				return err
			}
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     func(c game.Card) bool { return c.Name == "Chandra, Flame's Fury" },
				Dest:          game.ZoneHand,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				AlsoGraveyard: true,
				Reason:        "Chandra's Flame Wave — a card named Chandra, Flame's Fury",
			}.Apply(ctx)
		},
	})
}

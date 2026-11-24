package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volcanic Vision — Sorcery {5}{R}{R} (Commander 2021, #5504):
//
//	"Return target instant or sorcery card from your graveyard to your
//	 hand. Volcanic Vision deals damage equal to that card's mana value to
//	 each creature your opponents control. Exile Volcanic Vision."
//
// The damage is the card's mana value as it sat in the graveyard (CR
// 608.2h, an X is 0), dealt as one event by the spell to each creature the
// caster's opponents control. A target gone by resolution makes the spell
// do nothing (CR 608.2b). The spell is exiled last, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe3b4c74-2bf8-464a-8f55-9de401a8c4b9",
		Name:         "Volcanic Vision",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, OpponentsOnly: true}},
		Targets:      effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			mv := card.ManaValue()
			if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
			if err := effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.OpponentControls()), mv); err != nil {
				return err
			}
			return effects.ExileTarget{Target: ctx.Source()}.Apply(ctx)
		},
	})
}

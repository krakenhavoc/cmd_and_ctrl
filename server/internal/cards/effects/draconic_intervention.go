package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Draconic Intervention — Sorcery {2}{R}{R} (#3537, ADR 0189 slice 3):
//
//	"As an additional cost to cast this spell, exile an instant or sorcery
//	 card from your graveyard.
//	 Draconic Intervention deals X damage to each non-Dragon creature,
//	 where X is the exiled card's mana value. If a creature dealt damage
//	 this way would die this turn, exile it instead.
//	 Exile Draconic Intervention."
//
// The cost is ExileFromGraveyardCost (#3536) with an instant-or-sorcery
// match, so the caster names the card on cast_spell as exile_ids, it is
// exiled with the spell on the stack, and ctx.Exiled() hands it back at
// resolution. X is that card's mana value, read from the exile zone where
// the cost left it (CR 202.3). The damage is the Anger of the Gods shape
// (damageEachExileIfDealtDies): each creature is told what it was dealt, so
// one that took none is not marked. The spell exiles itself last, so a
// countered one goes to the graveyard like any other (CR 608.2n is moot: the
// instruction is part of the resolution).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fe75f56-c2e4-4c82-8113-039b8c486fae",
		Name:         "Draconic Intervention",
		Completeness: CompletenessFull,
		AdditionalCost: ExileFromGraveyardCost(1, "an instant or sorcery card", func(c game.Card) bool {
			return c.IsInstant() || c.IsSorcery()
		}),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := 0
			for _, id := range ctx.Exiled() {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok {
					x = c.ManaValue()
				}
			}
			if err := damageEachExileIfDealtDies(ctx, And(Creature(), Not(OfCreatureType("Dragon"))), x, false); err != nil {
				return err
			}
			if item.IsCopy {
				return nil
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}

package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tundra Fumarole — Snow Sorcery {1}{R}{R} (Kaldheim):
//
//	"Tundra Fumarole deals 4 damage to target creature or planeswalker. Add
//	 {C} for each {S} spent to cast this spell. Until end of turn, you don't
//	 lose this mana as steps and phases end. ({S} is mana from a snow source.)"
//
// The mana is counted from the snow mana that paid for the spell (CR 107.4h)
// and carries the keep mark (KeepManaUntilEndOfTurn), as Birgi's does. With a
// target gone the whole spell is countered on resolution (CR 608.2b), so no
// mana is added.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "52befc89-e895-4dd5-95e4-8f6d03e627aa",
		Name:           "Tundra Fumarole",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsPoolManaOnly,
		WantsManaFrom:  game.ManaSourceSnow,
		Targets:        effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 4}).Apply(ctx); err != nil {
				return err
			}
			n := ctx.ManaSpent().CountFrom(game.ManaSourceSnow)
			if n <= 0 {
				return nil
			}
			return effects.AddMana{
				Amount: effects.ManaN(n, "{C}"),
				Riders: []game.ManaSpendRider{effects.KeepManaUntilEndOfTurn()},
			}.Apply(ctx)
		},
	})
}

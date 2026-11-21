package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Price of Loyalty — Sorcery {2}{R} (Adventures in the Forgotten Realms):
//
//	"Gain control of target creature until end of turn. Untap that creature.
//	 It gains haste until end of turn. If mana from a Treasure was spent to
//	 cast this spell, that creature gets +2/+0 until end of turn."
//
// Act of Treason's three clauses, then the Treasure rider read from the spell's
// payment record as it resolves (Devour Intellect's reading).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "f25b97f9-8ca2-40f7-9d55-47fa8bb2054e",
		Name:           "Price of Loyalty",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   effects.CompletenessFull,
		Targets:        effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.GainControl{
				Target:   target,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Price of Loyalty — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"haste"},
				Label: "Price of Loyalty — haste until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			if ctx.ManaSpent().FromTreasure() {
				return effects.BoostUntilEOT{Target: target, Power: 2, Label: "Price of Loyalty — +2/+0 until end of turn"}.Apply(ctx)
			}
			return nil
		},
	})
}

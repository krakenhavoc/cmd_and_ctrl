package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undying Flames — Sorcery {4}{R}{R} (Saviors of Kamigawa, #5601):
//
//	"Exile cards from the top of your library until you exile a nonland
//	 card. Undying Flames deals damage to any target equal to that card's
//	 mana value.
//	 Epic (For the rest of the game, you can't cast spells. At the
//	 beginning of each of your upkeeps, copy this spell except for its epic
//	 ability. You may choose a new target for the copy.)"
//
// Breaching Dragonstorm's exile-until-nonland run (UntilNonland), then the
// damage: the nonland card is the last that landed in exile, and its mana
// value is read there (CR 202.3). A library with no nonland card exiles all
// of it and deals no damage. The target is re-checked as the damage is
// dealt (CR 608.2b). Epic is effects.Epic (CR 702.50a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5db128ed-69c0-41f3-b8b4-fb7cefdb236c",
		Name:         "Undying Flames",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OnResolve: effects.Epic(func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.MillToZoneKeyedForEffect(ctx.Controller(), 0, game.ZoneExile, effects.UntilNonland, ctx.ItemTail(undyingFlamesDamage))
		}),
	})
}

// undyingFlamesDamage deals the damage once the run has ended. An on-disk key.
var undyingFlamesDamage = effects.CardTail("undying-flames/damage", func(ctx *effects.Context, o game.TailOutcome) error {
	for i := len(o.IDs) - 1; i >= 0; i-- {
		c, ok := ctx.Game.LookupCardForEffect(o.IDs[i])
		if !ok || c.IsLand() {
			continue
		}
		t, ok := effects.FirstLegalTarget(ctx)
		if !ok {
			return nil
		}
		mv, _ := ctx.Game.ManaValueForEffect(c)
		return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: mv}.Apply(ctx)
	}
	return nil
})

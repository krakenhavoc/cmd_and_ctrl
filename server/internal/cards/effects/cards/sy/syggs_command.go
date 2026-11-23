package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sygg's Command — Kindred Sorcery — Merfolk {1}{W}{U} (Lorwyn Eclipsed, #4458):
//
//	"Choose two —
//	 • Create a token that's a copy of target Merfolk you control.
//	 • Creatures target player controls gain lifelink until end of turn.
//	 • Target player draws a card.
//	 • Tap target creature. Put a stun counter on it."
//
// Each bullet carries its own target group, and with at least one still
// legal the spell does what it can (the ruling). The Merfolk is any permanent
// you control with the type, a Kindred Merfolk included; X in its cost is 0.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fe5baab-b182-4cd9-bea5-912ce9b1464a",
		Name:         "Sygg's Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeDoing("Create a token that's a copy of target Merfolk you control.",
				effects.TargetPermanent("target Merfolk you control", effects.YouControl(), effects.HasSubtype("Merfolk")),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.CreateTokenCopy{Copy: t.ID, N: 1}.Apply(ctx)
				}),
			effects.ModeDoing("Creatures target player controls gain lifelink until end of turn.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.ControlledBy(t.ID)),
						Keywords: []string{"lifelink"},
						Label:    "Sygg's Command — lifelink until end of turn",
					}.Apply(ctx)
				}),
			effects.ModeDoing("Target player draws a card.",
				effects.TargetPlayer("target player"),
				effects.ModeTargetPlayerDraws),
			effects.ModeDoing("Tap target creature. Put a stun counter on it.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return effects.AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}.Apply(ctx)
				}),
		),
	})
}

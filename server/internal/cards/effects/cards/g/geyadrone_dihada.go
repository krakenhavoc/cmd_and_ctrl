package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Geyadrone Dihada — {1}{U}{B}{R} Legendary Planeswalker — Dihada,
// loyalty 4 (Modern Horizons 2, S82 OLD b039, #6149):
//
//	"Protection from permanents with corruption counters on them
//	 +1: Each opponent loses 2 life and you gain 2 life. Put a corruption
//	 counter on up to one other target creature or planeswalker.
//	 −3: Gain control of target creature or planeswalker until end of
//	 turn. Untap it and put a corruption counter on it. It gains haste
//	 until end of turn.
//	 −7: Gain control of each permanent with a corruption counter on it."
//
// The protection is #6149's counter quality (CR 702.16a): a permanent
// with a corruption counter on it can't target, damage, block or enchant
// Geyadrone, read off the source's counters as the check is made, so a
// permanent whose counter is removed is no longer stopped. The −7 takes
// each such permanent for as long as the effect lasts (no stated
// duration, CR 611.2a).
//
// No simplification.
func init() {
	corrupted := func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return effects.HasCounterOfKind(c, geyadroneCorruption)
	}
	effects.Register(effects.Spec{
		OracleID:        "e1105d84-86b5-44ef-9efc-d994ee1c35cc",
		Name:            "Geyadrone Dihada",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		PrintedKeywords: []string{"protection from permanents with corruption counters on them"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Each opponent loses 2 life and you gain 2 life. Put a corruption counter on up to one other target creature or planeswalker.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.Another(effects.TargetPermanent("up to one other target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())).WithCount(0, 1)),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if err := effects.EachOpponentLosesTwoYouGainTwo(ctx.Item, ctx, 0); err != nil {
						return err
					}
					if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
						return effects.AddCounter{Target: id, Kind: geyadroneCorruption, N: 1}.Apply(ctx)
					}
					return nil
				})),
			},
			{
				Label:   "−3: Gain control of target creature or planeswalker until end of turn. Untap it and put a corruption counter on it. It gains haste until end of turn.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.GainControl{
						Target:   id,
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Geyadrone Dihada — gain control until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.AddCounter{Target: id, Kind: geyadroneCorruption, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{
						Target:   id,
						Keywords: []string{"haste"},
						Label:    "Geyadrone Dihada — haste until end of turn",
					}.Apply(ctx)
				})),
			},
			{
				Label: "−7: Gain control of each permanent with a corruption counter on it.",
				Cost:  effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					for _, c := range effects.MatchingBattlefield(ctx, corrupted) {
						if err := (effects.GainControl{
							Target:   c.InstanceID,
							Duration: game.IndefiniteDuration(),
							Label:    "Geyadrone Dihada — gain control (no stated duration)",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})),
			},
		},
	})
}

// geyadroneCorruption is the corruption counter.
const geyadroneCorruption = "corruption"

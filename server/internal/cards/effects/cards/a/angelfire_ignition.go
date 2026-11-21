package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelfire Ignition — Sorcery {1}{R}{W} (Midnight Hunt, #5497):
//
//	"Put two +1/+1 counters on target creature. It gains vigilance,
//	 trample, lifelink, indestructible, and haste until end of turn.
//	 Flashback {2}{R}{W} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Two counters, then the five keywords until end of turn; Flashback is the
// shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "29e1c6e3-d76e-4388-b4e9-f732fdec4338",
		Name:             "Angelfire Ignition",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{2}{R}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"vigilance", "trample", "lifelink", "indestructible", "haste"},
				Label:    "Angelfire Ignition — vigilance, trample, lifelink, indestructible and haste",
			}.Apply(ctx)
		},
	})
}

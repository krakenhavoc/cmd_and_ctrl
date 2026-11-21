package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Homestead Courage — Sorcery {W} (Midnight Hunt, #5497):
//
//	"Put a +1/+1 counter on target creature you control. It gains
//	 vigilance until end of turn.
//	 Flashback {W} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The counter, then the keyword; both read the one target, and a target
// that left does nothing (CR 608.2b). Flashback is the shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "203eb24c-f135-44e2-bdae-80b7b3d0fb15",
		Name:             "Homestead Courage",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature you control", effects.YouControl()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{W}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"}, Label: "Homestead Courage — vigilance"}.Apply(ctx)
		},
	})
}

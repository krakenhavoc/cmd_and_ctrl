package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mark of Mutiny — Sorcery {2}{R} (Magic 2013):
//
//	"Gain control of target creature until end of turn. Put a +1/+1 counter on
//	 it and untap it. That creature gains haste until end of turn."
//
// Threaten's control effect (layer 2, ending at cleanup) and Song-Mad
// Treachery's untap and haste, with the +1/+1 counter put on in between, in
// printed order. The counter stays after the control effect ends.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5599ddb6-1404-4884-b6ca-c1852a521079",
		Name:         "Mark of Mutiny",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GainControl{
				Target:   id,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Mark of Mutiny — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"haste"},
				Label:    "Mark of Mutiny — haste until end of turn",
			}.Apply(ctx)
		},
	})
}

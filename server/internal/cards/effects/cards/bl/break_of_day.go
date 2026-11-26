package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Break of Day — Instant {1}{W}:
//
//	"Creatures you control get +1/+1 until end of turn.
//	 Fateful hour — If you have 5 or less life, those creatures gain
//	 indestructible until end of turn."
//
// "Those creatures" are the ones the pump found as it resolved (CR 611.2c),
// so both halves read one set; the life total is read after the pump, as
// printed.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59674589-b174-4150-b8d6-30d25c5ead4c",
		Name:         "Break of Day",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine := effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.Creature())
			if err := (effects.BoostUntilEOT{
				Match:     effects.And(effects.Creature(), effects.YouControl()),
				Power:     1,
				Toughness: 1,
				Label:     "Break of Day — +1/+1",
			}).Apply(ctx); err != nil {
				return err
			}
			if !effects.OldB441FatefulHour(ctx.Game, ctx.Controller()) {
				return nil
			}
			for _, id := range mine {
				if err := (effects.GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"indestructible"},
					Label:    "Break of Day — fateful hour: indestructible",
				}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

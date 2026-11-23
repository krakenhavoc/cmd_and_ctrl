package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overwhelming Victory — Instant — Lesson {4}{R} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Overwhelming Victory deals 5 damage to target creature. Each creature
//	 you control gains trample and gets +X/+0 until end of turn, where X
//	 is the amount of excess damage dealt this way."
//
// The excess is CR 120.4a's: the damage that landed beyond what the
// creature still needed to die, damage already marked included, and with
// a deathtouch source only the first point is needed
// (socB20DealDamageWithExcess). A prevented blow leaves less excess. The
// creatures you control are read once after the damage (CR 611.2c); they
// gain trample even when there is no excess.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20d94561-7526-4305-8d2d-7daf62729ff4",
		Name:         "Overwhelming Victory",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			excess, err := effects.SocB20DealDamageWithExcess(ctx, id, 5)
			if err != nil {
				return err
			}
			yours := effects.And(effects.Creature(), effects.YouControl())
			if err := (effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"trample"}, Label: "Overwhelming Victory — trample"}).Apply(ctx); err != nil {
				return err
			}
			return effects.BoostUntilEOT{Match: yours, Power: excess, Label: "Overwhelming Victory — +X/+0"}.Apply(ctx)
		},
	})
}
